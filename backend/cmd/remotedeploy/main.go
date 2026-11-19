package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type RemoteExecutor struct {
	client *ssh.Client
	pass   string
}

func dialWithProxyFallback(targetHost string, timeout time.Duration) (net.Conn, error) {
	// First try direct
	conn, err := net.DialTimeout("tcp", targetHost, 3*time.Second)
	if err == nil {
		return conn, nil
	}

	// Try via local office proxy (10.15.3.20:80)
	proxyAddr := "10.15.3.20:80"
	pConn, pErr := net.DialTimeout("tcp", proxyAddr, timeout)
	if pErr != nil {
		return nil, fmt.Errorf("direct connect failed (%v) and proxy connect failed: %w", err, pErr)
	}

	req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\nProxy-Connection: Keep-Alive\r\n\r\n", targetHost, targetHost)
	if _, err := pConn.Write([]byte(req)); err != nil {
		pConn.Close()
		return nil, fmt.Errorf("proxy write failed: %w", err)
	}

	br := bufio.NewReader(pConn)
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		pConn.Close()
		return nil, fmt.Errorf("failed reading proxy response: %w", err)
	}
	if resp.StatusCode != 200 {
		pConn.Close()
		return nil, fmt.Errorf("proxy CONNECT rejected with status: %s", resp.Status)
	}
	return pConn, nil
}

func NewRemoteExecutor(host, user, pass string) (*RemoteExecutor, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         20 * time.Second,
	}

	conn, err := dialWithProxyFallback(host, 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to reach host %s: %w", host, err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, host, config)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("SSH handshake failed: %w", err)
	}

	client := ssh.NewClient(sshConn, chans, reqs)
	return &RemoteExecutor{client: client, pass: pass}, nil
}

func (r *RemoteExecutor) Close() {
	if r.client != nil {
		r.client.Close()
	}
}

func (r *RemoteExecutor) Run(cmd string) (string, error) {
	session, err := r.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(cmd)
	out := stdout.String()
	if stderr.Len() > 0 {
		out += "\nSTDERR:\n" + stderr.String()
	}
	return strings.TrimSpace(out), err
}

func (r *RemoteExecutor) Sudo(script string) (string, error) {
	session, err := r.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return "", err
	}

	var stdout bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stdout

	sudoCmd := "sudo -S -p 'PIDIE_SUDO_PROMPT:' bash -s"
	if err := session.Start(sudoCmd); err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		fmt.Fprintf(stdin, "%s\n", r.pass)
		io.WriteString(stdin, script+"\n")
	}()

	err = session.Wait()
	res := stdout.String()
	res = strings.ReplaceAll(res, "PIDIE_SUDO_PROMPT:", "")
	return strings.TrimSpace(res), err
}

func (r *RemoteExecutor) UploadTarGz(tarGzData []byte, remoteDir string) error {
	b64 := base64.StdEncoding.EncodeToString(tarGzData)
	script := fmt.Sprintf(`
mkdir -p %s
cat << 'EOF_B64_DATA' | base64 -d | tar -xzf - -C %s
%s
EOF_B64_DATA
`, remoteDir, remoteDir, b64)

	_, err := r.Sudo(script)
	return err
}

func createTarGz(srcDir string) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		headerName := filepath.ToSlash(relPath)

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = headerName

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func main() {
	host := "43.157.247.26:22"
	user := "ubuntu"
	pass := "shadow-47$-thunder"

	exec, err := NewRemoteExecutor(host, user, pass)
	if err != nil {
		log.Fatalf("SSH connection error: %v", err)
	}
	defer exec.Close()

	fmt.Println("=== 1. Pulling latest commit from GitHub ===")
	gitPullCmd := `
cd /home/ubuntu/pidiekab
git fetch origin main
git reset --hard origin/main
git log -1 --oneline
`
	out, err := exec.Sudo(gitPullCmd)
	if err != nil {
		log.Fatalf("Git pull failed: %v\nOutput: %s", err, out)
	}
	fmt.Printf("Git state:\n%s\n", out)

	fmt.Println("\n=== 2. Compiling Go binary on server ===")
	compileCmd := `
export PATH="/snap/bin:/usr/local/bin:$PATH"
cd /home/ubuntu/pidiekab/backend
CGO_ENABLED=0 go build -ldflags="-s -w" -o /opt/pidiekab/current/bin/pidiekab ./cmd/pidiekab
chown pidiekab:pidiekab /opt/pidiekab/current/bin/pidiekab
chmod +x /opt/pidiekab/current/bin/pidiekab
`
	out, err = exec.Sudo(compileCmd)
	if err != nil {
		log.Fatalf("Compile failed: %v\nOutput: %s", err, out)
	}
	fmt.Println("Binary compiled successfully from pulled commit.")

	fmt.Println("\n=== 3. Packaging & Syncing CMS Dist to Server ===")
	cmsDistPath := "../frontend/admin-cms/dist"
	if _, err := os.Stat(cmsDistPath); os.IsNotExist(err) {
		cmsDistPath = "frontend/admin-cms/dist"
	}
	tarGz, err := createTarGz(cmsDistPath)
	if err != nil {
		log.Fatalf("Failed to package CMS dist: %v", err)
	}
	fmt.Printf("CMS dist tarball size: %d bytes\n", len(tarGz))

	// Sync to /opt/pidiekab/current/frontend/admin-cms/dist
	targetDir := "/opt/pidiekab/current/frontend/admin-cms/dist"
	if err := exec.UploadTarGz(tarGz, targetDir); err != nil {
		log.Fatalf("Failed to upload CMS dist: %v", err)
	}
	// Also mirror to /home/ubuntu/pidiekab/frontend/admin-cms/dist
	exec.UploadTarGz(tarGz, "/home/ubuntu/pidiekab/frontend/admin-cms/dist")

	// Set ownership and permissions
	exec.Sudo("chown -R pidiekab:pidiekab /opt/pidiekab/current/frontend/admin-cms/dist")
	fmt.Println("CMS dist successfully synced to production locations.")

	fmt.Println("\n=== 4. Executing SSG Build on Server ===")
	ssgBuildCmd := `
cd /opt/pidiekab/current
su -s /bin/bash pidiekab -c 'PIDIEKAB_DATABASE_URL="postgres://pidiekab:pidiekab_strong_pass_2026!@127.0.0.1:5432/pidiekab_production?sslmode=disable" PIDIEKAB_SSG_OUTPUT_ROOT="/var/lib/pidiekab/ssg" PIDIEKAB_BASE_URL="https://pidiekab.web.id" /opt/pidiekab/current/bin/pidiekab ssg build'
`
	out, err = exec.Sudo(ssgBuildCmd)
	fmt.Printf("SSG build output:\n%s\n", out)

	fmt.Println("\n=== 5. Restarting pidiekab.service ===")
	exec.Sudo("systemctl restart pidiekab.service")
	time.Sleep(2 * time.Second)

	fmt.Println("\n=== 6. Verifying CMS index.html in Production ===")
	snippet, _ := exec.Run("curl -s -k https://127.0.0.1/cms/ | grep -i 'interactive-widget'")
	fmt.Printf("CMS Viewport check:\n%s\n", snippet)
	loginSnippet, _ := exec.Run("curl -s -k https://127.0.0.1/cms/ | head -n 15")
	fmt.Printf("CMS HTML head:\n%s\n", loginSnippet)
}

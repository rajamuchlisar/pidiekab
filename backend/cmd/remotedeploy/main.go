package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type RemoteExecutor struct {
	client *ssh.Client
	pass   string
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

	conn, err := net.DialTimeout("tcp", host, 20*time.Second)
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

func main() {
	host := "43.157.247.26:22"
	user := "ubuntu"
	pass := "shadow-47$-thunder"

	exec, err := NewRemoteExecutor(host, user, pass)
	if err != nil {
		log.Fatalf("SSH connection error: %v", err)
	}
	defer exec.Close()

	fmt.Println("=== 1. Updating /etc/pidiekab/pidiekab.env to use https://pidiekab.web.id ===")
	envFileContent := `# PidieKab Production Environment
PIDIEKAB_ENV=production
PIDIEKAB_ADDR=:80
PIDIEKAB_LOG_LEVEL=info

PIDIEKAB_READ_HEADER_TIMEOUT=5s
PIDIEKAB_READ_TIMEOUT=15s
PIDIEKAB_WRITE_TIMEOUT=30s
PIDIEKAB_IDLE_TIMEOUT=60s
PIDIEKAB_SHUTDOWN_TIMEOUT=15s

PIDIEKAB_DATABASE_URL=postgres://pidiekab:pidiekab_strong_pass_2026!@127.0.0.1:5432/pidiekab_production?sslmode=disable
PIDIEKAB_DATABASE_MAX_CONNS=25
PIDIEKAB_DATABASE_MIN_CONNS=5
PIDIEKAB_DATABASE_MAX_CONN_LIFETIME=1h
PIDIEKAB_DATABASE_MAX_CONN_IDLE_TIME=30m
PIDIEKAB_DATABASE_HEALTH_CHECK_PERIOD=1m

PIDIEKAB_PUBLIC_STATIC_DIR=/opt/pidiekab/current/frontend/public-site/dist
PIDIEKAB_CMS_STATIC_DIR=/opt/pidiekab/current/frontend/admin-cms/dist
PIDIEKAB_STORAGE_PATH=/var/lib/pidiekab/media
PIDIEKAB_SSG_OUTPUT_ROOT=/var/lib/pidiekab/ssg
PIDIEKAB_SSG_RELEASES_TO_KEEP=5
PIDIEKAB_SSG_POLL_INTERVAL=5s
PIDIEKAB_SSG_JOB_MAX_ATTEMPTS=3

PIDIEKAB_BASE_URL=https://pidiekab.web.id
`
	writeEnvCmd := fmt.Sprintf(`cat << 'EOF' > /etc/pidiekab/pidiekab.env
%s
EOF
chown root:pidiekab /etc/pidiekab/pidiekab.env
chmod 0640 /etc/pidiekab/pidiekab.env
`, envFileContent)

	if _, err := exec.Sudo(writeEnvCmd); err != nil {
		log.Fatalf("Failed to write env file: %v", err)
	}
	fmt.Println("Environment updated with PIDIEKAB_BASE_URL=https://pidiekab.web.id.")

	fmt.Println("\n=== 2. Re-building SSG with https://pidiekab.web.id ===")
	ssgBuildCmd := `
cd /opt/pidiekab/current
su -s /bin/bash pidiekab -c 'PIDIEKAB_DATABASE_URL="postgres://pidiekab:pidiekab_strong_pass_2026!@127.0.0.1:5432/pidiekab_production?sslmode=disable" PIDIEKAB_SSG_OUTPUT_ROOT="/var/lib/pidiekab/ssg" PIDIEKAB_BASE_URL="https://pidiekab.web.id" /opt/pidiekab/current/bin/pidiekab ssg build'
`
	out, err := exec.Sudo(ssgBuildCmd)
	fmt.Printf("SSG build output:\n%s\n", out)

	fmt.Println("\n=== 3. Restarting pidiekab.service ===")
	exec.Sudo("systemctl restart pidiekab.service")
	time.Sleep(2 * time.Second)

	fmt.Println("\n=== 4. Verifying Sitemap and Canonical URL ===")
	sitemapOut, _ := exec.Run("curl -s http://127.0.0.1/sitemap.xml")
	fmt.Printf("sitemap.xml (first 400 chars):\n%s\n\n", sitemapOut[:min(400, len(sitemapOut))])

	robotsOut, _ := exec.Run("curl -s http://127.0.0.1/robots.txt")
	fmt.Printf("robots.txt:\n%s\n\n", robotsOut)

	homeCanonical, _ := exec.Run("curl -s http://127.0.0.1/id/ | grep -E 'canonical|alternate'")
	fmt.Printf("HTML Canonical & Alternate tags:\n%s\n", homeCanonical)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

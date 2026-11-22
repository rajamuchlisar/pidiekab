package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type RemoteExecutor struct {
	client *ssh.Client
	pass   string
}

func dialWithProxyFallback(targetHost string, timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", targetHost, 3*time.Second)
	if err == nil {
		return conn, nil
	}

	proxyAddr := "10.15.3.20:80"
	pConn, pErr := net.DialTimeout("tcp", proxyAddr, timeout)
	if pErr != nil {
		return nil, fmt.Errorf("direct failed (%v) and proxy failed: %w", err, pErr)
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
		return nil, fmt.Errorf("proxy read failed: %w", err)
	}
	if resp.StatusCode != 200 {
		pConn.Close()
		return nil, fmt.Errorf("proxy rejected status: %s", resp.Status)
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

	fmt.Println("==================================================")
	fmt.Println("   PIDIEKAB SERVER OPTIMIZATION & CLEANUP TOOL    ")
	fmt.Println("==================================================")

	// Step 1: Initial Diagnosis
	fmt.Println("\n[1/5] Memeriksa Status Awal Server (RAM, Swap, Disk)...")
	diagScript := `
echo "=== Cloudflared Config ==="
cat /etc/cloudflared/config.yml 2>/dev/null
`
	out, err := exec.Sudo(diagScript)
	if err != nil {
		log.Printf("Diag warning: %v", err)
	}
	fmt.Println(out)

	// Step 2: Swap Configuration
	fmt.Println("\n[2/5] Memeriksa & Mengoptimalkan Swap File (2GB)...")
	swapScript := `
if swapon --show | grep -q "/swapfile"; then
    echo " [+] Swapfile /swapfile sudah aktif:"
    swapon --show
else
    echo " [*] Swapfile belum aktif. Membuat /swapfile sebesar 2GB..."
    # Matikan swapfile lama jika rusak
    swapoff /swapfile 2>/dev/null || true
    
    # Gunakan fallocate atau dd
    fallocate -l 2G /swapfile || dd if=/dev/zero of=/swapfile bs=1M count=2048 status=progress
    chmod 600 /swapfile
    mkswap /swapfile
    swapon /swapfile
    
    # Daftarkan ke /etc/fstab jika belum ada
    if ! grep -q '/swapfile' /etc/fstab; then
        echo '/swapfile none swap sw 0 0' >> /etc/fstab
        echo " [+] Didaftarkan ke /etc/fstab permanen."
    fi
    echo " [+] Swapfile 2GB berhasil diaktifkan!"
fi

# Optimasi kernel swappiness & cache pressure
echo " [*] Mengonfigurasi vm.swappiness=10 dan vm.vfs_cache_pressure=50..."
mkdir -p /etc/sysctl.d
cat << 'EOF' > /etc/sysctl.d/99-serveropt.conf
# Optimasi Web Server PIDIEKAB
vm.swappiness = 10
vm.vfs_cache_pressure = 50
net.core.somaxconn = 4096
net.ipv4.tcp_fastopen = 3
EOF

sysctl -p /etc/sysctl.d/99-serveropt.conf
`
	out, err = exec.Sudo(swapScript)
	if err != nil {
		log.Printf("Swap error: %v\nOutput: %s", err, out)
	} else {
		fmt.Println(out)
	}

	// Step 3: Cleaning Garbage, Packages, and Systemd Logs
	fmt.Println("\n[3/5] Membersihkan Sampah Sistem, APT Cache, & Log Usang...")
	cleanScript := `
echo " [*] Ukuran Journal Log Sebelum Dibersihkan:"
journalctl --disk-usage

# Bersihkan Journal log di atas 50MB atau lebih dari 7 hari
journalctl --vacuum-size=50M --vacuum-time=7d

# Batasi journald agar tidak membesar liar di masa depan
mkdir -p /etc/systemd/journald.conf.d
cat << 'EOF' > /etc/systemd/journald.conf.d/size.conf
[Journal]
SystemMaxUse=80M
SystemKeepFree=1G
MaxRetentionSec=1month
EOF
systemctl restart systemd-journald

echo " [*] Membersihkan Package Cache & Dependensi Usang (APT)..."
export DEBIAN_FRONTEND=noninteractive
apt-get autoremove --purge -y
apt-get autoclean
apt-get clean

echo " [*] Membersihkan File Temporary & Crash Dumps..."
rm -rf /var/crash/*
rm -rf /tmp/* /tmp/.* 2>/dev/null || true
rm -rf /var/tmp/* /var/tmp/.* 2>/dev/null || true
find /var/log -type f -name "*.gz" -delete
find /var/log -type f -name "*.[0-9]" -delete
`
	out, err = exec.Sudo(cleanScript)
	if err != nil {
		log.Printf("Clean error: %v\nOutput: %s", err, out)
	} else {
		fmt.Println(out)
	}

	// Step 4: Cleaning Go Build Cache & Pruning Stale SSG Releases
	fmt.Println("\n[4/5] Membersihkan Cache Build (Go/Node) & Rilis SSG Lama...")
	appCleanScript := `
echo " [*] Membersihkan Go compiler build cache di server..."
rm -rf /root/.cache/go-build /home/ubuntu/.cache/go-build
rm -rf /root/.cache/yarn /home/ubuntu/.cache/yarn 2>/dev/null || true
rm -rf /root/.npm /home/ubuntu/.npm 2>/dev/null || true

echo " [*] Memeriksa direktori rilis SSG (/var/lib/pidiekab/ssg/releases)..."
if [ -d "/var/lib/pidiekab/ssg/releases" ]; then
    cd /var/lib/pidiekab/ssg/releases
    CURRENT_SYMLINK=$(readlink /var/lib/pidiekab/ssg/current || true)
    CURRENT_BASE=$(basename "$CURRENT_SYMLINK")
    echo " [=] Rilis aktif saat ini: $CURRENT_BASE"
    
    # Simpan 3 rilis terbaru + rilis aktif, hapus yang lebih lama
    RELEASES_TO_DELETE=$(ls -1dt 20* 2>/dev/null | grep -v "$CURRENT_BASE" | tail -n +4)
    if [ -n "$RELEASES_TO_DELETE" ]; then
        echo " [-] Menghapus rilis SSG usang: $RELEASES_TO_DELETE"
        rm -rf $RELEASES_TO_DELETE
    else
        echo " [+] Jumlah rilis SSG sudah efisien (tidak ada folder usang berlebih)."
    fi
fi
`
	out, err = exec.Sudo(appCleanScript)
	if err != nil {
		log.Printf("App clean error: %v\nOutput: %s", err, out)
	} else {
		fmt.Println(out)
	}

	// Step 5: Final Validation & Benchmark
	fmt.Println("\n[5/5] Hasil Akhir Kondisi Server Pasca-Optimasi...")
	finalScript := `
echo "=== RAM & Swap Final ==="
free -h
swapon --show

echo -e "\n=== Kernel Swap & Network Parameters ==="
sysctl vm.swappiness vm.vfs_cache_pressure net.core.somaxconn

echo -e "\n=== Disk Space Final ==="
df -h /

echo -e "\n=== Status Service PidieKab ==="
systemctl is-active pidiekab.service
systemctl status pidiekab.service --no-pager | head -n 12

echo -e "\n=== Uptime & Load Average ==="
uptime
`
	out, err = exec.Sudo(finalScript)
	if err != nil {
		log.Printf("Final check error: %v\nOutput: %s", err, out)
	} else {
		fmt.Println(out)
	}

	fmt.Println("\n==================================================")
	fmt.Println("   OPTIMASI & PEMBERSIHAN SERVER BERHASIL!       ")
	fmt.Println("==================================================")
}

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

const certContent = `-----BEGIN CERTIFICATE-----
MIIEqjCCA5KgAwIBAgIUY8CamyhXU3jr9eJZRyFtNIuJaJ0wDQYJKoZIhvcNAQEL
BQAwgYsxCzAJBgNVBAYTAlVTMRkwFwYDVQQKExBDbG91ZEZsYXJlLCBJbmMuMTQw
MgYDVQQLEytDbG91ZEZsYXJlIE9yaWdpbiBTU0wgQ2VydGlmaWNhdGUgQXV0aG9y
aXR5MRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMRMwEQYDVQQIEwpDYWxpZm9ybmlh
MB4XDTI2MTExODE0NTcwMFoXDTQxMTExNDE0NTcwMFowYjEZMBcGA1UEChMQQ2xv
dWRGbGFyZSwgSW5jLjEdMBsGA1UECxMUQ2xvdWRGbGFyZSBPcmlnaW4gQ0ExJjAk
BgNVBAMTHUNsb3VkRmxhcmUgT3JpZ2luIENlcnRpZmljYXRlMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0edJYWnJa7mnnRb/bOs6PI+kziibYim0Vzu3
UjNhPRzlN56GMf1xlOMcqYAvS7WGDyp89MBRDHY7szdcW7W1dRdillRgVopliEqI
hOZ1lKlfQWDe4OUH7fw07+QtguCObfbfAD2xxch+prxFTqHJmMFWOsxm8+2WnrRl
ZT5/JMj92B/RBUbxj30TzhSHpuhwWO7UmYJuvEsy+pjWNpkSeuXOrUojYq4se3Kb
mrhow7t4Tw8V+85GM0f3iGq9a9Dp0F30PO9ecmJaGM51b7GIXsMDQJKC13n70nCw
G6JGFXOU8cYiVpWalYqi2bHjWfqJjK19H7jvyHW7BCwVaVgvoQIDAQABo4IBLDCC
ASgwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMCBggrBgEFBQcD
ATAMBgNVHRMBAf8EAjAAMB0GA1UdDgQWBBRMxQL1kWxmDgXAk5ysSOc2Q8Y92TAf
BgNVHSMEGDAWgBQk6FNXXXw0QIep65TbuuEWePwppDBABggrBgEFBQcBAQQ0MDIw
MAYIKwYBBQUHMAGGJGh0dHA6Ly9vY3NwLmNsb3VkZmxhcmUuY29tL29yaWdpbl9j
YTAtBgNVHREEJjAkghEqLnBpZGlla2FiLndlYi5pZIIPcGlkaWVrYWIud2ViLmlk
MDgGA1UdHwQxMC8wLaAroCmGJ2h0dHA6Ly9jcmwuY2xvdWRmbGFyZS5jb20vb3Jp
Z2luX2NhLmNybDANBgkqhkiG9w0BAQsFAAOCAQEAIl0ygYftBq2bZ3zQ9lo+NAgN
6UmunfMReUgH2a1mw+cBvQOC5vj7/iARzzYAFHs6E+opQibO48QbbQ8F1eOu3gQb
xpPaP7GoVBityPpWAkT/Sn2W9TUsxj69IQ4LSWtmpUCNV8aryakAeGR5wbYfBwjc
Tiu+EzFQMu6C2MYw8YP7F0oSPPzBUuVq7pneS3bVjSiB/y1jFhAMk+V+fO5N8tkJ
tD6TYvLNL8br3Z4QuuoeePBrwZRRs9ERPzNMAiGjQqjW1mLVfOPVtVIuTwQ3PZsV
GBBvqMm1gebxLi6+MZojKyUAKVV9tWf2xMf/95K2DxVUKo49oUFARx2qIXddNg==
-----END CERTIFICATE-----`

const keyContent = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDR50lhaclruaed
Fv9s6zo8j6TOKJtiKbRXO7dSM2E9HOU3noYx/XGU4xypgC9LtYYPKnz0wFEMdjuz
N1xbtbV1F2KWVGBWimWISoiE5nWUqV9BYN7g5Qft/DTv5C2C4I5t9t8APbHFyH6m
vEVOocmYwVY6zGbz7ZaetGVlPn8kyP3YH9EFRvGPfRPOFIem6HBY7tSZgm68SzL6
mNY2mRJ65c6tSiNirix7cpuauGjDu3hPDxX7zkYzR/eIar1r0OnQXfQ8715yYloY
znVvsYhewwNAkoLXefvScLAbokYVc5TxxiJWlZqViqLZseNZ+omMrX0fuO/IdbsE
LBVpWC+hAgMBAAECgf8WFAB2lpdGasi5OpSzSCq1TrFEXfu6Ci5F3qocTfs0X7hc
bqLUL0Gl1MhMVodNrKeP1iWT44ai4OzL9Plp9rrsuKujVEytzZt7B8aZnq6kwbab
PIKKSozYThGOCgTKIgSD3ID9mCpSu2asSFv3NrKVtqrzPvy1O5BUMSAD04cGv0xl
OTvQTpBiwfpGg8+iiWb3nH/RdDUB8XziutzlgqMjhpEluSCvrcAFbv+fIDb7SnLt
74Hpr6QmuhPIoMZpx/YtrqNwv3czfG4cxziO2yLpB2CdBNrTdXJIpBzF2NGp232k
Efob3s6aeFOvLLoE8WFzvvISf1bhOsKnRk4bPoECgYEA/kzLekJ7Ir5Iz1cmp1v7
BnS9NGeeDZgs/kKls1cf0kN5iEEVLB4PLl8YjTDBlEq7dZHJNwiSerAQaVBeyVuw
n/eTGhNCcLg6mnFvtSYP9HOFyCePnB7AClD/p8BKw6tyzODnHPrSScGBTeXwKxeg
t0KrROKzBvv6CK+T07O2Z4ECgYEA006DJJfTmxKtGUSXr+1NhMpqZI3X3AKzZJfF
/bjE40r31kHV2rd1T5h59Il2dGEt1HoJw3lEMa0THgNj1MXQEsXlc0pY81yIix8Z
0vorKvzINkemm733CA6fed/v8HByZA8aT2p8HkYpfK7Qh8UlZMQzUm2x4Nrhjmes
KReA2CECgYEAq8FijSjp0gWDV8cdLFxLp4ihYgu6DKSDfxe0ovRtYa6BHy4WBr4m
ssM36eTMWkQ/aKEWrjzz8jmIXBkUbBupp+GnUS1A6rEDo9yJ/me+zvRXn2BY9rEu
M8mRYtPVq2LZHteoQUKHVZmEpABxvikvO5ej5drX4BfAQK2Vwp7WzAECgYEAknX7
3jRCHl0+j6IEhZM0xNLMFbBfKwbQMuwj9q3IAkcnK2HsibHZZPcpqL5S1W26J2kD
NXLrOoddFdOBsav+gdSbKmMg7mV32v4OlzaBD0WwvDbDm3V7MSF5li8KjxxRnVr+
gIctSoqVHO/KaCDUv1hivkJ8JX9xEZR4Nu8shiECgYEA3zorlN/r3GeLoZQ5ZNvV
e5x1Nf0baw9YsKUkgA3wa9mH+qrSxVoLL+miKQaChMfxKuwnnZnucMSKA3n2AfHI
iucCicEZqlFqXnoGn9wBJreLe/nBkekPZ5QDHoPVt84oivHcQX5tXsZvft9emgdd
c7Zvb26qFyteKwva8J39hYg=
-----END PRIVATE KEY-----`

func main() {
	host := "43.157.247.26:22"
	user := "ubuntu"
	pass := "shadow-47$-thunder"

	exec, err := NewRemoteExecutor(host, user, pass)
	if err != nil {
		log.Fatalf("SSH connection error: %v", err)
	}
	defer exec.Close()

	fmt.Println("=== 1. Deploying Cloudflare Origin Certificate and Private Key ===")
	writeCertScript := fmt.Sprintf(`cat << 'EOF' > /etc/pidiekab/pidiekab.crt
%s
EOF
cat << 'EOF' > /etc/pidiekab/pidiekab.key
%s
EOF
chown root:pidiekab /etc/pidiekab/pidiekab.crt /etc/pidiekab/pidiekab.key
chmod 0640 /etc/pidiekab/pidiekab.crt /etc/pidiekab/pidiekab.key
`, certContent, keyContent)

	if out, err := exec.Sudo(writeCertScript); err != nil {
		log.Fatalf("Failed to write certificates: %v\nOutput: %s", err, out)
	}
	fmt.Println("Certificates deployed to /etc/pidiekab/pidiekab.crt and /etc/pidiekab/pidiekab.key with 0640 permissions.")

	fmt.Println("\n=== 2. Updating /etc/pidiekab/pidiekab.env for Native TLS (:443) and HTTP Redirect (:80) ===")
	envFileContent := `# PidieKab Production Environment
PIDIEKAB_ENV=production
PIDIEKAB_ADDR=:443
PIDIEKAB_HTTP_REDIRECT_ADDR=:80
PIDIEKAB_TLS_CERT_FILE=/etc/pidiekab/pidiekab.crt
PIDIEKAB_TLS_KEY_FILE=/etc/pidiekab/pidiekab.key
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
	writeEnvScript := fmt.Sprintf(`cat << 'EOF' > /etc/pidiekab/pidiekab.env
%s
EOF
chown root:pidiekab /etc/pidiekab/pidiekab.env
chmod 0640 /etc/pidiekab/pidiekab.env
`, envFileContent)

	if out, err := exec.Sudo(writeEnvScript); err != nil {
		log.Fatalf("Failed to write env file: %v\nOutput: %s", err, out)
	}
	fmt.Println("Environment configuration updated.")

	fmt.Println("\n=== 3. Restarting pidiekab.service ===")
	if out, err := exec.Sudo("systemctl restart pidiekab.service"); err != nil {
		log.Fatalf("Failed to restart service: %v\nOutput: %s", err, out)
	}
	time.Sleep(3 * time.Second)

	fmt.Println("\n=== 4. Verifying Listening Ports (80 & 443) ===")
	ports, _ := exec.Sudo("ss -tulpn | grep pidiekab")
	fmt.Println(ports)

	fmt.Println("\n=== 5. Verifying Service Status ===")
	status, _ := exec.Sudo("systemctl status pidiekab.service --no-pager")
	fmt.Println(status)

	fmt.Println("\n=== 6. Testing HTTPS and HTTP Redirect on Server ===")
	httpTest, _ := exec.Run("curl -s -i http://127.0.0.1/")
	fmt.Printf("HTTP port 80 (should redirect to HTTPS):\n%s\n\n", httpTest)

	httpsTest, _ := exec.Run("curl -s -k -i https://127.0.0.1/health/live")
	fmt.Printf("HTTPS port 443 /health/live:\n%s\n", httpsTest)
}

package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func dialWithProxyFallback(targetHost string, timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", targetHost, 3*time.Second)
	if err == nil {
		return conn, nil
	}
	proxyAddr := "10.15.3.20:80"
	pConn, pErr := net.DialTimeout("tcp", proxyAddr, timeout)
	if pErr != nil {
		return nil, fmt.Errorf("direct failed and proxy failed: %w", pErr)
	}
	req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\nProxy-Connection: Keep-Alive\r\n\r\n", targetHost, targetHost)
	if _, err := pConn.Write([]byte(req)); err != nil {
		pConn.Close()
		return nil, err
	}
	br := bufio.NewReader(pConn)
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		pConn.Close()
		return nil, err
	}
	if resp.StatusCode != 200 {
		pConn.Close()
		return nil, fmt.Errorf("proxy HTTP %s", resp.Status)
	}
	return pConn, nil
}

func main() {
	var cmd string
	if len(os.Args) > 1 {
		cmd = strings.Join(os.Args[1:], " ")
	} else {
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Read stdin failed: %v", err)
		}
		cmd = string(bytes)
	}

	conn, err := dialWithProxyFallback("43.157.247.26:22", 10*time.Second)
	if err != nil {
		log.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	sshConfig := &ssh.ClientConfig{
		User:            "ubuntu",
		Auth:            []ssh.AuthMethod{ssh.Password("shadow-47$-thunder")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, "43.157.247.26:22", sshConfig)
	if err != nil {
		log.Fatalf("SSH handshake failed: %v", err)
	}
	client := ssh.NewClient(ncc, chans, reqs)
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Session failed: %v", err)
	}
	defer session.Close()

	b64 := base64.StdEncoding.EncodeToString([]byte(cmd))
	remoteCmd := fmt.Sprintf("echo %s | base64 -d | bash", b64)

	out, err := session.CombinedOutput(remoteCmd)
	if err != nil {
		fmt.Printf("ERROR (%v): %s\n", err, string(out))
		os.Exit(1)
	}
	fmt.Print(string(out))
}

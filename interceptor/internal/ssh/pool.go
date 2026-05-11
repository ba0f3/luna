package ssh

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"strings"
	"sync"
	"time"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// ExecResult holds the output of a remote command execution.
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// Pool manages a cache of SSH client connections.
// Connections are established lazily and reused across calls.
type Pool struct {
	mu      sync.Mutex
	clients map[string]*gossh.Client
}

// NewPool creates a new SSH connection pool.
func NewPool() *Pool {
	return &Pool{
		clients: make(map[string]*gossh.Client),
	}
}

// parseTarget splits a target string like user@host:port into its components.
func parseTarget(target string) (username, host, port string) {
	username = "root"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}
	port = "22"
	host = target

	if idx := strings.Index(host, "@"); idx != -1 {
		username = host[:idx]
		host = host[idx+1:]
	}

	if strings.Contains(host, ":") {
		h, p, err := net.SplitHostPort(host)
		if err == nil {
			host = h
			port = p
		}
	}

	return username, host, port
}

// Execute runs command on the named target and returns the result.
func (p *Pool) Execute(target, command string, timeout time.Duration) (ExecResult, error) {
	client, err := p.getClient(target)
	if err != nil {
		return ExecResult{}, err
	}

	session, err := client.NewSession()
	if err != nil {
		p.evict(target)
		client, err = p.getClient(target)
		if err != nil {
			return ExecResult{}, err
		}
		session, err = client.NewSession()
		if err != nil {
			return ExecResult{}, fmt.Errorf("create SSH session: %w", err)
		}
	}
	defer session.Close()

	var stdoutBuf, stderrBuf strings.Builder
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	start := time.Now()

	type result struct {
		err      error
		exitCode int
	}
	ch := make(chan result, 1)
	go func() {
		runErr := session.Run(command)
		code := 0
		if runErr != nil {
			if exitErr, ok := runErr.(*gossh.ExitError); ok {
				code = exitErr.ExitStatus()
				runErr = nil
			}
		}
		ch <- result{err: runErr, exitCode: code}
	}()

	select {
	case res := <-ch:
		dur := time.Since(start)
		if res.err != nil {
			return ExecResult{}, fmt.Errorf("SSH run: %w", res.err)
		}
		return ExecResult{
			Stdout:   stdoutBuf.String(),
			Stderr:   stderrBuf.String(),
			ExitCode: res.exitCode,
			Duration: dur,
		}, nil
	case <-time.After(timeout):
		session.Signal(gossh.SIGKILL) //nolint:errcheck
		return ExecResult{ExitCode: -1}, fmt.Errorf("command timed out after %s", timeout)
	}
}

// getClient returns a cached or newly-dialed client for the given target.
func (p *Pool) getClient(target string) (*gossh.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if c, ok := p.clients[target]; ok {
		if _, _, err := c.SendRequest("keepalive@openssh.com", true, nil); err == nil {
			return c, nil
		}
		c.Close() //nolint:errcheck
		delete(p.clients, target)
	}

	username, host, port := parseTarget(target)
	addr := net.JoinHostPort(host, port)

	authMethods, err := buildAuthMethods()
	if err != nil {
		return nil, fmt.Errorf("build auth for %q: %w", target, err)
	}

	khCallback, err := buildKnownHostsCallback()
	if err != nil {
		return nil, fmt.Errorf("load known_hosts: %w", err)
	}

	sshCfg := &gossh.ClientConfig{
		User:            username,
		Auth:            authMethods,
		HostKeyCallback: khCallback,
		Timeout:         15 * time.Second,
	}

	log.Printf("connecting to %s@%s", username, addr)
	client, err := gossh.Dial("tcp", addr, sshCfg)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", target, err)
	}

	p.clients[target] = client
	return client, nil
}

// evict removes a target's cached client without closing it.
func (p *Pool) evict(target string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.clients, target)
}

// Close shuts down all cached connections.
func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for alias, c := range p.clients {
		if err := c.Close(); err != nil {
			log.Printf("close %s: %v", alias, err)
		}
	}
	p.clients = make(map[string]*gossh.Client)
}

// buildAuthMethods assembles SSH auth methods using agent and default keys.
func buildAuthMethods() ([]gossh.AuthMethod, error) {
	var methods []gossh.AuthMethod

	if agentMethod := sshAgentAuth(); agentMethod != nil {
		methods = append(methods, agentMethod)
	}

	for _, name := range []string{"id_ed25519", "id_rsa", "id_ecdsa"} {
		path := fmt.Sprintf("%s/.ssh/%s", mustHome(), name)
		if _, err := os.Stat(path); err == nil {
			if signer, err := loadPrivateKey(path); err == nil {
				methods = append(methods, gossh.PublicKeys(signer))
			}
		}
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no SSH auth methods available (no agent, no default keys)")
	}
	return methods, nil
}

// loadPrivateKey reads and parses a PEM-encoded private key file.
func loadPrivateKey(path string) (gossh.Signer, error) {
	pem, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key %q: %w", path, err)
	}
	signer, err := gossh.ParsePrivateKey(pem)
	if err != nil {
		return nil, fmt.Errorf("parse key %q: %w", path, err)
	}
	return signer, nil
}

// sshAgentAuth returns an auth method that delegates to the running SSH agent.
func sshAgentAuth() gossh.AuthMethod {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return nil
	}
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil
	}
	agentClient := agent.NewClient(conn)
	return gossh.PublicKeysCallback(agentClient.Signers)
}

// buildKnownHostsCallback creates a host key callback from ~/.ssh/known_hosts.
func buildKnownHostsCallback() (gossh.HostKeyCallback, error) {
	khPath := fmt.Sprintf("%s/.ssh/known_hosts", mustHome())
	if _, err := os.Stat(khPath); os.IsNotExist(err) {
		log.Printf("WARN: %s not found; new host connections will be rejected", khPath)
		return func(hostname string, remote net.Addr, key gossh.PublicKey) error {
			return fmt.Errorf("host %q not in known_hosts and file is missing — run: ssh-keyscan %s >> ~/.ssh/known_hosts", hostname, hostname)
		}, nil
	}
	return knownhosts.New(khPath)
}

func mustHome() string {
	h, _ := os.UserHomeDir()
	return h
}

package rssh

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sshtools/internal/pkg/rsftp"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type ClientConfig struct {
	Addr           string `json:"addr" mapstructure:"addr"`
	Username       string `json:"username" mapstructure:"username"`
	Password       string `json:"password" mapstructure:"password"`
	PrivateKeyPath string `json:"privateKeyPath" mapstructure:"privateKeyPath"`
}

type Response struct {
	Addr       string
	Output     string
	ExitStatus int
	Err        error
}

type Client struct {
	*ssh.Client

	Addr string
}

func NewClient(cfg ClientConfig) (*Client, error) {
	sshConfig := &ssh.ClientConfig{
		User:            cfg.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 2,
	}

	if cfg.Password != "" {
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(cfg.Password),
		}
	} else if cfg.PrivateKeyPath != "" {
		pemBytes, err := os.ReadFile(cfg.PrivateKeyPath)
		if err != nil {
			return nil, err
		}
		singner, err := ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			return nil, err
		}
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(singner),
		}
	} else {
		return nil, errors.New("Provide the password or privateKeyPath.")
	}

	sshClient, err := ssh.Dial("tcp", cfg.Addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect %s, %s", cfg.Addr, err)
	}

	c := &Client{
		Client: sshClient,
		Addr:   cfg.Addr,
	}

	return c, nil
}

func (c *Client) ExecCmd(cmd string, ch chan<- Response) {
	session, err := c.NewSession()
	defer session.Close()
	if err != nil {
		ch <- Response{
			Addr:       c.Addr,
			Output:     "",
			ExitStatus: -1,
			Err:        fmt.Errorf("failed to create session, %s", err),
		}
		return
	}

	outByte, err := session.CombinedOutput(cmd)
	if err != nil {
		exitErr, ok := err.(*ssh.ExitError)
		if !ok {
			ch <- Response{
				Addr:       c.Addr,
				Output:     "",
				ExitStatus: -1,
				Err:        fmt.Errorf("unable to execute command, %s", err),
			}
			return
		}
		ch <- Response{
			Addr:       c.Addr,
			Output:     "",
			ExitStatus: exitErr.ExitStatus(),
			Err:        fmt.Errorf("Failed to execute command '%s', %s", cmd, err),
		}
		return
	}

	ch <- Response{
		Addr:       c.Addr,
		Output:     string(outByte),
		ExitStatus: 0,
		Err:        nil,
	}
}

// Upload shell script to remote SSH server and execute it
func (c *Client) ExecShellScript(localFile string, ch chan<- Response) {
	sc, err := rsftp.NewClient(c.Client, c.Addr)
	if err != nil {
		ch <- Response{
			Addr:       c.Addr,
			Output:     "",
			ExitStatus: -1,
			Err:        err,
		}
		return
	}

	name := fmt.Sprintf("%s_%s", time.Now().Format("20060102150405"), filepath.Base(localFile))
	remoteFile := filepath.Join("/tmp/scripts", name)
	if err := sc.UploadFile(localFile, remoteFile, true); err != nil {
		ch <- Response{
			Addr:       c.Addr,
			Output:     "",
			ExitStatus: -1,
			Err:        err,
		}
		return
	}

	var shell string
	shells := []string{"/bin/bash", "/bin/sh"}
	for _, s := range shells {
		if _, err := sc.Stat(s); err == nil {
			shell = s
			break
		}
	}
	if shell == "" {
		ch <- Response{
			Addr:       c.Addr,
			Output:     "",
			ExitStatus: -1,
			Err:        fmt.Errorf("These files '%s' do not exists on the remote ssh server", strings.Join(shells, ",")),
		}
		return
	}

	command := fmt.Sprintf("%s %s", shell, remoteFile)
	c.ExecCmd(command, ch)
}

func newClientWithChannel(cfg ClientConfig, ch chan<- *Client, errch chan<- error) {
	client, err := NewClient(cfg)
	errch <- err
	if err != nil {
		return
	}
	ch <- client
}

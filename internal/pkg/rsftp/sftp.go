package rsftp

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type ClientConfig struct {
	Addr           string `json:"addr" mapstructure:"addr"`
	Username       string `json:"username" mapstructure:"username"`
	Password       string `json:"password" mapstructure:"password"`
	PrivateKeyPath string `json:"privateKeyPath" mapstructure:"privateKeyPath"`
}

type Response struct {
	Addr      string
	Output    string
	Err       error
	FileInfos []fs.FileInfo
}

type Client struct {
	*sftp.Client

	Addr string
}

func NewClient(conn *ssh.Client, addr string) (*Client, error) {
	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client, %s", err)
	}

	c := &Client{
		Client: sftpClient,
		Addr:   addr,
	}
	return c, nil
}

func NewForConfig(cfg ClientConfig) (*Client, error) {
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

	return NewClient(sshClient, cfg.Addr)
}

func (c *Client) ListFiles(remotePath string) ([]Response, error) {
	return nil, nil
}

// UploadFile upload file from local to remote SSH server
func (c *Client) UploadFile(localFile, remoteFile string, force bool) error {
	localInfo, err := os.Stat(localFile)
	if err != nil {
		return fmt.Errorf("local %s file is not exist, %s", localFile, err)
	}

	if localInfo.IsDir() {
		return fmt.Errorf("%s is directory, require a file", localFile)
	}

	if _, err := c.Stat(remoteFile); err == nil && !force {
		return fmt.Errorf("remote file %s already exists", remoteFile)
	}

	if err := c.MkdirAll(filepath.Dir(remoteFile)); err != nil {
		return err
	}

	content, err := os.ReadFile(localFile)
	if err != nil {
		return err
	}

	f, err := c.Create(remoteFile)
	if err != nil {
		return err
	}
	_, err = f.Write(content)

	return err
}

// UploadFiles upload file or directory from local to remote SSH server
func (c *Client) UploadFiles(localPath, remotePath string, force bool, ch chan<- Response) {
	if _, err := os.Stat(localPath); err != nil {
		ch <- Response{
			Addr:   c.Addr,
			Output: "",
			Err:    fmt.Errorf("local %s is not exist, %s", localPath, err),
		}
		return
	}

	remoteInfo, err := c.Stat(remotePath)
	if err == nil && remoteInfo.IsDir() {
		remotePath = filepath.Join(remotePath, filepath.Base(localPath))
	}

	err = filepath.Walk(localPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			remoteDir := filepath.Join(remotePath, path[len(localPath):])
			if err := c.MkdirAll(remoteDir); err != nil {
				return err
			}
			return nil
		}

		remoteFile := filepath.Join(remotePath, path[len(localPath):])
		return c.UploadFile(path, remoteFile, force)
	})

	if err != nil {
		ch <- Response{
			Addr:   c.Addr,
			Output: "",
			Err:    err,
		}
		return
	}

	ch <- Response{
		Addr:   c.Addr,
		Output: fmt.Sprintf("%s -> %s:%s", localPath, c.Addr, remotePath),
		Err:    nil,
	}
}

// DownloadFile download file from remote SSH server to local
func (c *Client) DownloadFile(localFile, remoteFile string) error {
	remoteInfo, err := c.Stat(remoteFile)
	if err != nil {
		return fmt.Errorf("remote file %s is not exist, %s", remoteFile, err)
	}

	if remoteInfo.IsDir() {
		return fmt.Errorf("%s is directory, require a file", remoteFile)
	}

	if _, err := os.Stat(localFile); err == nil {
		return fmt.Errorf("local file %s already exists", localFile)
	}

	if err := os.MkdirAll(filepath.Dir(localFile), 0755); err != nil {
		return err
	}

	rf, err := c.Open(remoteFile)
	if err != nil {
		return err
	}
	content, err := ioutil.ReadAll(rf)
	if err != nil {
		return err
	}

	lf, err := os.Create(localFile)
	if err != nil {
		return err
	}
	_, err = lf.Write(content)

	return err
}

// DownloadFiles download file or directory from remote SSH server to local
func (c *Client) DownloadFiles(localPath, remotePath string, ch chan<- Response) {
	if _, err := c.Stat(remotePath); err != nil {
		ch <- Response{
			Addr:   c.Addr,
			Output: "",
			Err:    fmt.Errorf("remote path %s is not exist, %s", remotePath, err),
		}
		return
	}

	localInfo, err := c.Stat(localPath)
	if err == nil && localInfo.IsDir() {
		localPath = filepath.Join(localPath, filepath.Base(remotePath))
	}

	w := c.Walk(remotePath)
	for w.Step() {
		if w.Err() != nil {
			ch <- Response{
				Addr:   c.Addr,
				Output: "",
				Err:    err,
			}
			return
		}

		path := w.Path()
		if w.Stat().IsDir() {
			localDir := filepath.Join(localPath, path[len(remotePath):])
			if err := os.MkdirAll(localDir, w.Stat().Mode()); err != nil {
				ch <- Response{
					Addr:   c.Addr,
					Output: "",
					Err:    err,
				}
				return
			}
			continue
		}

		localFile := filepath.Join(localPath, path[len(remotePath):])
		if err := c.DownloadFile(localFile, path); err != nil {
			ch <- Response{
				Addr:   c.Addr,
				Output: "",
				Err:    err,
			}
			return
		}
	}

	ch <- Response{
		Addr:   c.Addr,
		Output: fmt.Sprintf("%s:%s -> %s", c.Addr, remotePath, localPath),
		Err:    nil,
	}
}

func newClientWithChannel(cfg ClientConfig, ch chan<- *Client, errch chan<- error) {
	client, err := NewForConfig(cfg)
	errch <- err
	if err != nil {
		return
	}
	ch <- client
}

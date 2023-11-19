package rsftp

import (
	"fmt"
	"path/filepath"
	"sync"
)

type MultiClient struct {
	clients []*Client
}

func NewMultiClient(cfgs []ClientConfig) (*MultiClient, error) {
	clients := []*Client{}

	clientChan := make(chan *Client, len(cfgs))
	errChan := make(chan error, len(cfgs))
	var wg sync.WaitGroup
	for _, config := range cfgs {
		wg.Add(1)
		cfg := config
		go func() {
			defer wg.Done()
			newClientWithChannel(cfg, clientChan, errChan)
		}()
	}
	wg.Wait()
	close(clientChan)
	close(errChan)

	for c := range clientChan {
		clients = append(clients, c)
	}

	mc := &MultiClient{
		clients: clients,
	}

	return mc, nil
}

func (mc *MultiClient) UploadFiles(localPath, remotePath string, force bool) []Response {
	var wg sync.WaitGroup

	respChan := make(chan Response, len(mc.clients))
	for _, client := range mc.clients {
		wg.Add(1)
		c := client
		go func() {
			defer wg.Done()
			c.UploadFiles(localPath, remotePath, force, respChan)
		}()
	}
	wg.Wait()
	close(respChan)

	resps := []Response{}
	for resp := range respChan {
		resps = append(resps, resp)
	}

	return resps
}

func (mc *MultiClient) DownloadFiles(localPath, remotePath string) []Response {
	var wg sync.WaitGroup

	respChan := make(chan Response, len(mc.clients))
	for _, client := range mc.clients {
		wg.Add(1)
		c := client
		lp := filepath.Join(localPath, c.Addr)
		go func() {
			defer wg.Done()
			c.DownloadFiles(lp, remotePath, respChan)
		}()
	}
	wg.Wait()
	close(respChan)

	resps := []Response{}
	for resp := range respChan {
		resps = append(resps, resp)
	}

	return resps
}

func (mc *MultiClient) Close() error {
	var err error
	for _, c := range mc.clients {
		if e := c.Close(); e != nil {
			err = fmt.Errorf("%s; %s close failed, %s", err, c.Addr, e)
		}
	}
	return err
}

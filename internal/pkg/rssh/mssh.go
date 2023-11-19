package rssh

import (
	"fmt"
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

	var errs error
	for err := range errChan {
		if err != nil {
			errs = fmt.Errorf("%v;%v", errs, err)
		}
	}

	if errs != nil {
		return nil, errs
	}

	for c := range clientChan {
		clients = append(clients, c)
	}

	mc := &MultiClient{
		clients: clients,
	}

	return mc, nil
}

func (mc *MultiClient) execCmdOrScript(s string, isScript bool) []Response {
	var wg sync.WaitGroup

	respChan := make(chan Response, len(mc.clients))
	for _, client := range mc.clients {
		wg.Add(1)
		c := client
		go func() {
			defer wg.Done()
			if isScript {
				c.ExecShellScript(s, respChan)
			} else {
				c.ExecCmd(s, respChan)
			}
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

func (mc *MultiClient) ExecCmd(cmd string) []Response {
	return mc.execCmdOrScript(cmd, false)
}

func (mc *MultiClient) ExecShellScript(localFile string) []Response {
	return mc.execCmdOrScript(localFile, true)
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

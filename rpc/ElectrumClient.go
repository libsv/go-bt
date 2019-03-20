package rpc

import (
	"encoding/json"
	"fmt"
)

// Methods
const (
	SUBSCRIBE   = "server.peers.subscribe"
	LISTUNSPENT = "blockchain.address.listunspent"
)

// A ElectrumClient represents a Electrum client
type ElectrumClient struct {
	client *rpcClient
}

// NewElectrumClient return a new ElectrumClient
func NewElectrumClient() (*ElectrumClient, error) {
	rpcClient, err := newClient("electron.bitcoinsv.io:50002")
	if err != nil {
		return nil, err
	}
	return &ElectrumClient{rpcClient}, nil
}

// Close closes the underlying socket
func (ec *ElectrumClient) Close() {
	ec.client.close()
}

// GetServers comment
func (ec *ElectrumClient) GetServers() ([]string, error) {
	r, err := ec.client.call(SUBSCRIBE, []string{})

	if err != nil {
		return nil, err
	}

	if r.Err != nil {
		return nil, fmt.Errorf("%v", r.Err)
	}

	var data []interface{}
	err = json.Unmarshal(r.Result, &data)
	if err != nil {
		return nil, err
	}

	var servers []string
	for _, i := range data {
		ip := i.([]interface{})[0]
		servers = append(servers, ip.(string))

	}
	return servers, nil
}

// ListUnspent comment
func (ec *ElectrumClient) ListUnspent(addresses []string) error {
	r, err := ec.client.call(LISTUNSPENT, addresses)

	if err != nil {
		return err
	}

	if r.Err != nil {
		return fmt.Errorf("%v", r.Err)
	}

	var data interface{}
	err = json.Unmarshal(r.Result, &data)
	if err != nil {
		return err
	}

	fmt.Printf("%#v\n", data)
	return nil
}

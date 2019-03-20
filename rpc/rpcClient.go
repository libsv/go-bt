package rpc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"time"
)

type rpcClient struct {
	conn *tls.Conn
}

type rpcRequest struct {
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int64       `json:"id"`
	JSONRpc string      `json:"jsonrpc"`
}

type rpcResponse struct {
	ID     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
	Err    interface{}     `json:"error"`
}

func newClient(addr string) (*rpcClient, error) {
	if len(addr) == 0 {
		return nil, errors.New("Bad call missing argument addr")
	}

	config := tls.Config{InsecureSkipVerify: true}

	conn, err := tls.Dial("tcp", addr, &config)
	if err != nil {
		return nil, err
	}

	return &rpcClient{conn}, nil
}

func (c *rpcClient) close() {
	c.conn.Close()
}

func (c *rpcClient) call(method string, params interface{}) (resp rpcResponse, err error) {
	req := rpcRequest{
		Method:  method,
		Params:  params,
		ID:      time.Now().UnixNano(),
		JSONRpc: "2.0",
	}

	buffer := &bytes.Buffer{}
	jsonEncoder := json.NewEncoder(buffer)
	err = jsonEncoder.Encode(req)
	if err != nil {
		return
	}

	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	n, err := io.WriteString(c.conn, buffer.String())
	if err != nil {
		return
	}

	reply := make([]byte, 10240)

	c.conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	n, err = c.conn.Read(reply)
	if err != nil {
		return
	}

	err = json.Unmarshal(reply[:n], &resp)
	if err != nil {
		return
	}

	if resp.ID != req.ID {
		err = errors.New("RPC ID mismatch")
		return
	}

	return
}

package rpc

import "testing"

func TestSubscribe(t *testing.T) {
	ec, _ := NewElectrumClient()
	defer ec.Close()

	d, err := ec.GetServers()
	if err != nil {
		t.Error(err)
	}

	t.Log(d)
}

func TestListUnspent(t *testing.T) {
	ec, _ := NewElectrumClient()
	defer ec.Close()

	ec.ListUnspent([]string{"1FjKDzsT1aXTpKB3GUHjP21xuNEc7o4N4S"})

}

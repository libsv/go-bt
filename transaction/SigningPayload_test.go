package transaction

import (
	"encoding/json"
	"reflect"
	"testing"
)

func Test(t *testing.T) {
	payload := NewSigningPayload()

	payload.AddItem("simon", "bob")

	j, err := json.Marshal(payload)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%s\n", string(j))
	for _, p := range *payload {
		t.Logf("%+v\n", p)
	}
}

func TestUnmarshall(t *testing.T) {
	j := `[{"address":"simon","sigHash":"bob","publicKey":null,"signature":null}]`

	var payload SigningPayload

	err := json.Unmarshal([]byte(j), &payload)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%s\n", string(j))
	for _, p := range payload {
		t.Logf("%+v\n", p)
	}
}

func TestSigningPayloadFromTx(t *testing.T) {
	unsignedTx := "010000000236916d2d420bbd4ff8cd94a2b49d89daeeaeeedbf640cd2c9aa0c619bd806209000000001976a914bcd0bdbf5fcde5ed957396752d4bd2e01d36870288acffffffff3fdb6bf215bad39941525500337e9e7924f99da5a841c5dc7c1eab8036162fe2000000001976a914bcd0bdbf5fcde5ed957396752d4bd2e01d36870288acffffffff0380d1f008000000001976a91490d7b4c4df77b035616e53e2f3701ab562d6f87f88ac80f0fa02000000001976a91490e5bc4b4b5391b60c3fa9b568f916fa83819fce88ac000000000000000020006a1d536f6d652064617461203132333435363738383930206162636465666700000000"
	tx, err := NewFromString(unsignedTx)
	if err != nil {
		t.Error(err)
		return
	}

	signingPayload := NewSigningPayload()

	signingPayload.AddItem("bcd0bdbf5fcde5ed957396752d4bd2e01d368702", "80448cea404b51f82d409cbd1fbca66bf43fe1cd45d7660953e39ce3c5d8208d")
	signingPayload.AddItem("bcd0bdbf5fcde5ed957396752d4bd2e01d368702", "c62573ac749d9b202cd7b2e0d36a0f688a680810a70ee840f6de7bab4d615095")

	tx.Inputs[0].PreviousTxSatoshis = uint64(100000000)
	tx.Inputs[1].PreviousTxSatoshis = uint64(100000000)
	tx.Inputs[0].PreviousTxScript = NewScriptFromString("76a914bcd0bdbf5fcde5ed957396752d4bd2e01d36870288ac")
	tx.Inputs[1].PreviousTxScript = NewScriptFromString("76a914bcd0bdbf5fcde5ed957396752d4bd2e01d36870288ac")

	payload, err := tx.GetSighashPayload(0)
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(payload, signingPayload) {
		t.Errorf("Error payload created is not as expected.  GOT%+v \nEXPECTED%v+", payload, signingPayload)
	}
}

package transaction

import (
	"encoding/json"
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
	for _, p := range payload {
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

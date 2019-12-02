package transaction

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"bitbucket.org/simon_ordish/cryptolib"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
)

func TestRegTestCoinbase(t *testing.T) {
	hex := "02000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0e5101010a2f4542323030302e302fffffffff0100f2052a01000000232103db233bb9fc387d78b133ec904069d46e95ff17da657671b44afa0bc64e89ac18ac00000000"
	bt, err := NewFromString(hex)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Iscoinbase %t", bt.IsCoinbase())
	t.Logf("input count %d", bt.InputCount())

}

func TestGetVersion(t *testing.T) {
	const tx = "01000000014c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff021d784500000000001976a914e9b62e25d4c6f97287dfe62f8063b79a9638c84688ac60d64f00000000001976a914bb4bca2306df66d72c6e44a470873484d8808b8888ac00000000"
	bt, err := NewFromString(tx)
	if err != nil {
		t.Error(err)
		return
	}

	res := bt.Version
	if res != 1 {
		t.Errorf("Expecting 1, got %d", res)
	}
}

func TestConvertXPriv(t *testing.T) {
	const xprv = "xprv9s21ZrQH143K2beTKhLXFRWWFwH8jkwUssjk3SVTiApgmge7kNC3jhVc4NgHW8PhW2y7BCDErqnKpKuyQMjqSePPJooPJowAz5BVLThsv6c"
	const expected = "5f86e4023a4e94f00463f81b70ff951f83f896a0a3e6ed89cf163c152f954f8b"

	r, _ := cryptolib.NewPrivateKey(xprv)

	t.Logf("%x", r.PrivateKey)
}

func TestSignRedeemScript(t *testing.T) {
	var redeemScript, _ = hex.DecodeString("524c53ff0488b21e000000000000000000362f7a9030543db8751401c387d6a71e870f1895b3a62569d455e8ee5f5f5e5f03036624c6df96984db6b4e625b6707c017eb0e0d137cd13a0c989bfa77a4473fd000000004c53ff0488b21e0000000000000000008b20425398995f3c866ea6ce5c1828a516b007379cf97b136bffbdc86f75df14036454bad23b019eae34f10aff8b8d6d8deb18cb31354e5a169ee09d8a4560e8250000000052ae")
	const expected = "3044022041682b268531cf6209577deae34b92fdc83d9ef6e3abc190d4952e927761efd502201696256fba4dd6b05e44ed3871abbd1bc11356aea5ddc36816ca779f68cca6fa"

	const xprv = "xprv9s21ZrQH143K2beTKhLXFRWWFwH8jkwUssjk3SVTiApgmge7kNC3jhVc4NgHW8PhW2y7BCDErqnKpKuyQMjqSePPJooPJowAz5BVLThsv6c"
	const privHex = "5f86e4023a4e94f00463f81b70ff951f83f896a0a3e6ed89cf163c152f954f8b"

	pkBytes, err := hex.DecodeString(privHex)
	if err != nil {
		fmt.Println(err)
		return
	}
	privKey, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), pkBytes)

	// Sign a message using the private key.
	messageHash := chainhash.DoubleHashB(redeemScript)
	signature, err := privKey.Sign(messageHash)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Serialize and display the signature.
	fmt.Printf("Serialized Signature: %x\n", signature.Serialize())

	// Verify the signature for the message using the public key.
	verified := signature.Verify(messageHash, pubKey)
	fmt.Printf("Signature Verified? %v\n", verified)
}

func TestIsCoinbase(t *testing.T) {
	const coinbase = "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4303bfea07322f53696d6f6e204f726469736820616e642053747561727420467265656d616e206d61646520746869732068617070656e2f9a46434790f7dbdea3430000ffffffff018a08ac4a000000001976a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac00000000"
	bt1, err := NewFromString(coinbase)
	if err != nil {
		t.Error(err)
		return
	}

	cb1 := bt1.IsCoinbase()
	if cb1 == false {
		t.Errorf("Expecting true, got %t", cb1)
	}

	const tx = "01000000014c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff021d784500000000001976a914e9b62e25d4c6f97287dfe62f8063b79a9638c84688ac60d64f00000000001976a914bb4bca2306df66d72c6e44a470873484d8808b8888ac00000000"
	bt2, err := NewFromString(tx)
	if err != nil {
		t.Error(err)
		return
	}

	cb2 := bt2.IsCoinbase()
	if cb2 == true {
		t.Errorf("Expecting false, got %t", cb2)
	}
}

func TestGetSighashPayload(t *testing.T) {
	unsignedTx := "01000000017e419b1b2dc7d7988bf2c982878d7719bee096d31111a72d1c7470e5ab7d1a5b0000000000ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde47e976000000001976a9148fe80c75c9560e8b56ed64ea3c26e18d2c52211b88ac00000000"
	tx, err := NewFromString(unsignedTx)
	// Previous txid 5b1a7dabe570741c2da71111d396e0be19778d8782c9f28b98d7c72d1b9b417e

	//Add the UTXO amount and script.
	tx.Inputs[0].PreviousTxSatoshis = 2000000000
	tx.Inputs[0].PreviousTxScript = NewScriptFromString("76a9148fe80c75c9560e8b56ed64ea3c26e18d2c52211b88ac")
	// fmt.Printf("%x\n", tx.Hex())
	// tx with input 01000000017e419b1b2dc7d7988bf2c982878d7719bee096d31111a72d1c7470e5ab7d1a5b000000001976a9148fe80c75c9560e8b56ed64ea3c26e18d2c52211b88acffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde47e976000000001976a9148fe80c75c9560e8b56ed64ea3c26e18d2c52211b88ac00000000

	sigType := uint32(SighashAll | SighashForkID)
	sigHashes, err := tx.GetSighashPayload(sigType)
	if err != nil {
		t.Error(err)
	}

	if len(*sigHashes) != 1 {
		t.Errorf("Error expected payload to be 1 item long, got %d", len(*sigHashes))
	}

	expectedPayload := NewSigningPayload()
	// Add the expected payload for the single input.
	publicKeyhash := "8fe80c75c9560e8b56ed64ea3c26e18d2c52211b" // This is the PKH for address mtdruWYVEV1wz5yL7GvpBj4MgifCB7yhPd
	expectedPayload.AddItem(publicKeyhash, "8ea09cb667b276a886b79d8d6b7d073cc88e64f1640dc9bfd400f9301d4aaa98")

	if !reflect.DeepEqual(*sigHashes, *expectedPayload) {
		t.Errorf("Error expected payload does not match actual, \n      got %+v\n, \nexpected, %+v\n", (*sigHashes)[0], (*expectedPayload)[0])
	}

	if err != nil {
		t.Error(err)
		return
	}
}

func TestApplySignatures(t *testing.T) {

	unsignedTx := "010000000236916d2d420bbd4ff8cd94a2b49d89daeeaeeedbf640cd2c9aa0c619bd806209000000001976a914bcd0bdbf5fcde5ed957396752d4bd2e01d36870288acffffffff3fdb6bf215bad39941525500337e9e7924f99da5a841c5dc7c1eab8036162fe2000000001976a914bcd0bdbf5fcde5ed957396752d4bd2e01d36870288acffffffff0380d1f008000000001976a91490d7b4c4df77b035616e53e2f3701ab562d6f87f88ac80f0fa02000000001976a91490e5bc4b4b5391b60c3fa9b568f916fa83819fce88ac000000000000000020006a1d536f6d652064617461203132333435363738383930206162636465666700000000"
	tx, err := NewFromString(unsignedTx)
	if err != nil {
		t.Error(err)
		return
	}

	signingPayload := SigningPayload{}

	// Append a valid response received from the signing service for this Tx.
	signingItem := SigningItem{
		PublicKeyHash: "bcd0bdbf5fcde5ed957396752d4bd2e01d368702",
		SigHash:       "80448cea404b51f82d409cbd1fbca66bf43fe1cd45d7660953e39ce3c5d8208d",
		PublicKey:     "02ba6bc6906e4937bcde60dbbabdd994dbd0c23e86d834a856091efe677be378b1",
		Signature:     "3045022100a0a005f339978dd6945e44d524d576189f8f7546f41c4899beaa796facb0c4c40220719de9a73796d604b9ee32d7496234c488705fa73f0bd2ffeadcca57580f4cb3",
	}
	signingPayload = append(signingPayload, &signingItem)

	signingItem2 := SigningItem{
		PublicKeyHash: "bcd0bdbf5fcde5ed957396752d4bd2e01d368702",
		SigHash:       "c62573ac749d9b202cd7b2e0d36a0f688a680810a70ee840f6de7bab4d615095",
		PublicKey:     "02ba6bc6906e4937bcde60dbbabdd994dbd0c23e86d834a856091efe677be378b1",
		Signature:     "30440220399173272f0f56c06b4eb1ccce970603e305988788ab1468e0948ae340fc5380022067684423502f75c5b6e88ad302cc2a1cf739c824efbd5e83fa9e02d4b2975f64",
	}
	signingPayload = append(signingPayload, &signingItem2)

	tx, err = tx.ApplySignatures(&signingPayload, 0)
	if err != nil {
		t.Error(err)
		return
	}

	signedTxFromRegtest := "010000000236916d2d420bbd4ff8cd94a2b49d89daeeaeeedbf640cd2c9aa0c619bd806209000000006b483045022100a0a005f339978dd6945e44d524d576189f8f7546f41c4899beaa796facb0c4c40220719de9a73796d604b9ee32d7496234c488705fa73f0bd2ffeadcca57580f4cb3412102ba6bc6906e4937bcde60dbbabdd994dbd0c23e86d834a856091efe677be378b1ffffffff3fdb6bf215bad39941525500337e9e7924f99da5a841c5dc7c1eab8036162fe2000000006a4730440220399173272f0f56c06b4eb1ccce970603e305988788ab1468e0948ae340fc5380022067684423502f75c5b6e88ad302cc2a1cf739c824efbd5e83fa9e02d4b2975f64412102ba6bc6906e4937bcde60dbbabdd994dbd0c23e86d834a856091efe677be378b1ffffffff0380d1f008000000001976a91490d7b4c4df77b035616e53e2f3701ab562d6f87f88ac80f0fa02000000001976a91490e5bc4b4b5391b60c3fa9b568f916fa83819fce88ac000000000000000020006a1d536f6d652064617461203132333435363738383930206162636465666700000000"

	if hex.EncodeToString(tx.Hex()) != signedTxFromRegtest {
		t.Errorf("Error - tx with sigs applied does not match expcted signed tx from regtest.\nGot %s\nexpected %s\n", hex.EncodeToString(tx.Hex()), signedTxFromRegtest)
	}
}

// We don't expect to use the local tx.sign() function as the signing service does it, but we include it for completeness using the same methods.
func TestSignTx(t *testing.T) {
	unsignedTx := "010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d25072326510000000000ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000"
	tx, err := NewFromString(unsignedTx)

	//Add the UTXO amount and script.
	tx.Inputs[0].PreviousTxSatoshis = 100000000
	tx.Inputs[0].PreviousTxScript = NewScriptFromString("76a914c0a3c167a28cabb9fbb495affa0761e6e74ac60d88ac")

	// Our private key.
	wif, err := btcutil.DecodeWIF("cNGwGSc7KRrTmdLUZ54fiSXWbhLNDc2Eg5zNucgQxyQCzuQ5YRDq")
	if err != nil {
		t.Error(err)
		return
	}
	tx.Sign(wif.PrivKey, 0)
	expectedSignedTx := "010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d2507232651000000006b483045022100c1d77036dc6cd1f3fa1214b0688391ab7f7a16cd31ea4e5a1f7a415ef167df820220751aced6d24649fa235132f1e6969e163b9400f80043a72879237dab4a1190ad412103b8b40a84123121d260f5c109bc5a46ec819c2e4002e5ba08638783bfb4e01435ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000"

	if hex.EncodeToString(tx.Hex()) != expectedSignedTx {
		t.Errorf("Expecting %s\n, got %s\n", expectedSignedTx, hex.EncodeToString(tx.Hex()))
	}

	if unsignedTx == expectedSignedTx {
		t.Errorf("Expected and signed TX strings in code identical")
	}
}

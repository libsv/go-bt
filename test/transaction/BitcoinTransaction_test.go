package transaction

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/libsv/libsv/keys"
	"github.com/libsv/libsv/script"
	"github.com/libsv/libsv/transaction"
	"github.com/libsv/libsv/utils"
	"reflect"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
)

func TestRegTestCoinbase(t *testing.T) {
	hex := "02000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0e5101010a2f4542323030302e302fffffffff0100f2052a01000000232103db233bb9fc387d78b133ec904069d46e95ff17da657671b44afa0bc64e89ac18ac00000000"
	bt, err := transaction.NewFromString(hex)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Iscoinbase %t", bt.IsCoinbase())
	t.Logf("input count %d", bt.InputCount())

}

func TestGetVersion(t *testing.T) {
	const tx = "01000000014c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff021d784500000000001976a914e9b62e25d4c6f97287dfe62f8063b79a9638c84688ac60d64f00000000001976a914bb4bca2306df66d72c6e44a470873484d8808b8888ac00000000"
	bt, err := transaction.NewFromString(tx)
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

	r, _ := keys.NewPrivateKey(xprv)

	t.Logf("%x", r.PrivateKey)
}

func TestSignRedeemScript(t *testing.T) {
	var redeemScript, _ = hex.DecodeString("524c53ff0488b21e000000000000000000362f7a9030543db8751401c387d6a71e870f1895b3a62569d455e8ee5f5f5e5f03036624c6df96984db6b4e625b6707c017eb0e0d137cd13a0c989bfa77a4473fd000000004c53ff0488b21e0000000000000000008b20425398995f3c866ea6ce5c1828a516b007379cf97b136bffbdc86f75df14036454bad23b019eae34f10aff8b8d6d8deb18cb31354e5a169ee09d8a4560e8250000000052ae")
	const expected = "3044022041682b268531cf6209577deae34b92fdc83d9ef6e3abc190d4952e927761efd502201696256fba4dd6b05e44ed3871abbd1bc11356aea5ddc36816ca779f68cca6fa"

	const xprv = "xprv9s21ZrQH143K2beTKhLXFRWWFwH8jkwUssjk3SVTiApgmge7kNC3jhVc4NgHW8PhW2y7BCDErqnKpKuyQMjqSePPJooPJowAz5BVLThsv6c"
	const privHex = "5f86e4023a4e94f00463f81b70ff951f83f896a0a3e6ed89cf163c152f954f8b"

	pkBytes, err := hex.DecodeString(privHex)
	if err != nil {
		t.Error(err)
		return
	}
	privKey, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), pkBytes)

	// Sign a message using the private key.
	messageHash := chainhash.DoubleHashB(redeemScript)
	signature, err := privKey.Sign(messageHash)
	if err != nil {
		t.Error(err)
		return
	}

	// Serialize and display the signature.
	t.Logf("Serialized Signature: %x\n", signature.Serialize())

	// Verify the signature for the message using the public key.
	verified := signature.Verify(messageHash, pubKey)
	t.Logf("Signature Verified? %v\n", verified)
}

func TestIsCoinbase(t *testing.T) {
	const coinbase = "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4303bfea07322f53696d6f6e204f726469736820616e642053747561727420467265656d616e206d61646520746869732068617070656e2f9a46434790f7dbdea3430000ffffffff018a08ac4a000000001976a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac00000000"
	bt1, err := transaction.NewFromString(coinbase)
	if err != nil {
		t.Error(err)
		return
	}

	cb1 := bt1.IsCoinbase()
	if cb1 == false {
		t.Errorf("Expecting true, got %t", cb1)
	}

	const tx = "01000000014c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff021d784500000000001976a914e9b62e25d4c6f97287dfe62f8063b79a9638c84688ac60d64f00000000001976a914bb4bca2306df66d72c6e44a470873484d8808b8888ac00000000"
	bt2, err := transaction.NewFromString(tx)
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
	tx, err := transaction.NewFromString(unsignedTx)
	// Previous txid 5b1a7dabe570741c2da71111d396e0be19778d8782c9f28b98d7c72d1b9b417e

	//Add the UTXO amount and script.
	tx.Inputs[0].PreviousTxSatoshis = 2000000000
	tx.Inputs[0].PreviousTxScript = script.NewScriptFromString("76a9148fe80c75c9560e8b56ed64ea3c26e18d2c52211b88ac")
	// t.Logf("%x\n", tx.Hex())
	// tx with input 01000000017e419b1b2dc7d7988bf2c982878d7719bee096d31111a72d1c7470e5ab7d1a5b000000001976a9148fe80c75c9560e8b56ed64ea3c26e18d2c52211b88acffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde47e976000000001976a9148fe80c75c9560e8b56ed64ea3c26e18d2c52211b88ac00000000

	sigType := uint32(transaction.SighashAll | transaction.SighashForkID)
	sigHashes, err := tx.GetSighashPayload(sigType)
	if err != nil {
		t.Error(err)
	}

	if len(*sigHashes) != 1 {
		t.Errorf("Error expected payload to be 1 item long, got %d", len(*sigHashes))
	}

	expectedPayload := transaction.NewSigningPayload()
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
	tx, err := transaction.NewFromString(unsignedTx)
	if err != nil {
		t.Error(err)
		return
	}

	signingPayload := transaction.SigningPayload{}

	// Append a valid response received from the signing service for this Tx.
	signingItem := transaction.SigningItem{
		PublicKeyHash: "bcd0bdbf5fcde5ed957396752d4bd2e01d368702",
		SigHash:       "80448cea404b51f82d409cbd1fbca66bf43fe1cd45d7660953e39ce3c5d8208d",
		PublicKey:     "02ba6bc6906e4937bcde60dbbabdd994dbd0c23e86d834a856091efe677be378b1",
		Signature:     "3045022100a0a005f339978dd6945e44d524d576189f8f7546f41c4899beaa796facb0c4c40220719de9a73796d604b9ee32d7496234c488705fa73f0bd2ffeadcca57580f4cb3",
	}
	signingPayload = append(signingPayload, &signingItem)

	signingItem2 := transaction.SigningItem{
		PublicKeyHash: "bcd0bdbf5fcde5ed957396752d4bd2e01d368702",
		SigHash:       "c62573ac749d9b202cd7b2e0d36a0f688a680810a70ee840f6de7bab4d615095",
		PublicKey:     "02ba6bc6906e4937bcde60dbbabdd994dbd0c23e86d834a856091efe677be378b1",
		Signature:     "30440220399173272f0f56c06b4eb1ccce970603e305988788ab1468e0948ae340fc5380022067684423502f75c5b6e88ad302cc2a1cf739c824efbd5e83fa9e02d4b2975f64",
	}
	signingPayload = append(signingPayload, &signingItem2)

	err = tx.ApplySignatures(&signingPayload, 0)
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
	tx, err := transaction.NewFromString(unsignedTx)

	//Add the UTXO amount and script.
	tx.Inputs[0].PreviousTxSatoshis = 100000000
	tx.Inputs[0].PreviousTxScript = script.NewScriptFromString("76a914c0a3c167a28cabb9fbb495affa0761e6e74ac60d88ac")

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

func TestTxID(t *testing.T) {
	tx, err := transaction.NewFromString("010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d2507232651000000006b483045022100c1d77036dc6cd1f3fa1214b0688391ab7f7a16cd31ea4e5a1f7a415ef167df820220751aced6d24649fa235132f1e6969e163b9400f80043a72879237dab4a1190ad412103b8b40a84123121d260f5c109bc5a46ec819c2e4002e5ba08638783bfb4e01435ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000")
	if err != nil {
		t.Error(err)
	} else {
		id := tx.GetTxID()
		expected := "19dcf16ecc9286c3734fdae3d45d4fc4eb6b25f841131e06460f4939bba0026e"

		if expected != id {
			t.Errorf("Bad TXID")
		}
	}
}

// this test was used to try signing a hash puzzle transaction that
// needed to append the pre-image of the hash to the sigScript
func TestSignTxForced(t *testing.T) {
	unsignedTx := "0100000001f59f8ee5745b020dd3e3a561a539defb626117befc554e168c3bfb88b56ab0f20000000000ffffffff01d0200000000000001976a91447862fe165e6121af80d5dde1ecb478ed170565b88ac00000000"
	tx, err := transaction.NewFromString(unsignedTx)

	//Add the UTXO amount and script.
	tx.Inputs[0].PreviousTxSatoshis = 8519
	tx.Inputs[0].PreviousTxScript = script.NewScriptFromString("a914d3f9e3d971764be5838307b175ee4e08ba427b908876a914c28f832c3d539933e0c719297340b34eee0f4c3488ac")

	// Our private key.
	wif, err := btcutil.DecodeWIF("L31FJtAimeRhprhFEuXpnw1E1sKKuKVgPNUaQ7MjpW3dCWEVuV6R")
	if err != nil {
		t.Error(err)
		return
	}
	err = tx.SignWithoutP2PKHCheck(wif.PrivKey, 0)
	if err != nil {
		t.Error(err)
		return
	}

	secret := "secret1"
	tx.GetInputs()[0].SigScript.AppendPushDataStringToScript(secret)

	expectedSignedTx := "0100000001f59f8ee5745b020dd3e3a561a539defb626117befc554e168c3bfb88b56ab0f20000000073483045022100b30ce9d7e143c3d48a9202b82cf8a32cbe1ee1d9c2a36976bf78a65e71c2255b02203b6152deb3c041179856cc85874a599f2ac41fdbefff28745cafb551630762f9412102adbf278425824e49c1b9f09679451f8754b609544ff72512190ed21881d1ca510773656372657431ffffffff01d0200000000000001976a91447862fe165e6121af80d5dde1ecb478ed170565b88ac00000000"

	if hex.EncodeToString(tx.Hex()) != expectedSignedTx {
		t.Errorf("Expecting %s\n, got %s\n", expectedSignedTx, hex.EncodeToString(tx.Hex()))
	}

	if unsignedTx == expectedSignedTx {
		t.Errorf("Expected and signed TX strings in code identical")
	}
}

func TestValidSignature(t *testing.T) {
	txHex := "02000000011dd7ad77d93879f00dcfeee50ef258775ab13fe0bcfb8f51994ec6f2d295be45000000006a47304402204dbf87fe0bbf435170eea32ed9fa573cf41214b9a7146ca4101eed5738d03e3b02204d86617d7c2bba34874e4a00d3471ff5846d504ece7c67ae0623e2ca516fd0fd412103f4563d1b75b914dfba48fec433b35f56307504ec9fdaa568725619bbae26adf8ffffffff0298ad5a16000000001976a91442f9682260509ac80722b1963aec8a896593d16688ac4de86189030000001976a914c36538e91213a8100dcb2aed456ade363de8483f88ac00000000"
	tx, err := transaction.NewFromString(txHex)
	if err != nil {
		t.Error(err)
		return
	}

	// txid := tx.GetTxID()
	// fmt.Println(txid)

	sigScript := tx.GetInputs()[0].SigScript

	publicKeyBytes := []byte(*sigScript)[len(*sigScript)-33:]
	sigBytes := []byte(*sigScript)[1 : len(*sigScript)-35]
	sigHashType, _ := binary.Uvarint([]byte(*sigScript)[len(*sigScript)-35 : len(*sigScript)-34])

	publicKey, err := btcec.ParsePubKey(publicKeyBytes, btcec.S256())
	if err != nil {
		t.Error(err)
		return
	}
	sig, err := btcec.ParseDERSignature(sigBytes, btcec.S256())
	if err != nil {
		t.Error(err)
		return
	}

	var previousTxSatoshis uint64 = 15564838601
	var previousTxScript *script.Script = script.NewScriptFromString("76a914c7c6987b6e2345a6b138e3384141520a0fbc18c588ac")
	var prevIndex uint32 = 0
	var outIndex uint32 = 0

	sighash := transaction.GetSighashForInputValidation(tx, uint32(sigHashType), outIndex, prevIndex, previousTxSatoshis, previousTxScript)

	h, err := hex.DecodeString(sighash)
	if err != nil {
		t.Error(err)
		return
	}
	valid := sig.Verify(utils.ReverseBytes(h), publicKey)
	t.Logf("%v\n", valid)

}

func TestValidSignature2(t *testing.T) {
	txHex := "0200000001483116c62abe84c0431f6701d1c543b08c50ed7d8cfad882afadcbe3a2eafa64010000006a4730440220665740bdf8cf402f0a3cfeb9a7b82645132190e3c3bd605e0811b79c9dd675e002207929a958673cebe60a6af9fa1fa89e7f3fc397727df5798500d58906c3886a44412103401136395f6c679c6176cdf499ff54720acfb56c07028feaafdce68d79463a45feffffff0200562183000000001976a9140108b364bbbddb222e2d0fac1ad4f6f86b10317688ac9697e4a6000000001976a9143ac52294c730e7a4e9671abe3e7093d8834126ed88ac6f640800"
	tx, err := transaction.NewFromString(txHex)
	if err != nil {
		t.Error(err)
		return
	}

	// txid := tx.GetTxID()
	// fmt.Println(txid)

	sigScript := tx.GetInputs()[0].SigScript

	publicKeyBytes := []byte(*sigScript)[len(*sigScript)-33:]
	sigBytes := []byte(*sigScript)[1 : len(*sigScript)-35]
	sigHashType, _ := binary.Uvarint([]byte(*sigScript)[len(*sigScript)-35 : len(*sigScript)-34])

	publicKey, err := btcec.ParsePubKey(publicKeyBytes, btcec.S256())
	if err != nil {
		t.Error(err)
		return
	}
	sig, err := btcec.ParseDERSignature(sigBytes, btcec.S256())
	if err != nil {
		t.Error(err)
		return
	}

	var previousTxSatoshis uint64 = 5000000000
	var previousTxScript *script.Script = script.NewScriptFromString("76a914343cadc47d08a14ef773d70b3b2a90870b67b3ad88ac")
	var prevIndex uint32 = 1
	var outIndex uint32 = 0

	sighash := transaction.GetSighashForInputValidation(tx, uint32(sigHashType), outIndex, prevIndex, previousTxSatoshis, previousTxScript)

	h, err := hex.DecodeString(sighash)
	if err != nil {
		t.Error(err)
		return
	}
	valid := sig.Verify(utils.ReverseBytes(h), publicKey)
	t.Logf("%v\n", valid)

}

func TestBareMultiSigValidation(t *testing.T) {
	txHex := "0100000001cfb38c76cadeb5b96c3863d9e298fe96e24e594b75f69c37aa709f45b76d1b25000000009200483045022100d83dc84d3ea3fb36b006f6887e1e16811c59fe9a9b79b84142874a90d5b834160220052967be98c26270de0082b0fecab5a40d5bc48d5034b6cdfc2b8e47210e1469414730440220099ffa89363f9a05f23a4fa318ddbefeeeec4b41f6abde7083a3be6696ed904902201722110a488df3780a260ba09b7de6363bfce7f6beec9819e9b9f47f6e978d8141ffffffff01a8840100000000001976a91432b996f742e774b0241be9007f831558ba06d20b88ac00000000"
	tx, err := transaction.NewFromString(txHex)
	if err != nil {
		t.Error(err)
		return
	}

	// txid := tx.GetTxID()
	// fmt.Println(txid)

	var sigs = make([]*btcec.Signature, 2)
	var sigHashTypes = make([]uint32, 2)
	var publicKeys = make([]*btcec.PublicKey, 3)

	sigScript := tx.GetInputs()[0].SigScript

	sig0Bytes := []byte(*sigScript)[2:73]
	sig0HashType, _ := binary.Uvarint([]byte(*sigScript)[73:74])
	sig1Bytes := []byte(*sigScript)[75:145]
	sig1HashType, _ := binary.Uvarint([]byte(*sigScript)[145:146])

	pk0, _ := hex.DecodeString("023ff15e2676e03b2c0af30fc17b7fb354bbfa9f549812da945194d3407dc0969b")
	pk1, _ := hex.DecodeString("039281958c651c013f5b3b007c78be231eeb37f130b925ceff63dc3ac8886f22a3")
	pk2, _ := hex.DecodeString("03ac76121ffc9db556b0ce1da978021bd6cb4a5f9553c14f785e15f0e202139e3e")

	publicKeys[0], err = btcec.ParsePubKey(pk0, btcec.S256())
	if err != nil {
		t.Error(err)
		return
	}
	publicKeys[1], err = btcec.ParsePubKey(pk1, btcec.S256())
	if err != nil {
		t.Error(err)
		return
	}
	publicKeys[2], err = btcec.ParsePubKey(pk2, btcec.S256())
	if err != nil {
		t.Error(err)
		return
	}

	sigs[0], err = btcec.ParseDERSignature(sig0Bytes, btcec.S256())
	if err != nil {
		t.Error(err)
		return
	}
	sigs[1], err = btcec.ParseDERSignature(sig1Bytes, btcec.S256())
	if err != nil {
		t.Error(err)
		return
	}
	sigHashTypes[0] = uint32(sig0HashType)
	sigHashTypes[1] = uint32(sig1HashType)

	var previousTxSatoshis uint64 = 99728
	var previousTxScript *script.Script = script.NewScriptFromString("5221023ff15e2676e03b2c0af30fc17b7fb354bbfa9f549812da945194d3407dc0969b21039281958c651c013f5b3b007c78be231eeb37f130b925ceff63dc3ac8886f22a32103ac76121ffc9db556b0ce1da978021bd6cb4a5f9553c14f785e15f0e202139e3e53ae")
	var prevIndex uint32 = 0
	var outIndex uint32 = 0

	for i, sig := range sigs {
		sighash := transaction.GetSighashForInputValidation(tx, sigHashTypes[i], outIndex, prevIndex, previousTxSatoshis, previousTxScript)
		h, err := hex.DecodeString(sighash)
		if err != nil {
			t.Error(err)
			return
		}
		for j, pk := range publicKeys {
			valid := sig.Verify(utils.ReverseBytes(h), pk)
			t.Logf("signature %d against pulbic key %d => %v\n", i, j, valid)
		}

	}

}

func TestP2SHMultiSigValidation(t *testing.T) { // NOT working properly!
	txHex := "0100000001d0219010e1f74ec8dd264a63ef01b5c72aab49a74c9bffd464c7f7f2b193b34700000000fdfd0000483045022100c2ffae14c7cfae5c1b45776f4b2d497b0d10a9e3be55b1386c555f90acd022af022025d5d1d33429fabd60c41763f9cda5c4b64adbddbd90023febc005be431b97b641473044022013f65e41abd6be856e7c7dd7527edc65231e027c42e8db7358759fc9ccd77b7d02206e024137ee54d2fac9f1dce858a85cb03fb7ba93b8e015d82e8a959b631f91ac414c695221021db57ae3de17143cb6c314fb206b56956e8ed45e2f1cbad3947411228b8d17f1210308b00cf7dfbb64604475e8b18e8450ac6ec04655cfa5c6d4d8a0f3f141ee419421030c7f9342ff6583599db8ee8b52383cadb4cf6fee3650c1ad8f66158a4ff0ebd953aefeffffff01b70f0000000000001976a91415067448220971206e6b4d90733d70fe9610631688ac56750900"
	tx, err := transaction.NewFromString(txHex)
	if err != nil {
		t.Error(err)
		return
	}

	// txid := tx.GetTxID()
	// fmt.Println(txid)

	var sigs = make([]*btcec.Signature, 2)
	var sigHashTypes = make([]uint32, 2)
	var publicKeys = make([]*btcec.PublicKey, 3)

	sigScript := tx.GetInputs()[0].SigScript

	sig0Bytes := []byte(*sigScript)[2:73]
	sig0HashType, _ := binary.Uvarint([]byte(*sigScript)[73:74])
	sig1Bytes := []byte(*sigScript)[75:145]
	sig1HashType, _ := binary.Uvarint([]byte(*sigScript)[145:146])

	pk0, _ := hex.DecodeString("021db57ae3de17143cb6c314fb206b56956e8ed45e2f1cbad3947411228b8d17f1")
	pk1, _ := hex.DecodeString("0308b00cf7dfbb64604475e8b18e8450ac6ec04655cfa5c6d4d8a0f3f141ee4194")
	pk2, _ := hex.DecodeString("030c7f9342ff6583599db8ee8b52383cadb4cf6fee3650c1ad8f66158a4ff0ebd9")

	publicKeys[0], err = btcec.ParsePubKey(pk0, btcec.S256())
	if err != nil {
		t.Error(err)
		return
	}
	publicKeys[1], err = btcec.ParsePubKey(pk1, btcec.S256())
	if err != nil {
		t.Error(err)
		return
	}
	publicKeys[2], err = btcec.ParsePubKey(pk2, btcec.S256())
	if err != nil {
		t.Error(err)
		return
	}

	sigs[0], err = btcec.ParseDERSignature(sig0Bytes, btcec.S256())
	if err != nil {
		t.Error(err)
		return
	}
	sigs[1], err = btcec.ParseDERSignature(sig1Bytes, btcec.S256())
	if err != nil {
		t.Error(err)
		return
	}
	sigHashTypes[0] = uint32(sig0HashType)
	sigHashTypes[1] = uint32(sig1HashType)

	var previousTxSatoshis uint64 = 8785040
	var previousTxScript *script.Script = script.NewScriptFromString("5221021db57ae3de17143cb6c314fb206b56956e8ed45e2f1cbad3947411228b8d17f1210308b00cf7dfbb64604475e8b18e8450ac6ec04655cfa5c6d4d8a0f3f141ee419421030c7f9342ff6583599db8ee8b52383cadb4cf6fee3650c1ad8f66158a4ff0ebd953ae")
	var prevIndex uint32 = 1
	var outIndex uint32 = 0

	for i, sig := range sigs {
		sighash := transaction.GetSighashForInputValidation(tx, sigHashTypes[i], outIndex, prevIndex, previousTxSatoshis, previousTxScript)
		h, err := hex.DecodeString(sighash)
		if err != nil {
			t.Error(err)
			return
		}
		for j, pk := range publicKeys {
			valid := sig.Verify(utils.ReverseBytes(h), pk)
			t.Logf("signature %d against pulbic key %d => %v\n", i, j, valid)
		}

	}

}

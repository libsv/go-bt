package sig_test

import (
	"encoding/hex"
	"testing"

	"github.com/bitcoinsv/bsvutil"
	"github.com/libsv/libsv/bt"
	"github.com/libsv/libsv/bt/sig"
	"github.com/libsv/libsv/script"
)

func TestSignAuto(t *testing.T) {
	unsignedTx := "010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d25072326510000000000ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000"
	tx, err := bt.NewFromString(unsignedTx)

	if err != nil {
		t.Fatal("Failed to create transaction")
	}
	// Add the UTXO amount and script.
	tx.Inputs[0].PreviousTxSatoshis = 100000000
	tx.Inputs[0].PreviousTxScript, _ = script.NewFromHexString("76a914c0a3c167a28cabb9fbb495affa0761e6e74ac60d88ac")

	// Our private key.
	wif, err := bsvutil.DecodeWIF("cNGwGSc7KRrTmdLUZ54fiSXWbhLNDc2Eg5zNucgQxyQCzuQ5YRDq")
	if err != nil {
		t.Fatal(err)
	}
	signer := sig.InternalSigner{PrivateKey: wif.PrivKey, SigHashFlag: 0}
	err = tx.SignAuto(&signer)
	if err != nil {
		t.Fatal(err)
	}
	expectedSignedTx := "010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d2507232651000000006b483045022100c1d77036dc6cd1f3fa1214b0688391ab7f7a16cd31ea4e5a1f7a415ef167df820220751aced6d24649fa235132f1e6969e163b9400f80043a72879237dab4a1190ad412103b8b40a84123121d260f5c109bc5a46ec819c2e4002e5ba08638783bfb4e01435ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000"

	if hex.EncodeToString(tx.ToBytes()) != expectedSignedTx {
		t.Errorf("Expecting %s\n, got %s\n", expectedSignedTx, hex.EncodeToString(tx.ToBytes()))
	}

	// TODO: what is this for?
	//if unsignedTx == expectedSignedTx {
	//	t.Errorf("Expected and signed TX strings in code identical")
	//}
}

// TODO: fix/ update to use internal signer
// func TestValidSignature(t *testing.T) {
// 	txHex := "02000000011dd7ad77d93879f00dcfeee50ef258775ab13fe0bcfb8f51994ec6f2d295be45000000006a47304402204dbf87fe0bbf435170eea32ed9fa573cf41214b9a7146ca4101eed5738d03e3b02204d86617d7c2bba34874e4a00d3471ff5846d504ece7c67ae0623e2ca516fd0fd412103f4563d1b75b914dfba48fec433b35f56307504ec9fdaa568725619bbae26adf8ffffffff0298ad5a16000000001976a91442f9682260509ac80722b1963aec8a896593d16688ac4de86189030000001976a914c36538e91213a8100dcb2aed456ade363de8483f88ac00000000"
// 	tx, err := transaction.NewFromString(txHex)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
//
// 	// txid := tx.GetTxID()
// 	// fmt.Println(txid)
//
// 	sigScript := tx.GetInputs()[0].UnlockingScript
//
// 	publicKeyBytes := []byte(*sigScript)[len(*sigScript)-33:]
// 	sigBytes := []byte(*sigScript)[1 : len(*sigScript)-35]
// 	sigHashType, _ := binary.Uvarint([]byte(*sigScript)[len(*sigScript)-35 : len(*sigScript)-34])
//
// 	publicKey, err := bsvec.ParsePubKey(publicKeyBytes, bsvec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	sig, err := bsvec.ParseDERSignature(sigBytes, bsvec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
//
// 	var previousTxSatoshis uint64 = 15564838601
//
// 	var previousTxScript, _ = script.NewFromHexString("76a914c7c6987b6e2345a6b138e3384141520a0fbc18c588ac")
// 	var prevIndex, outIndex uint32
//
// 	sighash := signature.GetSighashForInputValidation(tx, uint32(sigHashType), outIndex, prevIndex, previousTxSatoshis, previousTxScript)
//
// 	h, err := hex.DecodeString(sighash)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	valid := sig.Verify(utils.ReverseBytes(h), publicKey)
// 	t.Logf("%v\n", valid)
//
// }
//
// func TestValidSignature2(t *testing.T) {
// 	txHex := "0200000001483116c62abe84c0431f6701d1c543b08c50ed7d8cfad882afadcbe3a2eafa64010000006a4730440220665740bdf8cf402f0a3cfeb9a7b82645132190e3c3bd605e0811b79c9dd675e002207929a958673cebe60a6af9fa1fa89e7f3fc397727df5798500d58906c3886a44412103401136395f6c679c6176cdf499ff54720acfb56c07028feaafdce68d79463a45feffffff0200562183000000001976a9140108b364bbbddb222e2d0fac1ad4f6f86b10317688ac9697e4a6000000001976a9143ac52294c730e7a4e9671abe3e7093d8834126ed88ac6f640800"
// 	tx, err := transaction.NewFromString(txHex)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
//
// 	// txid := tx.GetTxID()
// 	// fmt.Println(txid)
//
// 	sigScript := tx.GetInputs()[0].UnlockingScript
//
// 	publicKeyBytes := []byte(*sigScript)[len(*sigScript)-33:]
// 	sigBytes := []byte(*sigScript)[1 : len(*sigScript)-35]
// 	sigHashType, _ := binary.Uvarint([]byte(*sigScript)[len(*sigScript)-35 : len(*sigScript)-34])
//
// 	publicKey, err := bsvec.ParsePubKey(publicKeyBytes, bsvec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	sig, err := bsvec.ParseDERSignature(sigBytes, bsvec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
//
// 	var previousTxSatoshis uint64 = 5000000000
//
// 	var previousTxScript, _ = script.NewFromHexString("76a914343cadc47d08a14ef773d70b3b2a90870b67b3ad88ac")
// 	var prevIndex uint32 = 1
// 	var outIndex uint32
//
// 	sighash := signature.GetSighashForInputValidation(tx, uint32(sigHashType), outIndex, prevIndex, previousTxSatoshis, previousTxScript)
//
// 	h, err := hex.DecodeString(sighash)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	valid := sig.Verify(utils.ReverseBytes(h), publicKey)
// 	t.Logf("%v\n", valid)
//
// }
//
// func TestBareMultiSigValidation(t *testing.T) {
// 	txHex := "0100000001cfb38c76cadeb5b96c3863d9e298fe96e24e594b75f69c37aa709f45b76d1b25000000009200483045022100d83dc84d3ea3fb36b006f6887e1e16811c59fe9a9b79b84142874a90d5b834160220052967be98c26270de0082b0fecab5a40d5bc48d5034b6cdfc2b8e47210e1469414730440220099ffa89363f9a05f23a4fa318ddbefeeeec4b41f6abde7083a3be6696ed904902201722110a488df3780a260ba09b7de6363bfce7f6beec9819e9b9f47f6e978d8141ffffffff01a8840100000000001976a91432b996f742e774b0241be9007f831558ba06d20b88ac00000000"
// 	tx, err := transaction.NewFromString(txHex)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
//
// 	// txid := tx.GetTxID()
// 	// fmt.Println(txid)
//
// 	var sigs = make([]*bsvec.Signature, 2)
// 	var sigHashTypes = make([]uint32, 2)
// 	var publicKeys = make([]*bsvec.PublicKey, 3)
//
// 	sigScript := tx.GetInputs()[0].UnlockingScript
//
// 	sig0Bytes := []byte(*sigScript)[2:73]
// 	sig0HashType, _ := binary.Uvarint([]byte(*sigScript)[73:74])
// 	sig1Bytes := []byte(*sigScript)[75:145]
// 	sig1HashType, _ := binary.Uvarint([]byte(*sigScript)[145:146])
//
// 	pk0, _ := hex.DecodeString("023ff15e2676e03b2c0af30fc17b7fb354bbfa9f549812da945194d3407dc0969b")
// 	pk1, _ := hex.DecodeString("039281958c651c013f5b3b007c78be231eeb37f130b925ceff63dc3ac8886f22a3")
// 	pk2, _ := hex.DecodeString("03ac76121ffc9db556b0ce1da978021bd6cb4a5f9553c14f785e15f0e202139e3e")
//
// 	publicKeys[0], err = bsvec.ParsePubKey(pk0, bsvec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	publicKeys[1], err = bsvec.ParsePubKey(pk1, bsvec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	publicKeys[2], err = bsvec.ParsePubKey(pk2, bsvec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
//
// 	sigs[0], err = bsvec.ParseDERSignature(sig0Bytes, bsvec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	sigs[1], err = bsvec.ParseDERSignature(sig1Bytes, bsvec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	sigHashTypes[0] = uint32(sig0HashType)
// 	sigHashTypes[1] = uint32(sig1HashType)
//
// 	var previousTxSatoshis uint64 = 99728
// 	var previousTxScript, _ = script.NewFromHexString("5221023ff15e2676e03b2c0af30fc17b7fb354bbfa9f549812da945194d3407dc0969b21039281958c651c013f5b3b007c78be231eeb37f130b925ceff63dc3ac8886f22a32103ac76121ffc9db556b0ce1da978021bd6cb4a5f9553c14f785e15f0e202139e3e53ae")
// 	var prevIndex uint32
// 	var outIndex uint32
//
// 	for i, sig := range sigs {
// 		sighash := signature.GetSighashForInputValidation(tx, sigHashTypes[i], outIndex, prevIndex, previousTxSatoshis, previousTxScript)
// 		h, err := hex.DecodeString(sighash)
// 		if err != nil {
// 			t.Error(err)
// 			return
// 		}
// 		for j, pk := range publicKeys {
// 			valid := sig.Verify(utils.ReverseBytes(h), pk)
// 			t.Logf("signature %d against pulbic key %d => %v\n", i, j, valid)
// 		}
//
// 	}
//
// }
//
// func TestP2SHMultiSigValidation(t *testing.T) { // NOT working properly!
// 	txHex := "0100000001d0219010e1f74ec8dd264a63ef01b5c72aab49a74c9bffd464c7f7f2b193b34700000000fdfd0000483045022100c2ffae14c7cfae5c1b45776f4b2d497b0d10a9e3be55b1386c555f90acd022af022025d5d1d33429fabd60c41763f9cda5c4b64adbddbd90023febc005be431b97b641473044022013f65e41abd6be856e7c7dd7527edc65231e027c42e8db7358759fc9ccd77b7d02206e024137ee54d2fac9f1dce858a85cb03fb7ba93b8e015d82e8a959b631f91ac414c695221021db57ae3de17143cb6c314fb206b56956e8ed45e2f1cbad3947411228b8d17f1210308b00cf7dfbb64604475e8b18e8450ac6ec04655cfa5c6d4d8a0f3f141ee419421030c7f9342ff6583599db8ee8b52383cadb4cf6fee3650c1ad8f66158a4ff0ebd953aefeffffff01b70f0000000000001976a91415067448220971206e6b4d90733d70fe9610631688ac56750900"
// 	tx, err := transaction.NewFromString(txHex)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
//
// 	// txid := tx.GetTxID()
// 	// fmt.Println(txid)
//
// 	var sigs = make([]*bsvec.Signature, 2)
// 	var sigHashTypes = make([]uint32, 2)
// 	var publicKeys = make([]*bsvec.PublicKey, 3)
//
// 	sigScript := tx.GetInputs()[0].UnlockingScript
//
// 	sig0Bytes := []byte(*sigScript)[2:73]
// 	sig0HashType, _ := binary.Uvarint([]byte(*sigScript)[73:74])
// 	sig1Bytes := []byte(*sigScript)[75:145]
// 	sig1HashType, _ := binary.Uvarint([]byte(*sigScript)[145:146])
//
// 	pk0, _ := hex.DecodeString("021db57ae3de17143cb6c314fb206b56956e8ed45e2f1cbad3947411228b8d17f1")
// 	pk1, _ := hex.DecodeString("0308b00cf7dfbb64604475e8b18e8450ac6ec04655cfa5c6d4d8a0f3f141ee4194")
// 	pk2, _ := hex.DecodeString("030c7f9342ff6583599db8ee8b52383cadb4cf6fee3650c1ad8f66158a4ff0ebd9")
//
// 	publicKeys[0], err = bsvec.ParsePubKey(pk0, bsvec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	publicKeys[1], err = bsvec.ParsePubKey(pk1, bsvec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	publicKeys[2], err = bsvec.ParsePubKey(pk2, bsvec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
//
// 	sigs[0], err = bsvec.ParseDERSignature(sig0Bytes, bsvec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	sigs[1], err = bsvec.ParseDERSignature(sig1Bytes, bsvec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	sigHashTypes[0] = uint32(sig0HashType)
// 	sigHashTypes[1] = uint32(sig1HashType)
//
// 	var previousTxSatoshis uint64 = 8785040
// 	var previousTxScript, _ = script.NewFromHexString("5221021db57ae3de17143cb6c314fb206b56956e8ed45e2f1cbad3947411228b8d17f1210308b00cf7dfbb64604475e8b18e8450ac6ec04655cfa5c6d4d8a0f3f141ee419421030c7f9342ff6583599db8ee8b52383cadb4cf6fee3650c1ad8f66158a4ff0ebd953ae")
// 	var prevIndex uint32 = 1
// 	var outIndex uint32 = 0
//
// 	for i, sig := range sigs {
// 		sighash := signature.GetSighashForInputValidation(tx, sigHashTypes[i], outIndex, prevIndex, previousTxSatoshis, previousTxScript)
// 		h, err := hex.DecodeString(sighash)
// 		if err != nil {
// 			t.Error(err)
// 			return
// 		}
// 		for j, pk := range publicKeys {
// 			valid := sig.Verify(utils.ReverseBytes(h), pk)
// 			t.Logf("signature %d against pulbic key %d => %v\n", i, j, valid)
// 		}
//
// 	}
//
// }

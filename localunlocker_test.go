package bt_test

import (
	"context"
	"testing"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/sighash"
	"github.com/stretchr/testify/assert"
)

func TestLocalSignatureUnlocker_UnlockAll(t *testing.T) {
	t.Parallel()

	incompleteTx := "010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d25072326510000000000ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000"
	tx, err := bt.NewTxFromString(incompleteTx)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Add the UTXO amount and script.
	tx.InputIdx(0).PreviousTxSatoshis = 100000000
	tx.InputIdx(0).PreviousTxScript, err = bscript.NewFromHexString("76a914c0a3c167a28cabb9fbb495affa0761e6e74ac60d88ac")
	assert.NoError(t, err)

	// Our private key
	var w *wif.WIF
	w, err = wif.DecodeWIF("cNGwGSc7KRrTmdLUZ54fiSXWbhLNDc2Eg5zNucgQxyQCzuQ5YRDq")
	assert.NoError(t, err)

	unlocker := bt.LocalSignatureUnlockerGetter{PrivateKey: w.PrivKey}
	err = tx.UnlockAll(context.Background(), &unlocker)
	assert.NoError(t, err)

	expectedSignedTx := "010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d2507232651000000006b483045022100c1d77036dc6cd1f3fa1214b0688391ab7f7a16cd31ea4e5a1f7a415ef167df820220751aced6d24649fa235132f1e6969e163b9400f80043a72879237dab4a1190ad412103b8b40a84123121d260f5c109bc5a46ec819c2e4002e5ba08638783bfb4e01435ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000"
	assert.Equal(t, expectedSignedTx, tx.String())
	assert.NotEqual(t, incompleteTx, tx.String())
}

func TestLocalSignatureUnlocker_ValidSignature(t *testing.T) {
	tests := map[string]struct {
		tx *bt.Tx
	}{
		"valid signature 1": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From("45be95d2f2c64e99518ffbbce03fb15a7758f20ee5eecf0df07938d977add71d", 0, "76a914c7c6987b6e2345a6b138e3384141520a0fbc18c588ac", 15564838601))

				script1, err := bscript.NewFromHexString("76a91442f9682260509ac80722b1963aec8a896593d16688ac")
				assert.NoError(t, err)

				assert.NoError(t, tx.AddP2PKHOutputFromScript(script1, 375041432))

				script2, err := bscript.NewFromHexString("76a914c36538e91213a8100dcb2aed456ade363de8483f88ac")
				assert.NoError(t, err)

				assert.NoError(t, tx.AddP2PKHOutputFromScript(script2, 15189796941))

				return tx
			}(),
		},
		"valid signature 2": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()

				assert.NoError(
					t,
					tx.From("64faeaa2e3cbadaf82d8fa8c7ded508cb043c5d101671f43c084be2ac6163148", 1, "76a914343cadc47d08a14ef773d70b3b2a90870b67b3ad88ac", 5000000000),
				)
				tx.Inputs[0].SequenceNumber = 0xfffffffe

				script1, err := bscript.NewFromHexString("76a9140108b364bbbddb222e2d0fac1ad4f6f86b10317688ac")
				assert.NoError(t, err)

				assert.NoError(t, tx.AddP2PKHOutputFromScript(script1, 2200000000))

				script2, err := bscript.NewFromHexString("76a9143ac52294c730e7a4e9671abe3e7093d8834126ed88ac")
				assert.NoError(t, err)

				assert.NoError(t, tx.AddP2PKHOutputFromScript(script2, 2799998870))
				return tx
			}(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tx := test.tx

			var w *wif.WIF
			w, err := wif.DecodeWIF("cNGwGSc7KRrTmdLUZ54fiSXWbhLNDc2Eg5zNucgQxyQCzuQ5YRDq")
			assert.NoError(t, err)

			unlocker := &bt.LocalSignatureUnlocker{PrivateKey: w.PrivKey}
			parts, err := unlocker.Unlock(context.Background(), tx, 0, sighash.AllForkID)
			assert.NoError(t, err)

			//unlockingScript := []byte(*tx.Inputs[0].UnlockingScript)

			publicKeyBytes := parts[0]
			sigBytes := parts[1]
			//publicKeyBytes := unlockingScript[len(unlockingScript)-33:]
			//sigBytes := unlockingScript[1 : len(unlockingScript)-35]

			publicKey, err := bec.ParsePubKey(publicKeyBytes, bec.S256())
			assert.NoError(t, err)

			sig, err := bec.ParseDERSignature(sigBytes, bec.S256())
			assert.NoError(t, err)

			sh, err := tx.CalcInputSignatureHash(0, sighash.AllForkID)
			assert.NoError(t, err)

			assert.True(t, sig.Verify(sh, publicKey))
		})
	}
}

type mockUnlockerGetter struct {
	t            *testing.T
	unlockerFunc func(ctx context.Context, lockingScript *bscript.Script) (bt.Unlocker, error)
}

func (m *mockUnlockerGetter) Unlocker(ctx context.Context, lockingScript *bscript.Script) (bt.Unlocker, error) {
	assert.NotNil(m.t, m.unlockerFunc, "unlockerFunc not set in this test")
	return m.unlockerFunc(ctx, lockingScript)
}

type mockUnlocker struct {
	t      *testing.T
	script string
}

func (m *mockUnlocker) Unlock(ctx context.Context, tx *bt.Tx, idx uint32, shf sighash.Flag) ([][]byte, error) {
	script, err := bscript.NewFromASM(m.script)
	assert.NoError(m.t, err)

	return [][]byte{*script}, tx.ApplyUnlockingScript(idx, script)
}

func TestLocalSignatureUnlocker_NonSignature(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		tx                  *bt.Tx
		unlockerFunc        func(ctx context.Context, lockingScript *bscript.Script) (bt.Unlocker, error)
		expUnlockingScripts []string
	}{
		"simple script": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From("45be95d2f2c64e99518ffbbce03fb15a7758f20ee5eecf0df07938d977add71d", 0, "52529387", 15564838601))
				return tx
			}(),
			unlockerFunc: func(ctx context.Context, lockingScript *bscript.Script) (bt.Unlocker, error) {
				asm, err := lockingScript.ToASM()
				assert.NoError(t, err)

				unlocker, ok := map[string]*mockUnlocker{
					"OP_2 OP_2 OP_ADD OP_EQUAL": {t: t, script: "OP_4"},
				}[asm]

				assert.True(t, ok)
				assert.NotNil(t, unlocker)

				return unlocker, nil
			},
			expUnlockingScripts: []string{"OP_4"},
		},
		"multiple inputs unlocked": {
			tx: func() *bt.Tx {
				tx := bt.NewTx()
				assert.NoError(t, tx.From("45be95d2f2c64e99518ffbbce03fb15a7758f20ee5eecf0df07938d977add71d", 0, "52529487", 15564838601))
				assert.NoError(t, tx.From("45be95d2f2c64e99518ffbbce03fb15a7758f20ee5eecf0df07938d977add71d", 0, "52589587", 15564838601))
				assert.NoError(t, tx.From("45be95d2f2c64e99518ffbbce03fb15a7758f20ee5eecf0df07938d977add71d", 0, "5a559687", 15564838601))
				return tx
			}(),
			unlockerFunc: func(ctx context.Context, lockingScript *bscript.Script) (bt.Unlocker, error) {
				asm, err := lockingScript.ToASM()
				assert.NoError(t, err)

				unlocker, ok := map[string]*mockUnlocker{
					"OP_2 OP_2 OP_SUB OP_EQUAL":  {t: t, script: "OP_FALSE"},
					"OP_2 OP_8 OP_MUL OP_EQUAL":  {t: t, script: "OP_16"},
					"OP_10 OP_5 OP_DIV OP_EQUAL": {t: t, script: "OP_2"},
				}[asm]

				assert.True(t, ok)
				assert.NotNil(t, unlocker)

				return unlocker, nil
			},
			expUnlockingScripts: []string{"OP_FALSE", "OP_16", "OP_2"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tx := test.tx
			assert.Equal(t, len(tx.Inputs), len(test.expUnlockingScripts))

			ug := &mockUnlockerGetter{
				t:            t,
				unlockerFunc: test.unlockerFunc,
			}
			assert.NoError(t, tx.UnlockAll(context.Background(), ug))
			for i, script := range test.expUnlockingScripts {
				asm, err := tx.Inputs[i].UnlockingScript.ToASM()
				assert.NoError(t, err)

				assert.Equal(t, script, asm)
			}
		})
	}
}

//
// func TestBareMultiSigValidation(t *testing.T) {
// 	txHex := "0100000001cfb38c76cadeb5b96c3863d9e298fe96e24e594b75f69c37aa709f45b76d1b25000000009200483045022100d83dc84d3ea3fb36b006f6887e1e16811c59fe9a9b79b84142874a90d5b834160220052967be98c26270de0082b0fecab5a40d5bc48d5034b6cdfc2b8e47210e1469414730440220099ffa89363f9a05f23a4fa318ddbefeeeec4b41f6abde7083a3be6696ed904902201722110a488df3780a260ba09b7de6363bfce7f6beec9819e9b9f47f6e978d8141ffffffff01a8840100000000001976a91432b996f742e774b0241be9007f831558ba06d20b88ac00000000"
// 	tx, err := transaction.NewTxFromString(txHex)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
//
// 	// txid := tx.GetTxID()
// 	// fmt.Println(txid)
//
// 	var sigs = make([]*bec.Signature, 2)
// 	var sigHashTypes = make([]uint32, 2)
// 	var publicKeys = make([]*bec.PublicKey, 3)
//
// 	sigScript := tx.Inputs[0].UnlockingScript
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
// 	publicKeys[0], err = bec.ParsePubKey(pk0, bec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	publicKeys[1], err = bec.ParsePubKey(pk1, bec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	publicKeys[2], err = bec.ParsePubKey(pk2, bec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
//
// 	sigs[0], err = bec.ParseDERSignature(sig0Bytes, bec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	sigs[1], err = bec.ParseDERSignature(sig1Bytes, bec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	sigHashTypes[0] = uint32(sig0HashType)
// 	sigHashTypes[1] = uint32(sig1HashType)
//
// 	var previousTxSatoshis uint64 = 99728
// 	var previousTxScript, _ = bbscript.NewFromHexString("5221023ff15e2676e03b2c0af30fc17b7fb354bbfa9f549812da945194d3407dc0969b21039281958c651c013f5b3b007c78be231eeb37f130b925ceff63dc3ac8886f22a32103ac76121ffc9db556b0ce1da978021bd6cb4a5f9553c14f785e15f0e202139e3e53ae")
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
// 			valid := sig.Verify(ReverseBytes(h), pk)
// 			t.Logf("signature %d against pulbic key %d => %v\n", i, j, valid)
// 		}
//
// 	}
//
// }
//
// func TestP2SHMultiSigValidation(t *testing.T) { // NOT working properly!
// 	txHex := "0100000001d0219010e1f74ec8dd264a63ef01b5c72aab49a74c9bffd464c7f7f2b193b34700000000fdfd0000483045022100c2ffae14c7cfae5c1b45776f4b2d497b0d10a9e3be55b1386c555f90acd022af022025d5d1d33429fabd60c41763f9cda5c4b64adbddbd90023febc005be431b97b641473044022013f65e41abd6be856e7c7dd7527edc65231e027c42e8db7358759fc9ccd77b7d02206e024137ee54d2fac9f1dce858a85cb03fb7ba93b8e015d82e8a959b631f91ac414c695221021db57ae3de17143cb6c314fb206b56956e8ed45e2f1cbad3947411228b8d17f1210308b00cf7dfbb64604475e8b18e8450ac6ec04655cfa5c6d4d8a0f3f141ee419421030c7f9342ff6583599db8ee8b52383cadb4cf6fee3650c1ad8f66158a4ff0ebd953aefeffffff01b70f0000000000001976a91415067448220971206e6b4d90733d70fe9610631688ac56750900"
// 	tx, err := transaction.NewTxFromString(txHex)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
//
// 	// txid := tx.GetTxID()
// 	// fmt.Println(txid)
//
// 	var sigs = make([]*bec.Signature, 2)
// 	var sigHashTypes = make([]uint32, 2)
// 	var publicKeys = make([]*bec.PublicKey, 3)
//
// 	sigScript := tx.Inputs[0].UnlockingScript
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
// 	publicKeys[0], err = bec.ParsePubKey(pk0, bec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	publicKeys[1], err = bec.ParsePubKey(pk1, bec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	publicKeys[2], err = bec.ParsePubKey(pk2, bec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
//
// 	sigs[0], err = bec.ParseDERSignature(sig0Bytes, bec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	sigs[1], err = bec.ParseDERSignature(sig1Bytes, bec.S256())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	sigHashTypes[0] = uint32(sig0HashType)
// 	sigHashTypes[1] = uint32(sig1HashType)
//
// 	var previousTxSatoshis uint64 = 8785040
// 	var previousTxScript, _ = bscript.NewFromHexString("5221021db57ae3de17143cb6c314fb206b56956e8ed45e2f1cbad3947411228b8d17f1210308b00cf7dfbb64604475e8b18e8450ac6ec04655cfa5c6d4d8a0f3f141ee419421030c7f9342ff6583599db8ee8b52383cadb4cf6fee3650c1ad8f66158a4ff0ebd953ae")
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
// 			valid := sig.Verify(ReverseBytes(h), pk)
// 			t.Logf("signature %d against pulbic key %d => %v\n", i, j, valid)
// 		}
//
// 	}
//
// }

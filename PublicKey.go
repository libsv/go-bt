package cryptolib

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

// Networks
var (
	Public      []byte // Public mainnet
	Private     []byte // Private mainnet
	TestPublic  []byte
	TestPrivate []byte
)

func init() {
	Public, _ = hex.DecodeString("0488B21E")
	Private, _ = hex.DecodeString("0488ADE4")
	TestPublic, _ = hex.DecodeString("043587CF")
	TestPrivate, _ = hex.DecodeString("04358394")
}

// A PublicKey contains metadata associated with an ECDSA public key.
type PublicKey struct {
	Version              []byte `json:"-"`
	Network              string `json:"network"`
	Depth                uint16 `json:"depth"`
	FingerPrint          []byte `json:"-"`                 // : -1440106556,
	FingerPrintStr       string `json:"fingerPrint"`       // : -1440106556,
	ParentFingerPrint    []byte `json:"-"`                 //: 0,
	ParentFingerPrintStr string `json:"parentFingerPrint"` //: 0,
	ChildIndex           []byte `json:"childIndex"`        //: 0,
	ChainCode            []byte `json:"-"`                 //   '41fc504936a63056da1a0f9dd44cad3651b64a17b53e523e18a8d228a489c16a',
	ChainCodeStr         string `json:"chainCode"`
	PrivateKey           []byte `json:"-"`          //   '0362e448fdb4c7c307a80cc3c8ede19cd2599a5ea5c05b188fc56a25c59bfcf125',
	PrivateKeyStr        string `json:"privateKey"` //   '0362e448fdb4c7c307a80cc3c8ede19cd2599a5ea5c05b188fc56a25c59bfcf125',
	PublicKey            []byte `json:"-"`          //   '0362e448fdb4c7c307a80cc3c8ede19cd2599a5ea5c05b188fc56a25c59bfcf125',
	PublicKeyStr         string `json:"publicKey"`  //   '0362e448fdb4c7c307a80cc3c8ede19cd2599a5ea5c05b188fc56a25c59bfcf125',
	Checksum             []byte `json:"-"`          //: 43286247,
	XPrvKey              string `json:"xprvkey"`    // 'xprv661My
	XPubKey              string `json:"xpubkey"`    // 'xpub661My
}

var curve = btcec.S256()

// NewPrivateKey comment TODO: public key or private key?
func NewPrivateKey(xprv string) (*PublicKey, error) {
	decoded, err := DecodeString(xprv)
	if err != nil {
		return nil, err
	}

	privateKey := decoded[46:78]

	// The fingerprint is the 1st 4 bytes of the ripemd160 of the sha256 of the public key.
	// hasher := ripemd160.New()
	// sha := sha256.Sum256(publicKey)
	// hasher.Write(sha[:])
	// hashBytes := hasher.Sum(nil)
	// fingerPrint := hashBytes[0:4]

	return &PublicKey{
		Network: "livenet",

		Version: decoded[0:4],
		Depth:   byteToUint16(decoded[4:5]),
		// FingerPrint:          fingerPrint,
		// FingerPrintStr:       hex.EncodeToString(fingerPrint),
		ParentFingerPrint:    decoded[5:9],
		ParentFingerPrintStr: hex.EncodeToString(decoded[5:9]),
		ChildIndex:           decoded[9:13],
		ChainCode:            decoded[13:45],
		ChainCodeStr:         hex.EncodeToString(decoded[13:45]),
		PrivateKey:           privateKey,
		PrivateKeyStr:        hex.EncodeToString(privateKey),
		Checksum:             decoded[78:82],
		XPrvKey:              xprv,
	}, nil
}

// NewPublicKey takes an xpub string and returns a PublicKey pointer.
// See BIP32 https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki
func NewPublicKey(xpub string) (*PublicKey, error) {
	decoded, err := DecodeString(xpub)
	if err != nil {
		return nil, err
	}

	publicKey := decoded[45:78]

	// The fingerprint is the 1st 4 bytes of the ripemd160 of the sha256 of the public key.
	hasher := ripemd160.New()
	sha := sha256.Sum256(publicKey)
	hasher.Write(sha[:])
	hashBytes := hasher.Sum(nil)
	fingerPrint := hashBytes[0:4]

	// fmt.Printf("Version:     %+v\n", decoded[0:4])
	// fmt.Printf("Depth:       %+v\n", decoded[4:5])
	// fmt.Printf("FingerPrint: %+v\n", fingerPrint)
	// fmt.Printf("ChildIndex:  %+v\n", decoded[9:13])
	// fmt.Printf("ChainCode:   %+v\n", decoded[13:45])
	// fmt.Printf("PublicKey:   %+v\n", publicKey)
	// fmt.Printf("Checksum:    %+v\n", decoded[78:82])

	return &PublicKey{
		Network: "livenet",

		Version:              decoded[0:4],
		Depth:                byteToUint16(decoded[4:5]),
		FingerPrint:          fingerPrint,
		FingerPrintStr:       hex.EncodeToString(fingerPrint),
		ParentFingerPrint:    decoded[5:9],
		ParentFingerPrintStr: hex.EncodeToString(decoded[5:9]),
		ChildIndex:           decoded[9:13],
		ChainCode:            decoded[13:45],
		ChainCodeStr:         hex.EncodeToString(decoded[13:45]),
		PublicKey:            publicKey,
		PublicKeyStr:         hex.EncodeToString(publicKey),
		Checksum:             decoded[78:82],
		XPubKey:              xpub,
	}, nil
}

func (pk *PublicKey) String() string {
	b, err := json.Marshal(pk)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// Child dervies a child public key for a specific index.
// See BIP32 https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki
func (pk *PublicKey) Child(i uint32) (*PublicKey, error) {
	mac := hmac.New(sha512.New, pk.ChainCode)
	if i >= uint32(0x80000000) {
		return nil, errors.New("Can't do Private derivation on Public key")
	}

	mac.Write(append(pk.PublicKey, uint32ToByte(i)...))
	I := mac.Sum(nil)

	iL := new(big.Int).SetBytes(I[:32])

	if iL.Cmp(curve.N) >= 0 || iL.Sign() == 0 {
		return nil, errors.New("Invalid Child")
	}
	newKey := addPubKeys(privToPub(I[:32]), pk.PublicKey)
	fingerPrint := hash160(pk.PublicKey)[:4]

	child := &PublicKey{
		Network: "livenet",

		Version:              pk.Version,
		Depth:                pk.Depth + 1,
		FingerPrint:          fingerPrint,
		FingerPrintStr:       hex.EncodeToString(fingerPrint),
		ParentFingerPrint:    pk.FingerPrint,
		ParentFingerPrintStr: pk.FingerPrintStr,
		ChildIndex:           uint32ToByte(i),
		ChainCode:            I[32:],
		ChainCodeStr:         hex.EncodeToString(I[32:]),
		PublicKey:            newKey,
		PublicKeyStr:         hex.EncodeToString(newKey),
	}

	child.XPubKey = child.GetXPub()

	return child, nil
}

// Address returns bitcoin address represented by public key...
func (pk *PublicKey) Address() (string, error) {
	var prefix []byte
	if bytes.Compare(pk.Version, TestPublic) == 0 || bytes.Compare(pk.Version, TestPrivate) == 0 {
		prefix, _ = hex.DecodeString("6F")
	} else {
		prefix, _ = hex.DecodeString("00")
	}
	addr1 := append(prefix, hash160(pk.PublicKey)...)
	chksum := Sha256d(addr1)
	return base58.Encode(append(addr1, chksum[:4]...)), nil
}

//2.3.4 of SEC1 - http://www.secg.org/index.php?action=secg,docs_secg
func expand(key []byte) (*big.Int, *big.Int) {
	params := curve.Params()
	exp := big.NewInt(1)
	exp.Add(params.P, exp)
	exp.Div(exp, big.NewInt(4))
	x := big.NewInt(0).SetBytes(key[1:33])
	y := big.NewInt(0).SetBytes(key[:1])
	beta := big.NewInt(0)
	beta.Exp(x, big.NewInt(3), nil)
	beta.Add(beta, big.NewInt(7))
	beta.Exp(beta, exp, params.P)
	if y.Add(beta, y).Mod(y, big.NewInt(2)).Int64() == 0 {
		y = beta
	} else {
		y = beta.Sub(params.P, beta)
	}
	return x, y
}

func compress(x, y *big.Int) []byte {
	two := big.NewInt(2)
	rem := two.Mod(y, two).Uint64()
	rem += 2
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(rem))
	rest := x.Bytes()
	pad := 32 - len(rest)
	if pad != 0 {
		zeroes := make([]byte, pad)
		rest = append(zeroes, rest...)
	}
	return append(b[1:], rest...)
}

func privToPub(key []byte) []byte {
	return compress(curve.ScalarBaseMult(key))
}

func hash160(data []byte) []byte {
	sha := sha256.New()
	ripe := ripemd160.New()
	sha.Write(data)
	ripe.Write(sha.Sum(nil))
	return ripe.Sum(nil)
}

func uint32ToByte(i uint32) []byte {
	a := make([]byte, 4)
	binary.BigEndian.PutUint32(a, i)
	return a
}

func addPubKeys(k1, k2 []byte) []byte {
	x1, y1 := expand(k1)
	x2, y2 := expand(k2)
	return compress(curve.Add(x1, y1, x2, y2))
}

// GetXPub returns an xpub string from a PublicKey.
// See BIP32 https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki
func (pk *PublicKey) GetXPub() string {
	depth := uint16ToByte(uint16(pk.Depth % 256))

	// fmt.Printf("Version:     %+v\n", pk.Version)
	// fmt.Printf("Depth:       %+v\n", depth)
	// fmt.Printf("FingerPrint: %+v\n", pk.FingerPrint)
	// fmt.Printf("ChildIndex:  %+v\n", pk.ChildIndex)
	// fmt.Printf("ChainCode:   %+v\n", pk.ChainCode)
	// fmt.Printf("PublicKey:   %+v\n", pk.PublicKey)

	//bindata = vbytes||depth||fingerprint||i||chaincode||key
	bindata := make([]byte, 78)
	copy(bindata, pk.Version)
	copy(bindata[4:], depth)
	copy(bindata[5:], pk.ParentFingerPrint)
	copy(bindata[9:], pk.ChildIndex)
	copy(bindata[13:], pk.ChainCode)
	copy(bindata[45:], pk.PublicKey)
	chksum := Sha256d(bindata)[:4]

	// fmt.Printf("Checksum:    %+v\n", chksum)

	return base58.Encode(append(bindata, chksum...))
}

func uint16ToByte(i uint16) []byte {
	a := make([]byte, 2)
	binary.BigEndian.PutUint16(a, i)
	return a[1:]
}

func byteToUint16(b []byte) uint16 {
	if len(b) == 1 {
		zero := make([]byte, 1)
		b = append(zero, b...)
	}
	return binary.BigEndian.Uint16(b)
}

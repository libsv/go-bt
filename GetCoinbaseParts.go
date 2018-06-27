package cryptolib

/*
Here is a real example coinbase broken down...

01000000 .............................. Version
01 .................................... Number of inputs
| 00000000000000000000000000000000
| 00000000000000000000000000000000 ...  Previous outpoint TXID
| ffffffff ............................ Previous outpoint index
|
| 43 .................................. Input coinbase count of bytes (4 block height + 12 (extra nonces) + Arbitary data length)
| |
| | 03 ................................ Bytes in height
| | | bfea07 .......................... Height: 518847
| |
| | 322f53696d6f6e204f72646973682061    (I think the 32 is wrong - we don't need another var int length here.)
| | 6e642053747561727420467265656d61
| | 6e206d61646520746869732068617070
| | 656e2f ............................ /Simon Ordish and Stuart Freeman made this happen/
| | 9a46 .............................. nonce.dat from seed1.hashzilla.io
| | 434790f7dbde ..................     Extranonce 1 (6 bytes)
| | a3430000 .......................... Extranonce 2 (4 bytes)
|
| ffffffff ............................ Sequence

01 .................................... Output count of bytes (1 or 2 if segwit)
| 8a08ac4a00000000 .................... Satoshis (25.04275756 BTC)
| 19 .................................. Size of locking script
| 76a9 ................................ opDUP, opHASH160
| 14 .................................. Length of hash - 20 bytes
| 8bf10d323ac757268eb715e613cb8e8e1d17
| 93aa ................................ Wallet (20 bytes)
| 88ac ................................ opEQUALVERIFY, opCHECKSIG
| 00000000 ............................ Locktime

*/

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
)

const (
	opDUP         = 0x76 // Duplicate the top item in the stack
	opHASH160     = 0xa9 // Return RIPEMD160(SHA256(x)) hash of top item
	opEQUALVERIFY = 0x88 // Same as OP_EQUAL, but run OP_VERIFY after to halt if not TRUE
	opCHECKSIG    = 0xac // Pop a public key and signature and validate the signature for the transaction's hashed data, return TRUE if matching
)

// BuildCoinbase comment
func BuildCoinbase(c1 []byte, c2 []byte, extraNonce1 string, extraNonce2 string) []byte {
	e1, _ := hex.DecodeString(extraNonce1)
	e2, _ := hex.DecodeString(extraNonce2)

	a := []byte{}
	a = append(a, c1...)
	a = append(a, e1...)
	a = append(a, e2...)
	a = append(a, c2...)
	return a
}

// GetCoinbaseParts comment
func GetCoinbaseParts(height uint32, coinbaseValue uint64, defaultWitnessCommitment string, coinbaseText string, walletAddress string) (coinbase1 []byte, coinbase2 []byte, err error) {
	coinbase1 = makeCoinbase1(height, coinbaseText)
	ot, err := makeCoinbaseOutputTransactions(coinbaseValue, defaultWitnessCommitment, walletAddress)
	coinbase2 = makeCoinbase2(ot)

	return
}

func makeCoinbaseInputTransaction(coinbaseData []byte) []byte {
	buf := make([]byte, 32)                              // 32 bytes - All bits are zero: Not a transaction hash reference
	buf = append(buf, []byte{0xff, 0xff, 0xff, 0xff}...) // 4 bytes - All bits are ones: 0xFFFFFFFF
	buf = append(buf, VarInt(len(coinbaseData))...)      // Length of the coinbase data, from 2 to 100 bytes
	buf = append(buf, coinbaseData...)                   // Arbitrary data used for extra nonce and mining tags. In v2 blocks; must begin with block height
	buf = append(buf, []byte{0xff, 0xff, 0xff, 0xff}...) //  4 bytes = Set to 0xFFFFFFFF
	return buf
}

// AddressToScript comment
func AddressToScript(address string) (script []byte, err error) {
	decoded, err := DecodeString(address)

	if err != nil {
		return nil, err
	}

	if len(decoded) != 25 {
		return nil, fmt.Errorf("invalid address length for '%s'", address)
	}

	pubkey := decoded[1 : len(decoded)-4]

	ret := []byte{
		opDUP,
		opHASH160,
		0x14,
	}
	ret = append(ret, pubkey...)
	ret = append(ret, opEQUALVERIFY)
	ret = append(ret, opCHECKSIG)

	return ret, nil
}

func makeCoinbaseOutputTransactions(coinbaseValue uint64, defaultWitnessCommitment string, wallet string) ([]byte, error) {
	lockingScript, err := AddressToScript(wallet)
	if err != nil {
		return nil, err
	}

	var buf = []byte{}

	buf = append(buf, GetLittleEndianBytes(1, 4)...) // 4 bytes - version number
	buf = append(buf, make([]byte, 4)...)            // 4 bytes of zeros

	binary.LittleEndian.PutUint64(buf[0:], coinbaseValue)

	buf = append(buf, VarInt(len(lockingScript))...)
	buf = append(buf, lockingScript...)

	numberOfTransactions := 1
	if defaultWitnessCommitment != "" {
		numberOfTransactions++
		byteArr := make([]byte, 8)
		buf = append(buf, byteArr...)
		wc, err := hex.DecodeString(defaultWitnessCommitment)
		if err != nil {
			log.Printf("Error decoding witness commitment: %+v", err)
		}
		wcl := VarInt(len(wc))
		buf = append(buf, wcl...)
		buf = append(buf, wc...)
	}

	buf = append(VarInt(numberOfTransactions), buf...)
	return buf, nil
}

// MakeCoinbase1 comment
func makeCoinbase1(height uint32, coinbaseText string) []byte {
	spaceForExtraNonce := 12

	blockHeightBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(blockHeightBytes, uint32(height)) // Block height

	arbitraryData := []byte{}
	arbitraryData = append(arbitraryData, 0x03)
	arbitraryData = append(arbitraryData, blockHeightBytes[:3]...)
	arbitraryData = append(arbitraryData, []byte(coinbaseText)...)

	//Arbitrary data should leave enough space for the extra nonce
	if len(arbitraryData) > (100 - spaceForExtraNonce) {
		arbitraryData = arbitraryData[:100-spaceForExtraNonce] // Slice the arbitrary text so everything fits in 100 bytes
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, 1) // Version

	buf = append(buf, 0x01)                              // Number of input transaction - always one
	buf = append(buf, make([]byte, 32)...)               // Transaction hash - 4 bytes all bits are zero
	buf = append(buf, []byte{0xff, 0xff, 0xff, 0xff}...) // Coinbase data size - 4 bytes - All bits are ones: 0xFFFFFFFF (ffffffff)

	buf = append(buf, VarInt(len(arbitraryData)+spaceForExtraNonce)...) // Length of the coinbase data, from 2 to 100 bytes
	buf = append(buf, arbitraryData...)

	return buf
}

func makeCoinbase2(ot []byte) []byte {
	sq := []byte{0xff, 0xff, 0xff, 0xff}
	lt := make([]byte, 4)

	ot = append(sq, ot...)
	ot = append(ot, lt...)

	return ot
}

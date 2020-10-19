package block

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
	"log"

	"github.com/libsv/libsv/bt/output"
	"github.com/libsv/libsv/utils"
)

// BuildCoinbase recombines the different parts of the coinbase transaction.
// See https://arxiv.org/pdf/1703.06545.pdf section 2.2 for more info.
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

// GetCoinbaseParts returns the two split coinbase parts from coinbase metadata.
// See https://arxiv.org/pdf/1703.06545.pdf section 2.2 for more info.
func GetCoinbaseParts(height uint32, coinbaseValue uint64, defaultWitnessCommitment string, coinbaseText string, walletAddress string, minerIDBytes []byte) (coinbase1 []byte, coinbase2 []byte, err error) {
	coinbase1 = makeCoinbase1(height, coinbaseText)

	ot, err := makeCoinbaseOutputTransactions(coinbaseValue, defaultWitnessCommitment, walletAddress, minerIDBytes)
	if err != nil {
		return
	}

	coinbase2 = makeCoinbase2(ot)

	return
}

func makeCoinbaseInputTransaction(coinbaseData []byte) []byte {
	buf := make([]byte, 32)                                       // 32 bytes - All bits are zero: Not a transaction hash reference
	buf = append(buf, []byte{0xff, 0xff, 0xff, 0xff}...)          // 4 bytes - All bits are ones: 0xFFFFFFFF
	buf = append(buf, utils.VarInt(uint64(len(coinbaseData)))...) // Length of the coinbase data, from 2 to 100 bytes
	buf = append(buf, coinbaseData...)                            // Arbitrary data used for extra nonce and mining tags. In v2 blocks; must begin with block height
	buf = append(buf, []byte{0xff, 0xff, 0xff, 0xff}...)          //  4 bytes = Set to 0xFFFFFFFF
	return buf
}

func makeCoinbaseOutputTransactions(coinbaseValue uint64, defaultWitnessCommitment string, wallet string, minerIDBytes []byte) ([]byte, error) {
	o, err := output.NewP2PKHFromAddress(wallet, coinbaseValue)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 8)

	binary.LittleEndian.PutUint64(buf[0:], coinbaseValue)

	buf = append(buf, utils.VarInt(uint64(len(*o.LockingScript)))...)
	buf = append(buf, *o.LockingScript...)

	numberOfTransactions := 1
	if defaultWitnessCommitment != "" {
		numberOfTransactions++
		byteArr := make([]byte, 8)
		buf = append(buf, byteArr...)
		wc, err := hex.DecodeString(defaultWitnessCommitment)
		if err != nil {
			log.Printf("Error decoding witness commitment: %+v", err)
			return nil, err
		}
		wcl := utils.VarInt(uint64(len(wc)))
		buf = append(buf, wcl...)
		buf = append(buf, wc...)
	}

	if len(minerIDBytes) > 0 {
		numberOfTransactions++

		byteArr := make([]byte, 8) // 8 bytes of 0 = 0 satoshis.
		buf = append(buf, byteArr...)
		buf = append(buf, utils.VarInt(uint64(len(minerIDBytes)))...)
		buf = append(buf, minerIDBytes...)
	}

	buf = append(utils.VarInt(uint64(numberOfTransactions)), buf...)
	return buf, nil
}

func makeCoinbase1(height uint32, coinbaseText string) []byte {
	spaceForExtraNonce := 12

	blockHeightBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(blockHeightBytes, uint32(height)) // Block height

	arbitraryData := []byte{}
	arbitraryData = append(arbitraryData, 0x03)
	arbitraryData = append(arbitraryData, blockHeightBytes[:3]...)
	arbitraryData = append(arbitraryData, []byte(coinbaseText)...)

	// Arbitrary data should leave enough space for the extra nonce
	if len(arbitraryData) > (100 - spaceForExtraNonce) {
		arbitraryData = arbitraryData[:100-spaceForExtraNonce] // Slice the arbitrary text so everything fits in 100 bytes
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, 1) // Version

	buf = append(buf, 0x01)                              // Number of input transaction - always one
	buf = append(buf, make([]byte, 32)...)               // Transaction hash - 4 bytes all bits are zero
	buf = append(buf, []byte{0xff, 0xff, 0xff, 0xff}...) // Coinbase data size - 4 bytes - All bits are ones: 0xFFFFFFFF (ffffffff)

	buf = append(buf, utils.VarInt(uint64(len(arbitraryData)+spaceForExtraNonce))...) // Length of the coinbase data, from 2 to 100 bytes
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

func chunk(msg string, limit int) (chunks []string) {

	chunkNumber := (len(msg) / limit) + 1

	for i := 0; i < chunkNumber; i++ {
		s := i * limit
		e := (i + 1) * limit
		if e > len(msg) {
			e = len(msg)
		}

		chunks = append(chunks, msg[s:e])
	}
	return
}

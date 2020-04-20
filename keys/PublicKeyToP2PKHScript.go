package keys

import (
	"github.com/libsv/libsv/crypto"
	"github.com/libsv/libsv/utils"
)

// TODO: consider moving to transactions/script

// PublicKeyToP2PKHScript turns a public key string (in compressed format) into a P2PKH script.
// Example:
// from 023717efaec6761e457f55c8417815505b695209d0bbfed8c3265be425b373c2d6
// to 76a9144d5d1920331b71735a97a606d9734aed83cb3dfa88ac
func PublicKeyToP2PKHScript(pubkey []byte) []byte {
	hash := crypto.Hash160(pubkey)

	ret := []byte{
		utils.OpDUP,
		utils.OpHASH160,
		0x14,
	}
	ret = append(ret, hash...)
	ret = append(ret, utils.OpEQUALVERIFY)
	ret = append(ret, utils.OpCHECKSIG)

	return ret
}

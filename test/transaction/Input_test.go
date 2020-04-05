package transaction

import (
	"encoding/hex"
	"github.com/jadwahab/libsv/transaction"
	"testing"
)

const input = "4c6ec863cf3e0284b407a1a1b8138c76f98280812cb9653231f385a0305fc76f010000006b483045022100f01c1a1679c9437398d691c8497f278fa2d615efc05115688bf2c3335b45c88602201b54437e54fb53bc50545de44ea8c64e9e583952771fcc663c8687dc2638f7854121037e87bbd3b680748a74372640628a8f32d3a841ceeef6f75626ab030c1a04824fffffffff"

func TestNew(t *testing.T) {
	bytes, _ := hex.DecodeString(input)
	i, s := transaction.NewInputFromBytes(bytes)

	// t.Errorf("\n%s\n", i)

	if s != 148 {
		t.Errorf("Expected 148, got %d", s)
	}

	if i.PreviousTxOutIndex != 1 {
		t.Errorf("Expected 1, got %d", i.PreviousTxOutIndex)
	}

	if len(*i.SigScript) != 107 {
		t.Errorf("Expected 107, got %d", len(*i.SigScript))
	}

	if i.SequenceNumber != 0xFFFFFFFF {
		t.Errorf("Expected 0xFFFFFFFF, got %x", i.SequenceNumber)
	}
}

// func TestArbitraryText(t *testing.T) {
// 	const coinbase = "0000000000000000000000000000000000000000000000000000000000000000ffffffff4303bfea07322f53696d6f6e204f726469736820616e642053747561727420467265656d616e206d61646520746869732068617070656e2f9a46434790f7dbdea3430000ffffffff018a08ac4a000000001976a9148bf10d323ac757268eb715e613cb8e8e1d1793aa88ac00000000"
// 	// const coinbase = "0000000000000000000000000000000000000000000000000000000000000000ffffffff2003298b082f626d67706f6f6c2e636f6d2f31646b62ff0b02843058aa9d410000ffffffff018524834a000000001976a9148be87b3978d8ef936b30ddd4ed903f8da7abd27788ac00000000"
// 	bytes, _ := hex.DecodeString(coinbase)
// 	i, _ := NewInput(bytes)

// 	length, size := libsv.DecodeVarInt(*i.script)
// 	heightPart := *i.script[size : size+int(length)]
// 	var heightBytes [4]byte
// 	copy(heightBytes[:], heightPart)
// 	height := binary.LittleEndian.Uint32(heightBytes[:])

// 	// t.Errorf("\nHeight: %d\n%s\n\n%x\n", height, string(i.script[size+int(length):]), i.script[size+int(length):])

// 	if height != 518847 {
// 		t.Errorf("Expected 518847, got %d", height)
// 	}
// }

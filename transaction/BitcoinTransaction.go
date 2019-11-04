package transaction

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"bitbucket.org/simon_ordish/cryptolib"
)

/*
General format of a Bitcoin transaction (inside a block)
--------------------------------------------------------
Field            Description                                                               Size

Version no	     currently 1	                                                             4 bytes

Flag	           If present, always 0001, and indicates the presence of witness data       optional 2 byte array

In-counter  	   positive integer VI = VarInt                                              1 - 9 bytes

list of inputs	 the first input of the first transaction is also called "coinbase"        <in-counter>-many inputs
                 (its content was ignored in earlier versions)

Out-counter    	 positive integer VI = VarInt                                              1 - 9 bytes

list of outputs  the outputs of the first transaction spend the mined                      <out-counter>-many outputs
								 bitcoins for the block

Witnesses        A list of witnesses, 1 for each input, omitted if flag above is missing	 variable, see Segregated_Witness

lock_time        if non-zero and sequence numbers are < 0xFFFFFFFF: block height or        4 bytes
                 timestamp when transaction is final
*/

// Signature constants
const (
	SighashAll          = 0x00000001
	SighashNone         = 0x00000002
	SighashSingle       = 0x00000003
	SighashForkID       = 0x00000040
	SighashAnyoneCanPay = 0x00000080
)

// A BitcoinTransaction wraps a bitcoin transaction
type BitcoinTransaction struct {
	Bytes    []byte
	Version  uint32
	Witness  bool
	Inputs   []*Input
	Outputs  []*Output
	Locktime uint32
}

// New comment
func New() *BitcoinTransaction {
	return &BitcoinTransaction{
		Version: 1,
	}
}

// NewFromString takes a hex string representation of a bitcoin transaction
// and returns a BitcoinTransaction object
func NewFromString(str string) (*BitcoinTransaction, error) {
	bytes, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return NewFromBytes(bytes)
}

// NewFromBytes takes an array of bytes and constructs a BitcoinTransaction
func NewFromBytes(bytes []byte) (*BitcoinTransaction, error) {
	bt := BitcoinTransaction{
		Bytes: bytes,
	}

	var offset = 0

	bt.Version = binary.LittleEndian.Uint32(bytes[offset:4])
	offset += 4

	// There is an optional Flag of 2 bytes after the version. It is always "0001".
	if bytes[4] == 0x00 && bytes[5] == 0x01 {
		bt.Witness = true
		offset += 2
	}

	inputCount, size := cryptolib.DecodeVarInt(bt.Bytes[offset:])
	offset += size

	var i uint64
	for ; i < inputCount; i++ {
		input, size := NewInputFromBytes(bt.Bytes[offset:])
		offset += size

		bt.Inputs = append(bt.Inputs, input)
	}

	outputCount, size := cryptolib.DecodeVarInt(bt.Bytes[offset:])
	offset += size

	for i = 0; i < outputCount; i++ {
		output, size := NewOutput(bt.Bytes[offset:])
		offset += size
		bt.Outputs = append(bt.Outputs, output)
	}

	bt.Locktime = binary.LittleEndian.Uint32(bytes[offset:])

	return &bt, nil
}

// HasWitnessData returns true if the optional Witness flag == 0001
func (bt *BitcoinTransaction) HasWitnessData() bool {
	return bt.Witness
}

// AddInput comment
func (bt *BitcoinTransaction) AddInput(input *Input) {
	bt.Inputs = append(bt.Inputs, input)
}

// InputCount returns the number of transaction inputs
func (bt *BitcoinTransaction) InputCount() int {
	return len(bt.Inputs)
}

// OutputCount returns the number of transaction inputs
func (bt *BitcoinTransaction) OutputCount() int {
	return len(bt.Outputs)
}

// AddOutput comment
func (bt *BitcoinTransaction) AddOutput(output *Output) {
	bt.Outputs = append(bt.Outputs, output)
}

// IsCoinbase determines if this transaction is a coinbase by
// seeing if any of the inputs have no inputs
func (bt *BitcoinTransaction) IsCoinbase() bool {
	if len(bt.Inputs) != 1 {
		return false
	}

	fmt.Println(bt.Inputs[0].PreviousTxOutIndex)
	fmt.Println(bt.Inputs[0].SequenceNumber)
	for _, v := range bt.Inputs[0].PreviousTxHash {
		if v != 0x00 {
			return false
		}
	}

	if bt.Inputs[0].PreviousTxOutIndex == 0xFFFFFFFF || bt.Inputs[0].SequenceNumber == 0xFFFFFFFF {
		return true
	}

	return false
}

// GetInputs comment
func (bt *BitcoinTransaction) GetInputs() []*Input {
	return bt.Inputs
}

// GetOutputs comment
func (bt *BitcoinTransaction) GetOutputs() []*Output {
	return bt.Outputs
}

// Hex comment
func (bt *BitcoinTransaction) Hex() []byte {
	return bt.hex(0, nil)
}

// HexWithClearedInputs comment
func (bt *BitcoinTransaction) HexWithClearedInputs(index int, scriptPubKey []byte) []byte {
	return bt.hex(index, scriptPubKey)
}

func (bt *BitcoinTransaction) hex(index int, scriptPubKey []byte) []byte {
	hex := make([]byte, 0)

	hex = append(hex, cryptolib.GetLittleEndianBytes(bt.Version, 4)...)

	if bt.Witness {
		hex = append(hex, 0x00)
		hex = append(hex, 0x01)
	}

	hex = append(hex, cryptolib.VarInt(uint64(len(bt.GetInputs())))...)

	for i, in := range bt.GetInputs() {
		script := in.Hex(scriptPubKey != nil)
		if i == index && scriptPubKey != nil {
			hex = append(hex, cryptolib.VarInt(uint64(len(scriptPubKey)))...)
			hex = append(hex, scriptPubKey...)
		} else {
			hex = append(hex, script...)
		}
	}

	hex = append(hex, cryptolib.VarInt(uint64(len(bt.GetOutputs())))...)
	for _, out := range bt.GetOutputs() {
		hex = append(hex, out.Hex()...)
	}

	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, bt.Locktime)
	hex = append(hex, lt...)

	return hex
}

// Sign comment
// func (bt *BitcoinTransaction) Sign(privateKey *btcec.PrivateKey, sigType int) {
// 	if sigType == 0 {
// 		sigType = SighashAll | SighashForkID
// 	}

// 	hashData := hash160(privateKey.PubKey().SerializeCompressed())
// 	fmt.Printf("hashdata: %x", hashData)
// 	// Go through each input and calculate a signature and then add it

// 	scriptPubKey, _ := hex.DecodeString("a9140e95261082d65c384a6106f114474bc0784ba67e87")

// 	for i, in := range bt.GetInputs() {
// 		if bytes.Compare(in.script[3:23], hashData) == 0 {
// 			hex := bt.HexWithClearedInputs(i, scriptPubKey)
// 			hex = append(hex, cryptolib.GetLittleEndianBytes(0x01, 4)...)
// 			log.Printf("hex: %x\n", hex)

// 			hash := sha256.Sum256(hex)

// 			log.Printf("hash: %x\n", hash)

// 			signature, err := privateKey.Sign(hash[:])
// 			if err != nil {
// 				fmt.Println(err)
// 				return
// 			}

// 			// Serialize and display the signature.
// 			fmt.Printf("Serialized Signature: %x\n", signature.Serialize())

// 		}

// 	}
// 	// hex := bt.HexWithClearedInputs()
// 	// parts, _ := cryptolib.DecodeParts(i.script)
// 	// if parts[0][0] == opZERO {
// 	// 	redeemScript, err := NewRedeemScriptFromElectrum(hex.EncodeToString(parts[len(parts)-1]))
// 	// 	if err != nil {
// 	// 		log.Println(err)
// 	// 	}

// 	// 	signatures := parts[1 : len(parts)-1]

// 	// 	for i, signature := range signatures {
// 	// 		if signature[0] == 0xff {
// 	// 			// xprivKey, err := cryptolib.NewPrivateKey("xprv9s21ZrQH143K2beTKhLXFRWWFwH8jkwUssjk3SVTiApgmge7kNC3jhVc4NgHW8PhW2y7BCDErqnKpKuyQMjqSePPJooPJowAz5BVLThsv6c")
// 	// 			xprivKey, err := cryptolib.NewPrivateKey("xprv9s21ZrQH143K3ShHqGb2ago1pjts78QvhAtYUbe1kPraUtjkxaftf28Pc6LdHKBAzi2jAH3EhQWgibbJxMFDW1yS8ZrPy172LEvwddxV55D")
// 	// 			if err != nil {
// 	// 				log.Println(err)
// 	// 			}

// 	// 			privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), xprivKey.PrivateKey)

// 	// 			signature, err := privKey.Sign(redeemScript.getRedeemScriptHash())
// 	// 			if err != nil {
// 	// 				log.Println(err)
// 	// 			}
// 	// 			signatures[i] = signature.Serialize()
// 	// 			fmt.Printf("NEW SIG %d: %x\n", i, signature.Serialize())
// 	// 		} else {
// 	// 			fmt.Printf("OLD SIG %d: %x\n", i, signature)
// 	// 		}
// 	// 	}

// 	// 	fmt.Printf("%v", parts)
// 	// }

// }

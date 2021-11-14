package interpreter

import (
	"math"
	"math/big"

	"github.com/libsv/go-bt/v2/bscript/interpreter/errs"
)

// Number a number larger than int64.
type Number struct {
	val          *big.Int
	afterGenesis bool
}

var zero = big.NewInt(0)
var one = big.NewInt(1)

// NewNumber returns a number.
func NewNumber(bb []byte, scriptNumLen int, requireMinimal, afterGenesis bool) (*Number, error) {
	if len(bb) > scriptNumLen {
		return nil, errs.NewError(
			errs.ErrNumberTooBig,
			"numeric value encoded as %x is %d bytes which exceeds the max allowed of %d",
			bb, len(bb), scriptNumLen,
		)
	}
	if requireMinimal {
		if err := checkMinimalDataEncoding(bb); err != nil {
			return nil, err
		}
	}
	if len(bb) == 0 {
		return &Number{
			afterGenesis: afterGenesis,
			val:          big.NewInt(0),
		}, nil
	}
	v := new(big.Int)
	for i, b := range bb {
		v.Or(v, new(big.Int).Lsh(new(big.Int).SetBytes([]byte{b}), uint(8*i)))
	}

	if bb[len(bb)-1]&0x80 != 0 {
		shift := big.NewInt(int64(0x80))
		shift.Not(shift.Lsh(shift, uint(8*(len(bb)-1))))
		v.And(v, shift).Neg(v)
	}
	return &Number{
		val:          v,
		afterGenesis: afterGenesis,
	}, nil
}

// Add adds the receiver and the number, sets the result over the receiver and returns.
func (n *Number) Add(o *Number) *Number {
	result := n.val.Add(n.val, o.val)
	*n.val = *result
	return n
}

// Sub subtracts the number from the receiver, sets the result over the receiver and returns.
func (n *Number) Sub(o *Number) *Number {
	result := n.val.Sub(n.val, o.val)
	*n.val = *result
	return n
}

// Mul multiplies the receiver by the number, sets the result over the receiver and returns.
func (n *Number) Mul(o *Number) *Number {
	result := n.val.Mul(n.val, o.val)
	*n.val = *result
	return n
}

// Div divides the receiver by the number, sets the result over the receiver and returns.
func (n *Number) Div(o *Number) *Number {
	result := n.val.Quo(n.val, o.val)
	*n.val = *result
	return n
}

// Mod divides the receiver by the number, sets the remainder over the receiver and returns.
func (n *Number) Mod(o *Number) *Number {
	result := n.val.Rem(n.val, o.val)
	*n.val = *result
	return n
}

// LessThanInt returns true if the receiver is smaller than the integer passed.
func (n *Number) LessThanInt(i int64) bool {
	return n.LessThan(&Number{val: big.NewInt(i)})
}

// LessThan returns true if the receiver is smaller than the number passed.
func (n *Number) LessThan(o *Number) bool {
	return n.val.Cmp(o.val) == -1
}

// LessThanOrEqual returns ture if the receiver is smaller or equal to the number passed.
func (n *Number) LessThanOrEqual(o *Number) bool {
	return n.val.Cmp(o.val) < 1
}

// GreaterThanInt returns true if the receiver is larger than the integer passed.
func (n *Number) GreaterThanInt(i int64) bool {
	return n.GreaterThan(&Number{val: big.NewInt(i)})
}

// GreaterThan returns true if the receiver is larger than the number passed.
func (n *Number) GreaterThan(o *Number) bool {
	return n.val.Cmp(o.val) == 1
}

// GreaterThanOrEqual returns true if the receiver is larger or equal to the number passed.
func (n *Number) GreaterThanOrEqual(o *Number) bool {
	return n.val.Cmp(o.val) > -1
}

// EqualInt returns true if the receiver is equal to the integer passed.
func (n *Number) EqualInt(i int64) bool {
	return n.Equal(&Number{val: big.NewInt(i)})
}

// Equal returns true if the receiver is equal to the number passed.
func (n *Number) Equal(o *Number) bool {
	return n.val.Cmp(o.val) == 0
}

// IsZero return strue if hte receiver equals zero.
func (n *Number) IsZero() bool {
	return n.val.Cmp(zero) == 0
}

// Incr increment the receiver by one.
func (n *Number) Incr() *Number {
	result := n.val.Add(n.val, one)
	*n.val = *result
	return n
}

// Decr decrement the receiver by one.
func (n *Number) Decr() *Number {
	result := n.val.Sub(n.val, one)
	*n.val = *result
	return n
}

// Neg sets the receiver to the negative of the receiver.
func (n *Number) Neg() *Number {
	result := n.val.Neg(n.val)
	*n.val = *result
	return n
}

// Abs sets the receiver to the absolute value of hte receiver.
func (n *Number) Abs() *Number {
	result := n.val.Abs(n.val)
	*n.val = *result
	return n
}

func (n *Number) Int() int {
	return int(n.val.Int64())
}

func (n *Number) Int32() int32 {
	return int32(n.val.Int64())
}

func (n *Number) Int64() int64 {
	return n.val.Int64()
}

func (n *Number) Set(i int64) *Number {
	val := big.NewInt(i)
	*n.val = *val
	return n
}

// Bytes return the receiver in Bytes form.
func (n *Number) Bytes() []byte {
	if n.val.Int64() == 0 {
		return []byte{}
	}
	isNegative := n.val.Cmp(zero) == -1

	var bb []byte
	if !n.afterGenesis {
		v := n.val.Int64()
		if v > math.MaxInt32 {
			bb = big.NewInt(int64(math.MaxInt32)).Bytes()
		} else if v < math.MinInt32 {
			bb = big.NewInt(int64(math.MinInt32)).Bytes()
		}
	}
	if bb == nil {
		bb = n.val.Bytes()
	}

	tmp := make([]byte, 0, len(bb)+1)
	cpy := new(big.Int).SetBytes(n.val.Bytes())
	for cpy.Cmp(zero) == 1 {
		tmp = append(tmp, byte(cpy.Int64()&0xff))
		cpy.Rsh(cpy, 8)
	}

	if tmp[len(tmp)-1]&0x80 != 0 {
		extraByte := byte(0x00)
		if isNegative {
			extraByte = 0x80
		}
		tmp = append(tmp, extraByte)
	} else if isNegative {
		tmp[len(tmp)-1] |= 0x80
	}

	return tmp
}

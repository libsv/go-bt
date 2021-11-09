package interpreter

import (
	"math"
	"math/big"

	"github.com/libsv/go-bt/v2/bscript/interpreter/errs"
)

type Number struct {
	val          *big.Int
	afterGenesis bool
}

var zero = big.NewInt(0)
var one = big.NewInt(1)

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

func (n *Number) Add(o *Number) *Number {
	result := n.val.Add(n.val, o.val)
	*n.val = *result
	return n
}

func (n *Number) Sub(o *Number) *Number {
	result := n.val.Sub(n.val, o.val)
	*n.val = *result
	return n
}

func (n *Number) Mul(o *Number) *Number {
	result := n.val.Mul(n.val, o.val)
	*n.val = *result
	return n
}

func (n *Number) Div(o *Number) *Number {
	result := n.val.Quo(n.val, o.val)
	*n.val = *result
	return n
}

func (n *Number) Mod(o *Number) *Number {
	result := n.val.Rem(n.val, o.val)
	*n.val = *result
	return n
}

func (n *Number) LessThanInt(i int64) bool {
	return n.LessThan(&Number{val: big.NewInt(i)})
}

func (n *Number) LessThan(o *Number) bool {
	return n.val.Cmp(o.val) == -1
}

func (n *Number) LessThanOrEqual(o *Number) bool {
	return n.val.Cmp(o.val) < 1
}

func (n *Number) GreaterThanInt(i int64) bool {
	return n.GreaterThan(&Number{val: big.NewInt(i)})
}

func (n *Number) GreaterThan(o *Number) bool {
	return n.val.Cmp(o.val) == 1
}

func (n *Number) GreaterThanOrEqual(o *Number) bool {
	return n.val.Cmp(o.val) > -1
}

func (n *Number) EqualInt(i int64) bool {
	return n.Equal(&Number{val: big.NewInt(i)})
}

func (n *Number) Equal(o *Number) bool {
	return n.val.Cmp(o.val) == 0
}

func (n *Number) IsZero() bool {
	return n.val.Cmp(zero) == 0
}

func (n *Number) Incr() *Number {
	result := n.val.Add(n.val, one)
	*n.val = *result
	return n
}

func (n *Number) Decr() *Number {
	result := n.val.Sub(n.val, one)
	*n.val = *result
	return n
}

func (n *Number) Neg() *Number {
	result := n.val.Neg(n.val)
	*n.val = *result
	return n
}

func (n *Number) Abs() *Number {
	result := n.val.Abs(n.val)
	*n.val = *result
	return n
}

func (n *Number) Bytes() []byte {
	if n.val.Int64() == 0 {
		return []byte{}
	}
	isNegative := n.val.Cmp(big.NewInt(0)) == -1
	if isNegative {
		n.val.Neg(n.val)
	}

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

	//tmp := make([]byte, len(bb), len(bb)+1)
	//for i := len(bb) - 1; i >= 0; i-- {
	//	tmp[i] = bb[i]
	//}
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

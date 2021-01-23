// Copyright (c) 2016-2021, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"io"
	"math/big"
	"strings"

	"github.com/jancajthaml-openbank/ledger-unit/support/cast"
)

// Dec represents "infinite" precision decimal number
type Dec struct {
	unscaled big.Int
	scale    int32
}

var zeros = "0000000000000000000000000000000000000000000000000000000000000000"
var lzeros = int32(len(zeros))
var bigInt = [...]*big.Int{
	big.NewInt(0),
	big.NewInt(1),
	big.NewInt(2),
	big.NewInt(3),
	big.NewInt(4),
	big.NewInt(5),
	big.NewInt(6),
	big.NewInt(7),
	big.NewInt(8),
	big.NewInt(9),
	big.NewInt(10),
}

var exp10cache [64]big.Int = func() [64]big.Int {
	e10, e10i := [64]big.Int{}, bigInt[1]
	for i := range e10 {
		e10[i].Set(e10i)
		e10i = new(big.Int).Mul(e10i, bigInt[10])
	}
	return e10
}()

// Sign of receiver
func (x *Dec) Sign() int {
	return x.unscaled.Sign()
}

// Add number to receiver
func (x *Dec) Add(y *Dec) {
	if x == nil || y == nil {
		return
	}
	if x.scale == y.scale {
		x.unscaled.Add(&x.unscaled, &y.unscaled)
	} else if x.scale > y.scale {
		y.rescale(x.scale)
		x.unscaled.Add(&x.unscaled, &y.unscaled)
	} else {
		x.rescale(y.scale)
		x.unscaled.Add(&x.unscaled, &y.unscaled)
	}
}

// Sub number from receiver
func (x *Dec) Sub(y *Dec) {
	if x == nil || y == nil {
		return
	}
	if x.scale == y.scale {
		x.unscaled.Sub(&x.unscaled, &y.unscaled)
	} else if x.scale > y.scale {
		y.rescale(x.scale)
		x.unscaled.Sub(&x.unscaled, &y.unscaled)
	} else {
		x.rescale(y.scale)
		x.unscaled.Sub(&x.unscaled, &y.unscaled)
	}
}

func (x *Dec) rescale(newScale int32) {
	if x == nil {
		return
	}
	shift := newScale - x.scale
	switch {
	case shift < 0:
		e := exp10(-shift)
		x.unscaled.Set(new(big.Int).Quo(&x.unscaled, e))
		x.scale = newScale
	case shift > 0:
		e := exp10(shift)
		x.unscaled.Set(new(big.Int).Mul(&x.unscaled, e))
		x.scale = newScale
	}
}

func (x *Dec) String() string {
	if x == nil || x.Sign() == 0 {
		return "0.0"
	}

	numbers := x.unscaled.Text(10)
	sign := x.unscaled.Sign()

	if x.scale <= 0 {
		if x.scale != 0 && sign != 0 {
			n := -x.scale
			for i := int32(0); i < n; i += lzeros {
				if n > i+lzeros {
					numbers += zeros
				} else {
					numbers += zeros[0 : n-i]
				}
			}
		}
		return numbers
	}

	var negbit int32
	if sign == -1 {
		negbit = 1
	}

	lens := int32(len(numbers))

	if lens-negbit > x.scale {
		return numbers[:lens-x.scale] + "." + numbers[lens-x.scale:]
	}

	var result string
	if negbit == 1 {
		result = "-0."
	} else {
		result = "0."
	}

	n := x.scale - lens + negbit
	for i := int32(0); i < n; i += lzeros {
		if n > i+lzeros {
			result += zeros
		} else {
			result += zeros[0 : n-i]
		}
	}

	result += numbers[negbit:]

	return result
}

// SetString value
func (x *Dec) SetString(s string) bool {
	r := strings.NewReader(s)
	unscaled := make([]byte, 0, 128)
	dot := -1
loop:
	ch, _, err := r.ReadRune()
	switch {
	case err == io.EOF:
		goto eos
	case err != nil:
		return false
	case ch == '+' || ch == '-':
		if len(unscaled) > 0 || dot >= 0 {
			r.UnreadRune()
			goto eos
		}
	case ch == '.':
		if dot >= 0 {
			r.UnreadRune()
			goto eos
		}
		dot = len(unscaled)
		goto loop
	case ch >= '0' && ch <= '9':
	default:
		r.UnreadRune()
		goto eos
	}
	unscaled = append(unscaled, byte(ch))
	goto loop
eos:
	if dot >= 0 {
		x.scale = int32(len(unscaled) - dot)
	} else {
		x.scale = 0
	}
	_, ok := x.unscaled.SetString(cast.BytesToString(unscaled), 10)
	return ok
}

func exp10(x int32) *big.Int {
	if int(x) < len(exp10cache) {
		return &exp10cache[int(x)]
	}
	return new(big.Int).Exp(bigInt[10], big.NewInt(int64(x)), nil)
}

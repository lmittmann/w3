package internal

import (
	"math/big"

	"github.com/google/go-cmp/cmp"
)

// EquateErrors returns a cmp.Option that can be used to compare errors by their string.
func EquateErrors() cmp.Option {
	return cmp.FilterValues(areConcreteErrors, cmp.Comparer(compareErrors))
}

func areConcreteErrors(x, y interface{}) bool {
	_, ok1 := x.(error)
	_, ok2 := y.(error)
	return ok1 && ok2
}

func compareErrors(x, y interface{}) bool {
	xe := x.(error)
	ye := y.(error)
	return xe == nil && ye == nil || xe != nil && ye != nil && xe.Error() == ye.Error()
}

func OrCopy(vals ...*big.Int) *big.Int {
	var zero *big.Int
	for _, val := range vals {
		if val != zero {
			return new(big.Int).Set(val)
		}
	}
	return zero
}

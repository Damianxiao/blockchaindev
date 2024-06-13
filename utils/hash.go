package utils

import "math/big"

func BigToHash(bigInt *big.Int) Hash {
	return Hash(bigInt.Bytes())
}

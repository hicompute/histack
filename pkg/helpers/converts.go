package helper

import "math/big"

func StringToBigInt(s string) *big.Int {
	if s == "" {
		return big.NewInt(0)
	}
	bi, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return big.NewInt(0)
	}
	return bi
}

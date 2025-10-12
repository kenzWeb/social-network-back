package utils

import (
	"crypto/rand"
	"math/big"
)

func GenerateDigits(length int) string {
	digits := "0123456789"
	out := make([]byte, length)
	for i := 0; i < length; i++ {
		nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		out[i] = digits[nBig.Int64()]
	}
	return string(out)
}

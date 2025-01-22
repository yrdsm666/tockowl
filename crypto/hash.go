package crypto

import "crypto/sha256"

func Hash(input []byte) []byte {

	sum := sha256.Sum256(input)
	return sum[:]
}

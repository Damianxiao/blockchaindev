package sha3

import (
	"golang.org/x/crypto/sha3"
	"blockchaindev/utils/hash"
)

func Sha3(value []byte) hash.Hash{
	sha := sha3.NewLegacyLeccak256()
	return hash.BytesToHash(sha.Sum(value))
}
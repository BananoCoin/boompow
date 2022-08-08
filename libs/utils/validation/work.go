package validation

import (
	"encoding/binary"
	"encoding/hex"

	"golang.org/x/crypto/blake2b"
)

const (
	baseMaxUint64  = uint64(1<<64 - 1)
	baseDifficulty = baseMaxUint64 - uint64(0xfffffe0000000000)
)

func CalculateDifficulty(multiplier int64) uint64 {
	if multiplier < 0 {
		return baseMaxUint64 - (baseDifficulty * ((baseMaxUint64 - uint64(multiplier)) + 1))
	}

	if multiplier == 0 {
		multiplier = 1
	}

	return baseMaxUint64 - (baseDifficulty / uint64(multiplier))
}

func IsWorkValid(previous string, difficultyMultiplier int, w string) bool {
	difficult := CalculateDifficulty(int64(difficultyMultiplier))
	previousEnc, err := hex.DecodeString(previous)
	if err != nil {
		return false
	}
	wEnc, err := hex.DecodeString(w)
	if err != nil {
		return false
	}

	hash, err := blake2b.New(8, nil)
	if err != nil {
		return false
	}

	n := make([]byte, 8)
	copy(n, wEnc[:])

	reverse(n)
	hash.Write(n)
	hash.Write(previousEnc[:])

	return binary.LittleEndian.Uint64(hash.Sum(nil)) >= difficult
}

func reverse(v []byte) {
	// binary.LittleEndian.PutUint64(v, binary.BigEndian.Uint64(v))
	v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7] = v[7], v[6], v[5], v[4], v[3], v[2], v[1], v[0] // It's works. LOL
}

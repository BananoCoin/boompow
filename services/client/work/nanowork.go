package work

import (
	"encoding/hex"

	serializableModels "github.com/bbedward/boompow-ng/libs/models"
	"github.com/inkeliz/nanopow"
)

func WorkGenerate(item *serializableModels.ClientWorkRequest) (string, error) {
	decoded, err := hex.DecodeString(item.Hash)
	if err != nil {
		return "", err
	}
	work, err := nanopow.GenerateWork(decoded, nanopow.CalculateDifficulty(int64(item.DifficutlyMultiplier)))
	if err != nil {
		return "", err
	}

	return WorkToString(work), nil
}

func WorkToString(w nanopow.Work) string {
	n := make([]byte, 8)
	copy(n, w[:])

	reverse(n)

	return hex.EncodeToString(n)
}

func reverse(v []byte) {
	// binary.LittleEndian.PutUint64(v, binary.BigEndian.Uint64(v))
	v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7] = v[7], v[6], v[5], v[4], v[3], v[2], v[1], v[0] // It's works. LOL
}

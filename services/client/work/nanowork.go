package work

import (
	"encoding/hex"
	"errors"

	serializableModels "github.com/bbedward/boompow-ng/libs/models"
	"github.com/golang/glog"
	"github.com/inkeliz/nanopow"
)

func WorkGenerate(item *serializableModels.ClientWorkRequest) (string, error) {
	decoded, err := hex.DecodeString(item.Hash)
	if err != nil {
		return "", err
	}
	work, err := nanopow.GenerateWork(decoded, nanopow.CalculateDifficulty(int64(item.DifficultyMultiplier)))
	if err != nil {
		return "", err
	}

	if !nanopow.IsValid(decoded, nanopow.CalculateDifficulty(int64(item.DifficultyMultiplier)), work) {
		glog.Errorf("⚠️ Generated invalid work for %s", item.Hash)
		return "", errors.New("Invalid work")
	}
	return WorkToString(work), nil
}

func WorkToString(w nanopow.Work) string {
	n := make([]byte, 8)
	copy(n, w[:])

	return hex.EncodeToString(n)
}

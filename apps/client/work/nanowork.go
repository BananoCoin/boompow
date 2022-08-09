package work

import (
	"encoding/hex"
	"errors"

	serializableModels "github.com/bananocoin/boompow-next/libs/models"
	"github.com/bananocoin/boompow-next/libs/utils/validation"
	"github.com/golang/glog"
	"github.com/inkeliz/nanopow"
)

func WorkGenerate(item *serializableModels.ClientMessage) (string, error) {
	decoded, err := hex.DecodeString(item.Hash)
	if err != nil {
		return "", err
	}
	work, err := nanopow.GenerateWork(decoded, validation.CalculateDifficulty(int64(item.DifficultyMultiplier)))
	if err != nil {
		return "", err
	}

	if !nanopow.IsValid(decoded, validation.CalculateDifficulty(int64(item.DifficultyMultiplier)), work) {
		glog.Errorf("\n⚠️ Generated invalid work for %s", item.Hash)
		return "", errors.New("Invalid work")
	}
	return WorkToString(work), nil
}

func WorkToString(w nanopow.Work) string {
	n := make([]byte, 8)
	copy(n, w[:])

	return hex.EncodeToString(n)
}

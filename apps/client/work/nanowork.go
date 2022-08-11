package work

import (
	"encoding/hex"
	"errors"
	"runtime"

	serializableModels "github.com/bananocoin/boompow/libs/models"
	"github.com/bananocoin/boompow/libs/utils/validation"
	"github.com/golang/glog"
	"github.com/inkeliz/nanopow"
)

type WorkPool struct {
	Pool *nanopow.Pool
}

func NewWorkPool(gpuOnly bool) *WorkPool {
	pool := nanopow.NewPool()
	gpu, gpuErr := nanopow.NewWorkerGPU()
	if gpuErr == nil {
		pool.Workers = append(pool.Workers, gpu)
	}
	if !gpuOnly {
		threads := runtime.NumCPU()
		cpu, cpuErr := nanopow.NewWorkerCPUThread(uint64(threads))
		if cpuErr == nil {
			pool.Workers = append(pool.Workers, cpu)
		} else {
			panic("Unable to initialize work pool for CPU")
		}
	} else if gpuErr != nil {
		panic("No GPU found, but gpu-only was set")
	}

	return &WorkPool{
		Pool: pool,
	}
}

func (p *WorkPool) WorkGenerate(item *serializableModels.ClientMessage) (string, error) {
	decoded, err := hex.DecodeString(item.Hash)
	if err != nil {
		return "", err
	}
	work, err := p.Pool.GenerateWork(decoded, validation.CalculateDifficulty(int64(item.DifficultyMultiplier)))
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

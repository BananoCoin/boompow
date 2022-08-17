package work

import (
	"encoding/hex"
	"errors"
	"fmt"
	"runtime"

	"github.com/Inkeliz/go-opencl/opencl"
	serializableModels "github.com/bananocoin/boompow/libs/models"
	"github.com/bananocoin/boompow/libs/utils/validation"
	"github.com/bbedward/nanopow"
	"k8s.io/klog/v2"
)

type WorkPool struct {
	Pool *nanopow.Pool
}

func NewWorkPool(gpuOnly bool, devices []opencl.Device) *WorkPool {
	pool := nanopow.NewPool()
	for _, device := range devices {
		gpu, gpuErr := nanopow.NewWorkerGPU(device)
		if gpuErr == nil {
			pool.Workers = append(pool.Workers, gpu)
		} else {
			fmt.Printf("\n⚠️ Unable to use GPU %v", gpuErr)
		}
	}

	if gpuOnly && len(pool.Workers) == 0 {
		panic("Unable to initialize any GPUs, but gpu-only was set")
	} else if len(pool.Workers) == 0 {
		fmt.Printf("\n⚠️ Unable to initialize any GPUs, using CPU")
	}

	if !gpuOnly {
		threads := runtime.NumCPU()
		cpu, cpuErr := nanopow.NewWorkerCPUThread(uint64(threads))
		if cpuErr == nil {
			pool.Workers = append(pool.Workers, cpu)
		} else {
			panic(fmt.Sprintf("Unable to initialize work pool for CPU %v", cpuErr))
		}
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
		klog.Errorf("\n⚠️ Generated invalid work for %s", item.Hash)
		return "", errors.New("Invalid work")
	}
	return WorkToString(work), nil
}

func WorkToString(w nanopow.Work) string {
	n := make([]byte, 8)
	copy(n, w[:])

	return hex.EncodeToString(n)
}

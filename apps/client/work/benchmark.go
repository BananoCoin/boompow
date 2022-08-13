package work

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Inkeliz/go-opencl/opencl"
	"github.com/bananocoin/boompow/libs/models"
)

func RunBenchmark(nHashes int, difficultyMultiplier int, gpuOnly bool, devices []opencl.Device) {
	workPool := NewWorkPool(gpuOnly, devices)
	if difficultyMultiplier < 1 {
		difficultyMultiplier = 1
	}
	totalDelta := 0.0
	for i := 0; i < nHashes; i++ {
		bytes := make([]byte, 32)
		if _, err := rand.Read(bytes); err != nil {
			panic("Failed to generate hash")
		}

		fmt.Printf("\nRun %d", i+1)
		startT := time.Now()

		_, err := workPool.WorkGenerate(&models.ClientMessage{
			Hash:                 hex.EncodeToString(bytes),
			DifficultyMultiplier: difficultyMultiplier,
		})
		if err != nil {
			panic("Failed to generate work")
		}
		endT := time.Now()
		delta := endT.Sub(startT).Seconds()
		totalDelta += delta
		fmt.Printf("\nTook: %fs", delta)
	}
	fmt.Printf("\n\nAverage: %fs", totalDelta/float64(nHashes))
}

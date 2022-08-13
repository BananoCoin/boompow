package work

import (
	"context"
	"fmt"
	"time"

	"github.com/Inkeliz/go-opencl/opencl"
	"github.com/bananocoin/boompow/apps/client/models"
	"github.com/bananocoin/boompow/apps/client/websocket"
	serializableModels "github.com/bananocoin/boompow/libs/models"
)

type WorkProcessor struct {
	Queue *models.RandomAccessQueue
	// WorkQueueChan is where we write requests from the websocket
	WorkQueueChan chan *serializableModels.ClientMessage
	WSService     *websocket.WebsocketService
	WorkPool      *WorkPool
}

func NewWorkProcessor(ws *websocket.WebsocketService, gpuOnly bool, devices []opencl.Device) *WorkProcessor {
	wp := NewWorkPool(gpuOnly, devices)
	return &WorkProcessor{
		Queue:         models.NewRandomAccessQueue(),
		WorkQueueChan: make(chan *serializableModels.ClientMessage, 100),
		WSService:     ws,
		WorkPool:      wp,
	}
}

// RequestQueueWorker - is a worker that receives work requests directly from the websocket, adds them to the queue, and determines what should be worked on next
func (wp *WorkProcessor) StartRequestQueueWorker() {
	for range wp.WorkQueueChan {
		// Pop random unit of work from queue, begin computation
		workItem := wp.Queue.PopRandom()
		if workItem != nil {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			// Generate work with timeout
			ch := make(chan string)

			// Benchmark
			// startT := time.Now()

			go func() {
				result, err := wp.WorkPool.WorkGenerate(workItem)
				if err != nil {
					result = ""
				}

				select {
				default:
					ch <- result
				case <-ctx.Done():
				}
			}()

			select {
			case result := <-ch:
				if result != "" {
					// endT := time.Now()
					// delta := endT.Sub(startT).Seconds()
					// fmt.Printf("\nWork result: %s in %.2fs", result, delta)
					// Send result back to server
					clientWorkResult := serializableModels.ClientWorkResponse{
						RequestID: workItem.RequestID,
						Hash:      workItem.Hash,
						Result:    result,
					}
					wp.WSService.WS.WriteJSON(clientWorkResult)
				} else {
					fmt.Printf("\n❌ Error: generate work for %s\n", workItem.Hash)
				}
			case <-time.After(10 * time.Second):
				fmt.Printf("\n❌ Error: took longer than 10s to generate work for %s", workItem.Hash)
			}
		}
	}
}

// Start both workers
func (wp *WorkProcessor) StartAsync() {
	go wp.StartRequestQueueWorker()
}

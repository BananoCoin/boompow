package work

import (
	"context"
	"fmt"
	"time"

	"github.com/bananocoin/boompow/apps/client/models"
	"github.com/bananocoin/boompow/apps/client/websocket"
	serializableModels "github.com/bananocoin/boompow/libs/models"
)

type WorkProcessor struct {
	Queue *models.RandomAccessQueue
	// WorkQueueChan is where we write requests from the websocket
	WorkQueueChan chan *serializableModels.ClientMessage
	// WorkProcessChan is where we actually read from the queue and compute work
	WorkProcessChan chan bool
	WSService       *websocket.WebsocketService
	WorkPool        *WorkPool
}

func NewWorkProcessor(ws *websocket.WebsocketService, nWorkProcesses int, gpuOnly bool) *WorkProcessor {
	wp := NewWorkPool(gpuOnly)
	return &WorkProcessor{
		Queue:           models.NewRandomAccessQueue(),
		WorkQueueChan:   make(chan *serializableModels.ClientMessage, 100),
		WorkProcessChan: make(chan bool, nWorkProcesses),
		WSService:       ws,
		WorkPool:        wp,
	}
}

// RequestQueueWorker - is a worker that receives work requests directly from the websocket, adds them to the queue, and determines what should be worked on next
func (wp *WorkProcessor) StartRequestQueueWorker() {
	for c := range wp.WorkQueueChan {
		// If the backlog is too large, no-op
		if wp.Queue.Len() > 100 {
			continue
		}
		// Add to queue
		wp.Queue.Put(*c)
		// Add to work processing channel
		wp.WorkProcessChan <- true
	}
}

// WorkProcessor - is a worker that actually generates work and sends the result  back over the websocket
func (wp *WorkProcessor) StartWorkProcessor() {
	for range wp.WorkProcessChan {
		// Get random work item
		workItem := wp.Queue.PopRandom()
		if workItem != nil {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			// Generate work
			ch := make(chan string, 1)
			go func() {
				result, err := wp.WorkPool.WorkGenerate(workItem)
				if err != nil {
					result = ""
				}

				select {
				default:
					ch <- result
				case <-ctx.Done():
					fmt.Printf("\n❌ Error: took longer than 10s to generate work for %s", workItem.Hash)
				}
			}()

			select {
			case result := <-ch:
				if result != "" {
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
	go wp.StartWorkProcessor()
}

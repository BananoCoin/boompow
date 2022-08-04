package work

import (
	"fmt"

	serializableModels "github.com/bananocoin/boompow-next/libs/models"
	"github.com/bananocoin/boompow-next/services/client/models"
	"github.com/bananocoin/boompow-next/services/client/websocket"
)

type WorkProcessor struct {
	Queue *models.RandomAccessQueue
	// WorkQueueChan is where we write requests from the websocket
	WorkQueueChan chan *serializableModels.ClientRequest
	// WorkProcessChan is where we actually read from the queue and compute work
	WorkProcessChan chan bool
	WSService       *websocket.WebsocketService
}

func NewWorkProcessor(ws *websocket.WebsocketService, nWorkProcesses int) *WorkProcessor {
	return &WorkProcessor{
		Queue:           models.NewRandomAccessQueue(),
		WorkQueueChan:   make(chan *serializableModels.ClientRequest, 100),
		WorkProcessChan: make(chan bool, nWorkProcesses),
		WSService:       ws,
	}
}

// RequestQueueWorker - is a worker that receives work requests directly from the websocket, adds them to the queue, and determines what should be worked on next
func (wp *WorkProcessor) StartRequestQueueWorker() {
	for c := range wp.WorkQueueChan {
		// If the backlog is too large, no-op
		if len(wp.Queue.Hashes) > 100 {
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
			// Generate work
			result, err := WorkGenerate(workItem)
			if err != nil {
				fmt.Printf("\n❌ Error: generate work for %s\n", workItem.Hash)
			}
			// Send result back to server
			clientWorkResult := serializableModels.ClientWorkResponse{
				RequestID: workItem.RequestID,
				Hash:      workItem.Hash,
				Result:    result,
			}
			wp.WSService.WS.WriteJSON(clientWorkResult)
		}
	}
}

// Start both workers
func (wp *WorkProcessor) StartAsync() {
	go wp.StartRequestQueueWorker()
	go wp.StartWorkProcessor()
}

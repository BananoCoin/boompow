package work

import (
	"fmt"

	serializableModels "github.com/bbedward/boompow-ng/libs/models"
	"github.com/bbedward/boompow-ng/services/client/models"
	"github.com/bbedward/boompow-ng/services/client/websocket"
)

type WorkProcessor struct {
	Queue           *models.RandomAccessQueue
	workProcessChan chan bool
	ws              *websocket.RecConn
}

func NewWorkProcessor(ws *websocket.RecConn, workProcessChan chan bool) *WorkProcessor {
	return &WorkProcessor{
		Queue:           models.NewRandomAccessQueue(),
		workProcessChan: workProcessChan,
		ws:              ws,
	}
}

// RequestQueueWorker - is a worker that receives work requests directly from the websocket, adds them to the queue, and determines what should be worked on next
func (wp *WorkProcessor) StartRequestQueueWorker(requestChan <-chan *serializableModels.ClientRequest) {
	for c := range requestChan {
		// If the backlog is too large, no-op
		if len(wp.Queue.Hashes) > 100 {
			continue
		}
		// Add to queue
		wp.Queue.Put(*c)
		// Add to work processing channel
		wp.workProcessChan <- true
	}
}

// WorkProcessor - is a worker that actually generates work and sends the result  back over the websocket
func (wp *WorkProcessor) StartWorkProcessor(workProcessChan <-chan bool) {
	for range workProcessChan {
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
			wp.ws.WriteJSON(clientWorkResult)
		}
	}
}

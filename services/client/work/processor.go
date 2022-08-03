package work

import (
	"fmt"

	serializableModels "github.com/bbedward/boompow-ng/libs/models"
	"github.com/bbedward/boompow-ng/services/client/models"
	"github.com/bbedward/boompow-ng/services/client/websocket"
)

type WorkProcessor struct {
	queue           *models.RandomAccessQueue
	workProcessChan chan bool
	ws              *websocket.RecConn
}

func NewWorkProcessor(ws *websocket.RecConn, workProcessChan chan bool) *WorkProcessor {
	return &WorkProcessor{
		queue:           models.NewRandomAccessQueue(),
		workProcessChan: workProcessChan,
		ws:              ws,
	}
}

// RequestQueueWorker - is a worker that receives work requests directly from the websocket, adds them to the queue, and determines what should be worked on next
func (wp *WorkProcessor) StartRequestQueueWorker(requestChan <-chan *serializableModels.ClientWorkRequest) {
	for c := range requestChan {
		// If the backlog is too large, no-op
		if len(wp.queue.Hashes) > 100 {
			continue
		}
		// Add to queue
		wp.queue.Put(*c)
		// Add to work processing channel
		wp.workProcessChan <- true
	}
}

// WorkProcessor - is a worker that actually generates work and sends the result  back over the websocket
func (wp *WorkProcessor) StartWorkProcessor(workProcessChan <-chan bool) {
	for range workProcessChan {
		// Get random work item
		workItem := wp.queue.PopRandom()
		if workItem != nil {
			// Generate work
			result, err := WorkGenerate(workItem)
			if err != nil {
				fmt.Printf("\n❌ Error: generate work for %s\n", workItem.Hash)
			}
			// Send result back to server
			clientWorkResult := serializableModels.ClientWorkResponse{
				Hash:   workItem.Hash,
				Result: result,
			}
			wp.ws.WriteJSON(clientWorkResult)
		}
	}
}

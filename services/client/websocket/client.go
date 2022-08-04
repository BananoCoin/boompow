package websocket

import (
	"context"
	"fmt"
	"net/http"
	"time"

	serializableModels "github.com/bbedward/boompow-ng/libs/models"
	"github.com/bbedward/boompow-ng/services/client/models"
)

var WS *RecConn
var AuthToken string

func CreateWS() {
	WS = &RecConn{}
}

func UpdateAuthToken(authToken string) {
	AuthToken = authToken
	WS.setReqHeader(http.Header{
		"Authorization": {AuthToken},
	})
}

func StartWSClient(ctx context.Context, requestChan *chan *serializableModels.ClientRequest, queue *models.RandomAccessQueue) {
	// Start the websocket connection
	WS.Dial("ws://localhost:8080/ws/worker", http.Header{
		"Authorization": {AuthToken},
	})

	for {
		select {
		case <-ctx.Done():
			go WS.Close()
			fmt.Printf("Websocket closed %s", WS.GetURL())
			return
		default:
			if !WS.IsConnected() {
				fmt.Printf("Websocket disconnected %s", WS.GetURL())
				time.Sleep(2 * time.Second)
				continue
			}

			var serverMsg serializableModels.ClientRequest
			err := WS.ReadJSON(&serverMsg)
			if err != nil {
				fmt.Printf("Error: ReadJSON %s", WS.GetURL())
				continue
			}

			// Determine type of message
			if serverMsg.RequestType == "work_generate" {
				fmt.Printf("\n🦋 Received work request %s with difficulty %dx", serverMsg.Hash, serverMsg.DifficultyMultiplier)

				if len(serverMsg.Hash) != 64 {
					fmt.Printf("\nReceived invalid hash, skipping")
				}

				// Queue
				*requestChan <- &serverMsg
			} else if serverMsg.RequestType == "work_cancel" {
				// Delete pending work from queue
				// ! TODO - can we cancel currently runing work calculations?
				var workCancelCmd serializableModels.ClientRequest
				queue.Delete(workCancelCmd.Hash)
			} else {
				fmt.Printf("\n🦋 Received unknown message %s\n", serverMsg.RequestType)
			}
		}
	}
}

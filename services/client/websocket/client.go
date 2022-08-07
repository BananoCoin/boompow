package websocket

import (
	"context"
	"fmt"
	"net/http"
	"time"

	serializableModels "github.com/bananocoin/boompow-next/libs/models"
	"github.com/bananocoin/boompow-next/services/client/models"
)

type WebsocketService struct {
	WS        *RecConn
	AuthToken string
}

func NewWebsockerService() *WebsocketService {
	return &WebsocketService{
		WS: &RecConn{},
	}
}

func (ws *WebsocketService) SetAuthToken(authToken string) {
	ws.AuthToken = authToken
	ws.WS.setReqHeader(http.Header{
		"Authorization": {ws.AuthToken},
	})
}

func (ws *WebsocketService) StartWSClient(ctx context.Context, workQueueChan chan *serializableModels.ClientMessage, queue *models.RandomAccessQueue) {
	if ws.AuthToken == "" {
		panic("Tired to start websocket client without auth token")
	}
	// Start the websocket connection
	ws.WS.Dial("ws://localhost:8080/ws/worker", http.Header{
		"Authorization": {ws.AuthToken},
	})

	for {
		select {
		case <-ctx.Done():
			go ws.WS.Close()
			fmt.Printf("Websocket closed %s", ws.WS.GetURL())
			return
		default:
			if !ws.WS.IsConnected() {
				fmt.Printf("Websocket disconnected %s", ws.WS.GetURL())
				time.Sleep(2 * time.Second)
				continue
			}

			var serverMsg serializableModels.ClientMessage
			err := ws.WS.ReadJSON(&serverMsg)
			if err != nil {
				fmt.Printf("Error: ReadJSON %s", ws.WS.GetURL())
				continue
			}

			// Determine type of message
			if serverMsg.MessageType == serializableModels.WorkGenerate {
				fmt.Printf("\n🦋 Received work request %s with difficulty %dx", serverMsg.Hash, serverMsg.DifficultyMultiplier)

				if len(serverMsg.Hash) != 64 {
					fmt.Printf("\nReceived invalid hash, skipping")
				}

				// Queue
				workQueueChan <- &serverMsg
			} else if serverMsg.MessageType == serializableModels.WorkCancel {
				// Delete pending work from queue
				// ! TODO - can we cancel currently runing work calculations?
				var workCancelCmd serializableModels.ClientMessage
				queue.Delete(workCancelCmd.Hash)
			} else if serverMsg.MessageType == serializableModels.BlockAwarded {
				fmt.Printf("\n💰 Received block awarded %s, %f, %f", serverMsg.Hash, serverMsg.PercentOfPool, serverMsg.EstimatedAward)
			} else {
				fmt.Printf("\n🦋 Received unknown message %s\n", serverMsg.MessageType)
			}
		}
	}
}

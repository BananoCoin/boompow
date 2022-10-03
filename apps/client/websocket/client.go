package websocket

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bananocoin/boompow/apps/client/models"
	serializableModels "github.com/bananocoin/boompow/libs/models"
)

type WebsocketService struct {
	WS            *RecConn
	AuthToken     string
	URL           string
	maxDifficulty int
	minDifficulty int
	skipPrecache  bool
}

func NewWebsocketService(url string, maxDifficulty int, minDifficulty int, skipPrecache bool) *WebsocketService {
	return &WebsocketService{
		WS:            &RecConn{},
		URL:           url,
		maxDifficulty: maxDifficulty,
		minDifficulty: minDifficulty,
		skipPrecache:  skipPrecache,
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
	ws.WS.Dial(ws.URL, http.Header{
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
				if serverMsg.DifficultyMultiplier > ws.maxDifficulty {
					fmt.Printf("\n😒 Ignoring work request %s with difficulty %dx above our max %dx", serverMsg.Hash, serverMsg.DifficultyMultiplier, ws.maxDifficulty)
					continue
				}
				if serverMsg.DifficultyMultiplier < ws.minDifficulty {
					fmt.Printf("\n😒 Ignoring work request %s with difficulty %dx below our min %dx", serverMsg.Hash, serverMsg.DifficultyMultiplier, ws.minDifficulty)
					continue
				}

				if ws.skipPrecache && serverMsg.Precache {
					fmt.Printf("\n😒 Ignoring precache request %s", serverMsg.Hash)
					continue
				}

				fmt.Printf("\n🦋 Received work request %s with difficulty %dx", serverMsg.Hash, serverMsg.DifficultyMultiplier)

				if len(serverMsg.Hash) != 64 {
					fmt.Printf("\nReceived invalid hash, skipping")
					continue
				}

				// If the backlog is too large, no-op
				if queue.Len() > 99 {
					fmt.Printf("\nBacklog is too large, skipping hash %s", serverMsg.Hash)
					continue
				}

				// Queue this work
				queue.Put(serverMsg)

				// Signal channel that we have work to do
				workQueueChan <- &serverMsg
			} else if serverMsg.MessageType == serializableModels.WorkCancel {
				// Delete pending work from queue
				// ! TODO - can we cancel currently runing work calculations?
				queue.Delete(serverMsg.Hash)
			} else if serverMsg.MessageType == serializableModels.BlockAwarded {
				fmt.Printf("\n💰 Received block awarded %s", serverMsg.Hash)
				fmt.Printf("\n💰 Your current estimated next payout is %f%% or %f BAN", serverMsg.PercentOfPool, serverMsg.EstimatedAward)
			} else {
				fmt.Printf("\n🦋 Received unknown message %s\n", serverMsg.MessageType)
			}
		}
	}
}

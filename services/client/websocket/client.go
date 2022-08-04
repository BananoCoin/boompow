package websocket

import (
	"context"
	"fmt"
	"net/http"
	"time"

	serializableModels "github.com/bbedward/boompow-ng/libs/models"
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

func StartWSClient(ctx context.Context, requestChan *chan *serializableModels.ClientWorkRequest) {
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

			var ClientWorkRequest serializableModels.ClientWorkRequest
			err := WS.ReadJSON(&ClientWorkRequest)
			if err != nil {
				fmt.Printf("Error: ReadJSON %s", WS.GetURL())
				continue
			}
			fmt.Printf("\n🦋 Received work request %s with difficulty %dx\n", ClientWorkRequest.Hash, ClientWorkRequest.DifficultyMultiplier)

			if len(ClientWorkRequest.Hash) != 64 {
				fmt.Printf("\nReceived invalid hash, skipping\n")
			}

			// Queue
			*requestChan <- &ClientWorkRequest
		}
	}
}

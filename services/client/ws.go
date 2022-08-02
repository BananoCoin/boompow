package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	serializableModels "github.com/bbedward/boompow-ng/libs/models"
	"github.com/inkeliz/nanopow"
	"github.com/recws-org/recws"
)

func StartWSClient(ctx context.Context) {
	fmt.Println("🚀 Starting websocket client...")

	// Start the websocket connection
	ws := recws.RecConn{}
	ws.Dial("ws://localhost:8080/ws/worker", nil)

	for {
		select {
		case <-ctx.Done():
			go ws.Close()
			fmt.Printf("Websocket closed %s", ws.GetURL())
			return
		default:
			if !ws.IsConnected() {
				fmt.Printf("Websocket disconnected %s", ws.GetURL())
				time.Sleep(2 * time.Second)
				continue
			}

			var ClientWorkRequest serializableModels.ClientWorkRequest
			err := ws.ReadJSON(&ClientWorkRequest)
			if err != nil {
				fmt.Printf("Error: ReadJSON %s", ws.GetURL())
				continue
			}
			fmt.Printf("Received work request %s", ClientWorkRequest.Hash)

			// Write response
			decoded, err := hex.DecodeString("782E7799FBFFBD13A5133DB42FCB64D1EBCAEF85E219FE37627B4660C4AF2A4A")
			work, err := nanopow.GenerateWork(decoded, nanopow.V1BaseDifficult)
			if err != nil {
				fmt.Printf("Error: GenerateWork")
				continue
			}

			ws.WriteJSON(serializableModels.ClientWorkResponse{Hash: ClientWorkRequest.Hash, Result: WorkToString(work)})
		}
	}
}

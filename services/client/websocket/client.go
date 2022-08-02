package websocket

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	serializableModels "github.com/bbedward/boompow-ng/libs/models"
	"github.com/inkeliz/nanopow"
)

var ws *RecConn
var AuthToken string

func UpdateAuthToken(authToken string) {
	AuthToken = authToken
	ws.setReqHeader(http.Header{
		"Authorization": {AuthToken},
	})
}

func StartWSClient(ctx context.Context) {
	// Start the websocket connection
	ws = &RecConn{}
	ws.Dial("ws://localhost:8080/ws/worker", http.Header{
		"Authorization": {AuthToken},
	})

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
			fmt.Printf("\n🦋 Received work request %s with difficulty %dx\n", ClientWorkRequest.Hash, ClientWorkRequest.DifficutlyMultiplier)

			if len(ClientWorkRequest.Hash) != 64 {
				fmt.Printf("\nReceived invalid hash, skipping\n")
			}

			// Write response
			decoded, err := hex.DecodeString(ClientWorkRequest.Hash)
			work, err := nanopow.GenerateWork(decoded, nanopow.CalculateDifficulty(int64(ClientWorkRequest.DifficutlyMultiplier)))
			if err != nil {
				fmt.Printf("Error: GenerateWork")
				continue
			}

			ws.WriteJSON(serializableModels.ClientWorkResponse{Hash: ClientWorkRequest.Hash, Result: WorkToString(work)})
		}
	}
}

func WorkToString(w nanopow.Work) string {
	n := make([]byte, 8)
	copy(n, w[:])

	reverse(n)

	return hex.EncodeToString(n)
}

func reverse(v []byte) {
	// binary.LittleEndian.PutUint64(v, binary.BigEndian.Uint64(v))
	v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7] = v[7], v[6], v[5], v[4], v[3], v[2], v[1], v[0] // It's works. LOL
}

package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	serializableModels "github.com/bbedward/boompow-ng/libs/models"
	"github.com/inkeliz/nanopow"
	"github.com/recws-org/recws"
)

func usage() {
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "INFO")
	flag.Set("v", "2")
	flag.Parse()
}

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func SetupCloseHandler(ctx context.Context, cancel context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		<-c
		fmt.Print("👋 Exiting...\n")
		cancel()
		os.Exit(0)
	}()
}

func main() {
	// Start the websocket connection
	ctx, cancel := context.WithCancel(context.Background())
	ws := recws.RecConn{}
	ws.Dial("ws://localhost:8080/ws/worker", nil)

	// Handle interrupts gracefully
	SetupCloseHandler(ctx, cancel)

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

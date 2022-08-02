package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	serializableModels "github.com/bbedward/boompow-ng/libs/models"
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
			ws.WriteJSON(serializableModels.ClientWorkResponse{Hash: ClientWorkRequest.Hash, Result: "hello world"})
		}
	}
}

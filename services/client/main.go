package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
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

func main() {
	// Start the websocket connection
	ctx, cancel := context.WithCancel(context.Background())
	ws := recws.RecConn{}
	ws.Dial("ws://localhost:8080/ws/worker", nil)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer func() {
		signal.Stop(sigc)
		cancel()
	}()

	for {
		select {
		case <-sigc:
			cancel()
			return
		case <-ctx.Done():
			go ws.Close()
			glog.Infof("Websocket closed %s", ws.GetURL())
			return
		default:
			if !ws.IsConnected() {
				glog.Infof("Websocket disconnected %s", ws.GetURL())
				time.Sleep(2 * time.Second)
				continue
			}

			_, msg, err := ws.ReadMessage()
			if err != nil {
				glog.Infof("Error: ReadJSON %s", ws.GetURL())
				continue
			}

			// Trigger callback

			glog.Infof("Received message %s", string(msg))
		}
	}
}

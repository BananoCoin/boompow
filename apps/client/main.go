package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bananocoin/boompow-next/apps/client/gql"
	"github.com/bananocoin/boompow-next/apps/client/websocket"
	"github.com/bananocoin/boompow-next/apps/client/work"
	"github.com/bananocoin/boompow-next/libs/utils/validation"
	"github.com/go-co-op/gocron"
	"github.com/mbndr/figlet4go"
	"golang.org/x/term"
)

// For pretty text
func printBanner() {
	ascii := figlet4go.NewAsciiRender()
	options := figlet4go.NewRenderOptions()
	color, _ := figlet4go.NewTrueColorFromHexString("44B542")
	options.FontColor = []figlet4go.Color{
		color,
	}

	renderStr, _ := ascii.RenderOpts("BoomPOW", options)
	fmt.Print(renderStr)
}

func usage() {
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "INFO")
	flag.Set("v", "2")
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

// Represents the number of simultaneous work calculations we will run
var NConcurrentWorkers int

// Instance of websocket service
var WSService *websocket.WebsocketService

func main() {
	// Parse flags
	threadCount := flag.Int("thread-count", 1, "The maximum number of concurrent work requests to process")
	flag.Parse()
	NConcurrentWorkers = *threadCount

	printBanner()
	gql.InitGQLClient()

	// Define context
	ctx, cancel := context.WithCancel(context.Background())

	// Handle interrupts gracefully
	SetupCloseHandler(ctx, cancel)

	// Create WS Service
	WSService = websocket.NewWebsockerService()

	// Loop to get username and password and login
	for {
		// Get username/password
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("➡️ Enter Email: ")
		email, err := reader.ReadString('\n')

		if err != nil {
			fmt.Printf("\n⚠️ Error reading email")
			continue
		}

		email = strings.TrimSpace(email)

		if !validation.IsValidEmail(email) {
			fmt.Printf("\n⚠️ Invalid email\n\n")
			continue
		}

		fmt.Print("➡️ Enter Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))

		if err != nil {
			fmt.Printf("\n⚠️ Error reading password")
			continue
		}

		password := strings.TrimSpace(string(bytePassword))

		// Login
		fmt.Printf("\n\n🔒 Logging in...")
		resp, gqlErr := gql.Login(ctx, email, password)
		if gqlErr == gql.InvalidUsernamePasssword {
			fmt.Printf("\n❌ Invalid email or password\n\n")
			continue
		} else if gqlErr == gql.ServerError {
			fmt.Printf("\n💥 Error reaching server, try again later\n")
			os.Exit(1)
		}
		fmt.Printf("\n\n🔓 Successfully logged in as %s\n\n", email)
		WSService.SetAuthToken(resp.Login.Token)
		break
	}

	// Setup a cron job to auto-update auth tokens
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every(1).Hour().Do(func() {
		authToken, err := gql.RefreshToken(ctx, WSService.AuthToken)
		if err == nil {
			WSService.SetAuthToken(authToken)
		}
	})
	scheduler.StartAt(time.Now().Add(time.Hour))
	scheduler.StartAsync()

	fmt.Printf("\n🚀 Initiating connection to BoomPOW...")

	// Create work processor
	workProcessor := work.NewWorkProcessor(WSService, NConcurrentWorkers)
	workProcessor.StartAsync()

	WSService.StartWSClient(ctx, workProcessor.WorkQueueChan, workProcessor.Queue)
}

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

	serializableModels "github.com/bbedward/boompow-ng/libs/models"
	"github.com/bbedward/boompow-ng/libs/utils/validation"
	"github.com/bbedward/boompow-ng/services/client/gql"
	"github.com/bbedward/boompow-ng/services/client/websocket"
	"github.com/bbedward/boompow-ng/services/client/work"
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
		websocket.AuthToken = resp.Login.Token
		break
	}

	// Setup a cron job to auto-update auth tokens
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every(1).Hour().Do(func() {
		authToken, err := gql.RefreshToken(ctx, websocket.AuthToken)
		if err == nil {
			websocket.UpdateAuthToken(authToken)
		}
	})
	scheduler.StartAsync()

	fmt.Printf("\n🚀 Initiating connection to BoomPOW...")

	// Create channel to receive work requests
	// This channel is larger than the actual channel that processes results
	workRequestChannel := make(chan *serializableModels.ClientWorkRequest, 100)
	// Create channel to generate work
	workGenerateChannel := make(chan bool, NConcurrentWorkers)

	// Create work processor
	// ! TODO - clean this up
	// It's a little wonky to the way creating the websocket and passing this stuff around works
	// Being able to access the authToken in another go-routine made it a little tricky
	websocket.CreateWS()
	workProcessor := work.NewWorkProcessor(websocket.WS, workGenerateChannel)
	go workProcessor.StartRequestQueueWorker(workRequestChannel)
	go workProcessor.StartWorkProcessor(workGenerateChannel)

	websocket.StartWSClient(ctx, &workRequestChannel)
}

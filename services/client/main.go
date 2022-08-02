package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bbedward/boompow-ng/libs/utils/validation"
	"github.com/bbedward/boompow-ng/services/client/gql"
	"github.com/inkeliz/nanopow"
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
		fmt.Println(resp.Login.Token)
		break
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

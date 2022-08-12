package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Inkeliz/go-opencl/opencl"
	"github.com/bananocoin/boompow/apps/client/gql"
	"github.com/bananocoin/boompow/apps/client/websocket"
	"github.com/bananocoin/boompow/apps/client/work"
	"github.com/bananocoin/boompow/libs/utils/validation"
	"github.com/go-co-op/gocron"
	"github.com/mbndr/figlet4go"
	"golang.org/x/term"
)

// Variables
var GraphQLURL = "http://localhost:8080/graphql"
var WSUrl = "ws://localhost:8080/ws/worker"
var Version = "dev"

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

// Determine GPU info
type gpuINFO struct {
	platformName  string
	vendor        string
	driverVersion string
	device        opencl.Device
}

func getGPUInfo() ([]*gpuINFO, error) {
	ret := []*gpuINFO{}
	platforms, err := opencl.GetPlatforms()
	if err != nil {
		return nil, err
	}

	var platform opencl.Platform
	var name string
	for _, curPlatform := range platforms {
		err = curPlatform.GetInfo(opencl.PlatformName, &name)
		if err != nil {
			return nil, err
		}

		var devices []opencl.Device
		devices, err = curPlatform.GetDevices(opencl.DeviceTypeAll)
		if err != nil {
			return nil, err
		}

		for _, device := range devices {
			var available bool
			err = device.GetInfo(opencl.DeviceAvailable, &available)
			if err == nil && available {
				platform = curPlatform
				device = devices[0]
			}

			var platformName string
			err := platform.GetInfo(opencl.PlatformName, &platformName)
			if err != nil {
				continue
			}

			var vendor string
			err = device.GetInfo(opencl.DeviceVendor, &vendor)
			if err != nil {
				continue
			}

			var driverVersion string
			err = device.GetInfo(opencl.DriverVersion, &driverVersion)
			if err != nil {
				continue
			}

			ret = append(ret, &gpuINFO{
				platformName:  platformName,
				vendor:        vendor,
				driverVersion: driverVersion,
				device:        device,
			})
		}
	}

	if len(ret) > 0 {
		return ret, nil
	}

	return nil, errors.New("No GPU found")
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
	gpuOnly := flag.Bool("gpu-only", false, "If set, will only run work on GPU (otherwise, both CPU and GPU)")
	maxDifficulty := flag.Int("max-difficulty", 128, "The maximum work difficulty to compute, less than this will be ignored")
	benchmark := flag.Int("benchmark", 0, "Run a benchmark for the given number of random hashes")
	benchmarkDifficulty := flag.Int("benchmark-difficulty", 64, "The difficulty multiplier for the benchmark")
	argEmail := flag.String("email", "", "The email (username) to use for the worker (optional)")
	argPassword := flag.String("password", "", "The password to use for the worker (optional)")
	registerProvider := flag.Bool("register-provider", false, "Register to be a provider (optional)")
	registerService := flag.Bool("register-service", false, "Register to be a service/work requester (optional)")
	resendConfirmationEmail := flag.Bool("resend-confirmation-email", false, "Resend the confirmation email (optional)")
	generateServiceToken := flag.Bool("generate-service-token", false, "Generate a service token (optional)")
	version := flag.Bool("version", false, "Display the version")
	flag.Parse()

	if *version {
		fmt.Printf("BoomPOW version: %s\n", Version)
		os.Exit(0)
	}

	printBanner()

	gpuInfo, err := getGPUInfo()
	if err != nil {
		fmt.Printf("\n🚨 No GPU Found!")
		fmt.Printf("\nThis error is safe to ignore if you intended to generate PoW on CPU only")
		fmt.Printf("\nOtherwise you may want to check your GPU drivers and ensure it is properly installed, as well as ensure your device supports OpenCL 2.0\n\n")
	} else {
		fmt.Printf("\n⚡ Using GPU")
		fmt.Printf("\nPlatform: %s", gpuInfo[0].platformName)
		fmt.Printf("\nVendor: %s", gpuInfo[0].vendor)
		fmt.Printf("\nDriver: %s\n", gpuInfo[0].driverVersion)
	}
	if *gpuOnly {
		fmt.Printf("\nOnly using GPU for work_generate...\n\n")
	} else {
		fmt.Printf("\nUsing GPU+CPU for work_generate...\n\n")
	}

	os.Exit(0)

	// Check benchmark
	if *benchmark > 0 {
		work.RunBenchmark(*benchmark, *benchmarkDifficulty, *gpuOnly)
		os.Exit(0)
	}

	// Define context
	ctx, cancel := context.WithCancel(context.Background())
	gql.InitGQLClient(GraphQLURL)

	// Handle interrupts gracefully
	SetupCloseHandler(ctx, cancel)

	// Short circuit for registration
	// These should be on a website eventually
	if *resendConfirmationEmail {
		// Loop to get credentials
		for {
			// Get email
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("➡️ Enter Email: ")
			rawEmail, err := reader.ReadString('\n')

			if err != nil {
				fmt.Printf("\n⚠️ Error reading email")
				continue
			}

			email := strings.TrimSpace(rawEmail)

			if !validation.IsValidEmail(email) {
				fmt.Printf("\n⚠️ Invalid email\n\n")
				continue
			}

			resp, err := gql.ResendConfirmationEmail(ctx, email)
			if err != nil {
				fmt.Printf("\n⚠️ Error resending confirmation email: %v\n\n", err)
				os.Exit(1)
			}
			if resp.ResendConfirmationEmail {
				fmt.Printf("\n\n✅ Successfully resent confirmation to %s, check your email for a confirmation link. Once confirmed you can login to start contributing.\n\n", email)
				os.Exit(0)
			}
		}
	}
	if *registerProvider {
		fmt.Printf("\n❤️ Register to be a BoomPoW Contributor!\nYou will receive daily rewards for your work that is accepted!\n")
		// Loop to get credentials
		for {
			// Get username/password
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("➡️ Enter Email: ")
			rawEmail, err := reader.ReadString('\n')

			if err != nil {
				fmt.Printf("\n⚠️ Error reading email")
				continue
			}

			email := strings.TrimSpace(rawEmail)

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

			err = validation.ValidatePassword(password)
			if err != nil {
				fmt.Printf("\n⚠️ Invalid password %v\n\n", err)
				continue
			}

			fmt.Print("\n➡️ Confirm Password: ")
			bytePasswordC, err := term.ReadPassword(int(syscall.Stdin))

			if err != nil {
				fmt.Printf("\n⚠️ Error reading password")
				continue
			}
			passwordC := strings.TrimSpace(string(bytePasswordC))

			if password != passwordC {
				fmt.Printf("\n⚠️ Passwords do not match\n\n")
				continue
			}

			fmt.Print("\n➡️ Enter Ban Address: ")
			rawBanAddress, err := reader.ReadString('\n')

			if err != nil {
				fmt.Printf("\n⚠️ Error reading ban address")
				continue
			}

			banAddress := strings.TrimSpace(rawBanAddress)

			if !validation.ValidateAddress(banAddress) {
				fmt.Printf("\n⚠️ Invalid banano address\n\n")
				continue
			}

			// Register
			fmt.Printf("\n\n⏱️ Registering...")
			_, err = gql.RegisterProvider(ctx, email, password, banAddress)
			if err != nil {
				fmt.Printf("\n⚠️ Error registering: %v\n\n", err)
				os.Exit(1)
			}
			fmt.Printf("\n\n✅ Successfully registered %s, check your email for a confirmation link. Once confirmed you can login to start contributing.\n\n", email)
			os.Exit(0)
		}
	}

	if *registerService {
		fmt.Printf("\n❤️ Register to gain access to the BoomPoW System!\nOnce accepted, you can request PoW from the BoomPoW API.\n")
		fmt.Printf("\n\n☢️ NOTE: You will be prompted to enter a service name and a service website")
		fmt.Printf("\nBoomPoW is free to use for services that utilize the BANANO or Nano networks")
		fmt.Printf("\nIt is generally NOT available for individual users, if you think you are an exception to this rule then you should DM @bbedward#9246 on Discord")
		fmt.Printf("\nWebsite must be live to be approved, a \"coming soon\" page is acceptable as long as it contains a description of what your service is")
		fmt.Printf("\nAbuse of the system will result in your account being banned\n\n")

		// Loop to get credentials
		for {
			// Get username/password
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("➡️ Enter Email: ")
			rawEmail, err := reader.ReadString('\n')

			if err != nil {
				fmt.Printf("\n⚠️ Error reading email")
				continue
			}

			email := strings.TrimSpace(rawEmail)

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

			err = validation.ValidatePassword(password)
			if err != nil {
				fmt.Printf("\n⚠️ Invalid password %v\n\n", err)
				continue
			}

			fmt.Print("\n➡️ Confirm Password: ")
			bytePasswordC, err := term.ReadPassword(int(syscall.Stdin))

			if err != nil {
				fmt.Printf("\n⚠️ Error reading password")
				continue
			}
			passwordC := strings.TrimSpace(string(bytePasswordC))

			if password != passwordC {
				fmt.Printf("\n⚠️ Passwords do not match\n\n")
				continue
			}

			fmt.Print("\n➡️ Enter Service Name: ")
			rawServiceName, err := reader.ReadString('\n')

			if err != nil {
				fmt.Printf("\n⚠️ Error reading service name")
				continue
			}

			serviceName := strings.TrimSpace(rawServiceName)

			if len(serviceName) < 3 {
				fmt.Printf("\n⚠️ Service name must be at least 3 characters long\n\n")
				continue
			}

			fmt.Print("\n➡️ Enter Service Website: ")
			rawServiceWebsite, err := reader.ReadString('\n')

			if err != nil {
				fmt.Printf("\n⚠️ Error reading service website")
				continue
			}

			serviceWebsite := strings.TrimSpace(rawServiceWebsite)

			if !strings.HasPrefix(serviceWebsite, "http://") && !strings.HasPrefix(serviceWebsite, "https://") {
				serviceWebsite = "https://" + serviceWebsite
			}

			_, err = url.ParseRequestURI(serviceWebsite)

			if err != nil {
				fmt.Printf("\n⚠️ Invalid website\n\n")
				continue
			}

			if strings.HasPrefix(serviceWebsite, "http://") {
				fmt.Printf("\n⚠️ Only https websites are supported\n\n")
				continue
			}

			// Register
			fmt.Printf("\n\n⏱️ Registering...")
			_, err = gql.RegisterService(ctx, email, password, serviceName, serviceWebsite)
			if err != nil {
				fmt.Printf("\n⚠️ Error registering: %v\n\n", err)
				os.Exit(1)
			}
			fmt.Printf("\n\n✅ Successfully registered %s, check your email for a confirmation link.\n", email)
			fmt.Printf("\nAfter you confirm your email address, you will receive a follow up email once you are approved to use the service with instructions on getting started.\n\n")
			os.Exit(0)
		}
	}

	if *generateServiceToken {
		// Loop to get username and password and login
		for {
			// Get username/password
			reader := bufio.NewReader(os.Stdin)

			var email string

			if *argEmail == "" {
				fmt.Print("➡️ Enter Email: ")
				rawEmail, err := reader.ReadString('\n')

				if err != nil {
					fmt.Printf("\n⚠️ Error reading email")
					continue
				}

				email = strings.TrimSpace(rawEmail)

				if !validation.IsValidEmail(email) {
					fmt.Printf("\n⚠️ Invalid email\n\n")
					continue
				}
			} else {
				if !validation.IsValidEmail(*argEmail) {
					fmt.Printf("\n⚠️ Invalid email\n\n")
					os.Exit(1)
				}
				email = *argEmail
			}

			var password string

			if *argPassword == "" {
				fmt.Print("➡️ Enter Password: ")
				bytePassword, err := term.ReadPassword(int(syscall.Stdin))

				if err != nil {
					fmt.Printf("\n⚠️ Error reading password")
					continue
				}

				password = strings.TrimSpace(string(bytePassword))
			} else {
				password = *argPassword
			}

			// Login
			fmt.Printf("\n\n🔒 Logging in...")
			resp, gqlErr := gql.Login(ctx, email, password)
			if gqlErr == gql.InvalidUsernamePasssword {
				fmt.Printf("\n❌ Invalid email or password\n\n")
				if *argPassword != "" {
					os.Exit(1)
				}
				continue
			} else if gqlErr == gql.ServerError {
				fmt.Printf("\n💥 Error reaching server, try again later\n")
				os.Exit(1)
			}
			fmt.Printf("\n\n🔓 Successfully logged in as %s\n\n", email)
			gql.InitGQLClientWithToken(GraphQLURL, resp.Login.Token)
			tokenRsp, err := gql.GenerateServiceToken(ctx)
			if err != nil {
				fmt.Printf("\n💥 Error generating token %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("\nGenerated token: %s\n\n", tokenRsp.GenerateOrGetServiceToken)
			os.Exit(1)
		}
	}

	// Create WS Service
	WSService = websocket.NewWebsocketService(WSUrl, *maxDifficulty)

	// Loop to get username and password and login
	for {
		// Get username/password
		reader := bufio.NewReader(os.Stdin)

		var email string

		if *argEmail == "" {
			fmt.Print("➡️ Enter Email: ")
			rawEmail, err := reader.ReadString('\n')

			if err != nil {
				fmt.Printf("\n⚠️ Error reading email")
				continue
			}

			email = strings.TrimSpace(rawEmail)

			if !validation.IsValidEmail(email) {
				fmt.Printf("\n⚠️ Invalid email\n\n")
				continue
			}
		} else {
			if !validation.IsValidEmail(*argEmail) {
				fmt.Printf("\n⚠️ Invalid email\n\n")
				os.Exit(1)
			}
			email = *argEmail
		}

		var password string

		if *argPassword == "" {
			fmt.Print("➡️ Enter Password: ")
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))

			if err != nil {
				fmt.Printf("\n⚠️ Error reading password")
				continue
			}

			password = strings.TrimSpace(string(bytePassword))
		} else {
			password = *argPassword
		}

		// Login
		fmt.Printf("\n\n🔒 Logging in...")
		resp, gqlErr := gql.Login(ctx, email, password)
		if gqlErr == gql.InvalidUsernamePasssword {
			fmt.Printf("\n❌ Invalid email or password\n\n")
			if *argPassword != "" {
				os.Exit(1)
			}
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
	workProcessor := work.NewWorkProcessor(WSService, *gpuOnly)
	workProcessor.StartAsync()

	WSService.StartWSClient(ctx, workProcessor.WorkQueueChan, workProcessor.Queue)
}

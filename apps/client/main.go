package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Inkeliz/go-opencl/opencl"
	"github.com/bananocoin/boompow/apps/client/gql"
	"github.com/bananocoin/boompow/apps/client/websocket"
	"github.com/bananocoin/boompow/apps/client/work"
	"github.com/bananocoin/boompow/libs/utils/misc"
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
			} else {
				continue
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
	maxDifficulty := flag.Int("max-difficulty", 128, "The maximum work difficulty to compute, higher than this will be ignored")
	minDifficulty := flag.Int("min-difficulty", 1, "The minimum work difficulty to compute, lower than this will be ignored")
	noPrecache := flag.Bool("no-precache", false, "If set, will not compute precached work requests")
	// Benchmark
	benchmark := flag.Int("benchmark", 0, "Run a benchmark for the given number of random hashes")
	benchmarkDifficulty := flag.Int("benchmark-difficulty", 64, "The difficulty multiplier for the benchmark")
	// To login without username and password prompt
	argEmail := flag.String("email", "", "The email (username) to use for the worker (optional)")
	argPassword := flag.String("password", "", "The password to use for the worker (optional)")
	// OpenCL related things
	listDevices := flag.Bool("list-devices", false, "List available OpenCL devices/GPUs (optional)")
	gpus := flag.String("gpus", "0", "The GPUs to use for PoW, comma separated e.g. --gpu 0,1,2 (optional, default 0)")
	version := flag.Bool("version", false, "Display the version")
	flag.Parse()

	if *version {
		fmt.Printf("BoomPOW version: %s\n", Version)
		os.Exit(0)
	}

	// Parse GPU argument
	gpuSplit := strings.Split(*gpus, ",")
	gpuSplitInt := []int{}
	// Validate
	for _, gpu := range gpuSplit {
		asInt, err := strconv.Atoi(gpu)
		if err != nil {
			fmt.Printf("⚠️ Invalid GPU argument - not a number: %s", gpu)
			os.Exit(1)
		}
		gpuSplitInt = append(gpuSplitInt, asInt)
	}

	printBanner()

	gpuInfo, err := getGPUInfo()

	// See if we just want to list the deviecs
	if *listDevices {
		for key := range gpuInfo {
			fmt.Printf("\n⚡ GPU %d", key)
			fmt.Printf("\nPlatform: %s", gpuInfo[key].platformName)
			fmt.Printf("\nVendor: %s", gpuInfo[key].vendor)
			fmt.Printf("\nDriver: %s", gpuInfo[key].driverVersion)
		}
		fmt.Printf("\n")
		os.Exit(0)
	}

	found := false
	var devicesToUse []opencl.Device

	if err != nil {
		fmt.Printf("\n🚨 No GPU Found!")
		fmt.Printf("\nThis error is safe to ignore if you intended to generate PoW on CPU only")
		fmt.Printf("\nOtherwise you may want to check your GPU drivers and ensure it is properly installed, as well as ensure your device supports OpenCL 2.0\n\n")
	} else {
		for key := range gpuInfo {
			if !misc.Contains(gpuSplitInt, key) {
				continue
			}
			found = true
			fmt.Printf("\n⚡ Using GPU %d", key)
			fmt.Printf("\nPlatform: %s", gpuInfo[key].platformName)
			fmt.Printf("\nVendor: %s", gpuInfo[key].vendor)
			fmt.Printf("\nDriver: %s", gpuInfo[key].driverVersion)
			devicesToUse = append(devicesToUse, gpuInfo[key].device)
		}
		fmt.Printf("\n")
		if !found {
			fmt.Printf("\n🚨 No GPU Found or Invalid GPU Selected!")
			fmt.Printf("\nThis error is safe to ignore if you intended to generate PoW on CPU only")
			fmt.Printf("\nOtherwise you may want to check your GPU drivers and ensure it is properly installed, as well as ensure your device supports OpenCL 2.0\n\n")
		}
	}
	if *gpuOnly && found {
		fmt.Printf("\nOnly using GPU for work_generate...\n\n")
	} else if !found {
		fmt.Printf("\nOnly using CPU for work_generate...\n\n")
	} else {
		fmt.Printf("\nUsing GPU+CPU for work_generate...\n\n")
	}

	// Check benchmark
	if *benchmark > 0 {
		work.RunBenchmark(*benchmark, *benchmarkDifficulty, *gpuOnly, devicesToUse)
		os.Exit(0)
	}

	// Define context
	ctx, cancel := context.WithCancel(context.Background())
	gql.InitGQLClient(GraphQLURL)

	// Handle interrupts gracefully
	SetupCloseHandler(ctx, cancel)

	// Create WS Service
	WSService = websocket.NewWebsocketService(WSUrl, *maxDifficulty, *minDifficulty, *noPrecache)

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
	workProcessor := work.NewWorkProcessor(WSService, *gpuOnly, devicesToUse)
	workProcessor.StartAsync()

	WSService.StartWSClient(ctx, workProcessor.WorkQueueChan, workProcessor.Queue)
}

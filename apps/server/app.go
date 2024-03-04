package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/bananocoin/boompow/apps/server/graph"
	"github.com/bananocoin/boompow/apps/server/graph/generated"
	"github.com/bananocoin/boompow/apps/server/src/controller"
	"github.com/bananocoin/boompow/apps/server/src/database"
	"github.com/bananocoin/boompow/apps/server/src/middleware"
	"github.com/bananocoin/boompow/apps/server/src/net"
	"github.com/bananocoin/boompow/apps/server/src/repository"
	serializableModels "github.com/bananocoin/boompow/libs/models"
	"github.com/bananocoin/boompow/libs/utils"
	netutils "github.com/bananocoin/boompow/libs/utils/net"
	"github.com/bitfield/script"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"k8s.io/klog/v2"
)

const defaultPort = "8080"

func usage() {
	flag.PrintDefaults()
	os.Exit(2)
}

func runServer() {
	database.GetRedisDB().WipeAllConnectedClients()
	godotenv.Load()
	// Setup database conn
	config := &database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}
	fmt.Println("🏡 Connecting to database...")
	db, err := database.NewConnection(config)
	if err != nil {
		panic(err)
	}

	// fmt.Println("🦋 Running database migrations...")
	// err = database.Migrate(db)
	// if err != nil {
	// 	fmt.Printf("Error running database migrations %v", err)
	// 	os.Exit(1)
	// }

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Create repositories
	userRepo := repository.NewUserService((db))
	workRepo := repository.NewWorkService(db, userRepo)
	paymentRepo := repository.NewPaymentService(db)
	fmt.Println("Repository created")

	precacheMap := &sync.Map{}

	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
		UserRepo:    userRepo,
		WorkRepo:    workRepo,
		PaymentRepo: paymentRepo,
		PrecacheMap: precacheMap,
	}}))
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	// // Configure WebSocket with CORS
	// srv.AddTransport(&transport.Websocket{
	// 	Upgrader: websocket.Upgrader{
	// 		CheckOrigin: func(r *http.Request) bool {
	// 			return false
	// 		},
	// 		ReadBufferSize:  1024,
	// 		WriteBufferSize: 1024,
	// 	},
	// 	KeepAlivePingInterval: 10 * time.Second,
	// })
	if utils.GetEnv("ENVIRONMENT", "development") == "development" {
		srv.Use(extension.Introspection{})
	}

	// Setup router
	router := chi.NewRouter()
	// ! TODO - this is temporary, need to set origins in prod
	router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		//AllowedOrigins:   []string{"*"},
		AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	// if utils.GetEnv("ENVIRONMENT", "development") == "development" {
	// 	router.Use(cors.New(cors.Options{
	// 		AllowOriginFunc: func(origin string) bool {
	// 			return true
	// 		},
	// 	}).Handler)
	// } else {
	// 	router.Use(cors.New(cors.Options{
	// 		AllowedOrigins:   []string{"https://*.banano.cc"},
	// 		AllowCredentials: true,
	// 		Debug:            true,
	// 	}).Handler)
	// }
	router.Use(middleware.AuthMiddleware(userRepo))
	// Rate limiting middleware
	router.Use(httprate.Limit(
		20,            // requests
		1*time.Minute, // per duration
		// an oversimplified example of rate limiting by a custom header
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			requester := middleware.AuthorizedServiceToken(r.Context())
			if requester != nil {
				// Return a random string, effectively disabling rate limiting for services
				return uuid.New().String(), nil
			}
			return netutils.GetIPAddress(r), nil
		}),
	))
	if utils.GetEnv("ENVIRONMENT", "development") == "development" {
		router.Handle("/", playground.Handler("GraphQL playground", "/graphql"))
		log.Printf("🚀 connect to http://localhost:%s/ for GraphQL playground", port)
	}
	router.Handle("/graphql", srv)

	// Setup channel for stats processing job
	statsChan := make(chan repository.WorkMessage, 100)
	// Setup channel for sending block awarded messages
	blockAwardedChan := make(chan serializableModels.ClientMessage)

	// Setup WS endpoint
	controller.ActiveHub = controller.NewHub(&statsChan)
	go controller.ActiveHub.Run()
	router.HandleFunc("/ws/worker", func(w http.ResponseWriter, r *http.Request) {
		controller.WorkerChl(controller.ActiveHub, w, r)
	})

	// Stats stats processing job
	go workRepo.StatsWorker(statsChan, &blockAwardedChan)
	// Job for sending block awarded messages to user
	go controller.ActiveHub.BlockAwardedWorker(blockAwardedChan)

	// Setup callback clients for pre-caching

	// Start nano WS client
	callbackChan := make(chan *net.WSCallbackMsg, 100)
	if utils.GetEnv("NANO_WS_URL", "") != "" {
		go net.StartNanoWSClient(utils.GetEnv("NANO_WS_URL", ""), &callbackChan)
	}

	// Read channel to notify clients of blocks of new blocks
	go func() {
		for msg := range callbackChan {
			_, ok := precacheMap.LoadAndDelete(msg.Block.Previous)
			if !ok {
				continue
			}

			// We want to precache this if we don't have it
			_, err := workRepo.RetrieveWorkFromCache(msg.Hash, 64)
			if err == nil {
				// Already cached
				continue
			}

			workRequest := serializableModels.ClientMessage{
				RequesterEmail:       "nano@banano.cc",
				BlockAward:           true,
				MessageType:          serializableModels.WorkGenerate,
				RequestID:            uuid.NewString(),
				Hash:                 msg.Hash,
				DifficultyMultiplier: 64,
				Precache:             true,
			}

			controller.BroadcastWorkRequestAndWait(workRequest)
			if msg.Block.Subtype != "send" {
				continue
			}
		}
	}()

	// Start banano WS client
	callbackChanBanano := make(chan *net.WSCallbackMsg, 100)
	if utils.GetEnv("BANANO_WS_URL", "") != "" {
		go net.StartNanoWSClient(utils.GetEnv("BANANO_WS_URL", ""), &callbackChanBanano)
	}

	// Read channel to notify clients of blocks of new blocks
	go func() {
		for msg := range callbackChanBanano {
			_, ok := precacheMap.LoadAndDelete(msg.Block.Previous)
			if !ok {
				continue
			}

			// We want to precache this if we don't have it
			_, err := workRepo.RetrieveWorkFromCache(msg.Hash, 1)
			if err == nil {
				// Already cached
				continue
			}

			workRequest := serializableModels.ClientMessage{
				RequesterEmail:       "all@banano.cc",
				BlockAward:           true,
				MessageType:          serializableModels.WorkGenerate,
				RequestID:            uuid.NewString(),
				Hash:                 msg.Hash,
				DifficultyMultiplier: 1,
				Precache:             true,
			}

			controller.BroadcastWorkRequestAndWait(workRequest)
			if msg.Block.Subtype != "send" {
				continue
			}
		}
	}()

	// Update stats and setup cron
	repository.UpdateStats(paymentRepo, workRepo)
	fmt.Println("🕒 Setting up cron...")
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every(10).Minutes().Do(func() {
		repository.UpdateStats(paymentRepo, workRepo)
	})

	fmt.Println("🚀 Starting server...")
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func createService(serviceName string, serviceURL string) {
	godotenv.Load()
	// Setup database conn
	config := &database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}
	fmt.Println("🏡 Connecting to database...")
	db, err := database.NewConnection(config)
	if err != nil {
		panic(err)
	}

	userRepo := repository.NewUserService((db))

	// Create user
	token, err := userRepo.CreateService(fmt.Sprintf("%s@banano.cc", serviceName), serviceName, serviceURL)
	if err != nil {
		panic(err)
	}

	fmt.Printf("🔑 Service created with token: %s", token)
}

func main() {
	flag.Usage = usage
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "WARNING")
	flag.Set("v", "2")
	if utils.GetEnv("ENVIRONMENT", "development") == "development" {
		flag.Set("stderrthreshold", "INFO")
		flag.Set("v", "3")
	}
	gqlGen := flag.Bool("gqlgen", false, "Run gqlgen")
	dbReset := flag.Bool("db-reset", false, "Reset database")
	startServer := flag.Bool("runServer", false, "Run server")
	addService := flag.Bool("addService", false, "Add service")
	serviceName := flag.String("serviceName", "", "Service name")
	serviceURL := flag.String("serviceURL", "", "Service URL")
	flag.Parse()

	if *gqlGen {
		fmt.Printf("🤖 Running graphql generate...")
		script.Exec("bash -c 'gqlgen generate --verbose'").Stdout()
		os.Exit(0)
	}
	if *dbReset {
		fmt.Printf("💥 Nuking database...")
		script.Exec("bash -c './scripts/reset_db.sh'").Stdout()
		os.Exit(0)
	}
	if *startServer {
		runServer()
		os.Exit(0)
	}
	if *addService {
		if *serviceName == "" || *serviceURL == "" {
			flag.Usage()
			os.Exit(1)
		}
		createService(*serviceName, *serviceURL)
		os.Exit(0)
	}
	usage()
	os.Exit(1)
}

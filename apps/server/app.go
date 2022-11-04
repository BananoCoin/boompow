package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql"
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
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"k8s.io/klog/v2"
)

const defaultPort = "8080"

func usage() {
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "WARNING")
	flag.Set("v", "2")
	if utils.GetEnv("ENVIRONMENT", "development") == "development" {
		flag.Set("stderrthreshold", "INFO")
		flag.Set("v", "3")
	}
	flag.Parse()
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

	fmt.Println("🦋 Running database migrations...")
	err = database.Migrate(db)
	if err != nil {
		fmt.Printf("Error running database migrations %v", err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Create repositories
	userRepo := repository.NewUserService((db))
	workRepo := repository.NewWorkService(db, userRepo)
	paymentRepo := repository.NewPaymentService(db)

	precacheMap := &sync.Map{}

	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
		UserRepo:    userRepo,
		WorkRepo:    workRepo,
		PaymentRepo: paymentRepo,
		PrecacheMap: precacheMap,
	}}))
	srv.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		oc := graphql.GetOperationContext(ctx)
		fmt.Printf("around: %s %s", oc.OperationName, oc.RawQuery)
		return next(ctx)
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	// Configure WebSocket with CORS
	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		KeepAlivePingInterval: 10 * time.Second,
	})
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
	// Rate limiting middleware
	router.Use(httprate.Limit(
		20,            // requests
		1*time.Minute, // per duration
		// an oversimplified example of rate limiting by a custom header
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			return netutils.GetIPAddress(r), nil
		}),
	))
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

	log.Fatal(http.ListenAndServe(":"+port, router))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Error: must specify at least 1 argument")
		os.Exit(1)
	}
	arg := os.Args[1]

	switch arg {
	case "gqlgen":
		fmt.Printf("🤖 Running graphql generate...")
		script.Exec("bash -c 'gqlgen generate --verbose'").Stdout()
	case "db:reset":
		fmt.Printf("💥 Nuking database...")
		script.Exec("bash -c './scripts/reset_db.sh'").Stdout()
	case "server":
		runServer()
	default:
		fmt.Printf("Invalid command %s\n", arg)
		os.Exit(1)
	}
}

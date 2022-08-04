package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/bananocoin/boompow-next/services/server/graph"
	"github.com/bananocoin/boompow-next/services/server/graph/generated"
	"github.com/bananocoin/boompow-next/services/server/src/controller"
	"github.com/bananocoin/boompow-next/services/server/src/database"
	"github.com/bananocoin/boompow-next/services/server/src/middleware"
	"github.com/bananocoin/boompow-next/services/server/src/repository"
	"github.com/bitfield/script"
	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
)

const defaultPort = "8080"

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
	database.Migrate(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Create repositories
	userRepo := repository.NewUserService((db))

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
		UserRepo: userRepo,
	}}))

	// Setup router
	router := chi.NewRouter()
	router.Use(middleware.Middleware(userRepo))
	router.Handle("/", playground.Handler("GraphQL playground", "/graphql"))
	router.Handle("/graphql", srv)

	// Setup WS endpoint
	controller.ActiveHub = controller.NewHub()
	go controller.ActiveHub.Run()
	router.HandleFunc("/ws/worker", func(w http.ResponseWriter, r *http.Request) {
		controller.WorkerChl(controller.ActiveHub, w, r)
	})

	log.Printf("🚀 connect to http://localhost:%s/ for GraphQL playground", port)
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

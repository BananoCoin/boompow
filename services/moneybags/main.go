package main

import (
	"fmt"
	"os"

	"github.com/bananocoin/boompow-next/apps/server/src/database"
	"github.com/bananocoin/boompow-next/apps/server/src/repository"
	"github.com/bananocoin/boompow-next/libs/models"
	"github.com/bananocoin/boompow-next/libs/utils"
	"github.com/bananocoin/boompow-next/libs/utils/number"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// The way this process works is:
// 1) We get the unpaid works for each user
// 2) We figure out what percentage of the total prize pool this user has earned
// 3) We build payments for each user based on that amount
// 4) We ship the payments

func main() {
	// dryRun := flag.Bool("dry-run", false, "Dry run")
	// flag.Parse()

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

	userRepo := repository.NewUserService(db)
	workRepo := repository.NewWorkService(db, userRepo)

	fmt.Println("👽 Getting unpaid works...")
	res, err := workRepo.GetUnpaidWorkCount()

	if err != nil {
		fmt.Printf("❌ Error retrieving unpaid works %v", err)
		os.Exit(1)
	}

	// Compute the entire sum of the unpaid works
	totalSum := 0
	for _, v := range res {
		totalSum += v.DifficultySum
	}

	sendRequestsRaw := []models.SendRequest{}

	// Compute the percentage each user has earned and build payments
	for _, v := range res {
		percentageOfPool := float64(v.DifficultySum) / float64(totalSum)
		paymentAmount := percentageOfPool * float64(utils.GetTotalPrizePool())

		sendRequestsRaw = append(sendRequestsRaw, models.SendRequest{
			BaseRequest: models.SendAction,
			Wallet:      "",
			Source:      "",
			Destination: "",
			AmountRaw:   number.BananoToRaw(paymentAmount),
			// Just a unique payment identifier
			ID: fmt.Sprintf("%s:%s", v.BanAddress, uuid.New().String()),
		})

		fmt.Printf("💸 %s has earned %f%% of the pool, and will be paid %f\n", v.BanAddress, percentageOfPool*100, paymentAmount)
	}
}

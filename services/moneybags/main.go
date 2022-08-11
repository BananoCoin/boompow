package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bananocoin/boompow/apps/server/src/database"
	"github.com/bananocoin/boompow/apps/server/src/repository"
	"github.com/bananocoin/boompow/libs/models"
	"github.com/bananocoin/boompow/libs/utils"
	"github.com/bananocoin/boompow/libs/utils/number"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// The way this process works is:
// 1) We get the unpaid works for each user
// 2) We figure out what percentage of the total prize pool this user has earned
// 3) We build payments for each user based on that amount and save in database
// ! TODO
// 4) We ship the payments

func main() {
	dryRun := flag.Bool("dry-run", false, "Dry run")
	rpcSend := flag.Bool("rpc-send", false, "Broadcast pending payments")
	flag.Parse()

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
	paymentRepo := repository.NewPaymentService(db)

	// Do all of this within a transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		if !*rpcSend {
			fmt.Println("👽 Getting unpaid works...")
			var res []repository.UnpaidWorkResult
			if *dryRun {
				fmt.Println("🏃 Dry run mode - not actually sending payments")
				res, err = workRepo.GetUnpaidWorkCount(tx)
			} else {
				res, err = workRepo.GetUnpaidWorkCountAndMarkAllPaid(tx)
			}

			if err != nil {
				fmt.Printf("❌ Error retrieving unpaid works %v", err)
				return err
			}

			if len(res) == 0 {
				fmt.Println("🤷 No unpaid works found")
				return nil
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
					Wallet:      utils.GetWalletID(),
					Source:      utils.GetWalletAddress(),
					Destination: v.BanAddress,
					AmountRaw:   number.BananoToRaw(paymentAmount),
					// Just a unique payment identifier
					ID:     fmt.Sprintf("%s:%s", v.BanAddress, uuid.New().String()),
					PaidTo: v.ProvidedBy,
				})

				fmt.Printf("💸 %s has earned %f%% of the pool, and will be paid %f\n", v.BanAddress, percentageOfPool*100, paymentAmount)
			}

			if !*dryRun {
				err = paymentRepo.BatchCreateSendRequests(tx, sendRequestsRaw)
				if err != nil {
					fmt.Printf("❌ Error creating send requests %v", err)
					return err
				}
			}
			return nil
		}

		// Alternative job retrieves all payments from database with null block-hash and broadcasts them to the node
		fmt.Println("👽 Getting pending payments...")

		return nil
	})

	if err != nil {
		os.Exit(1)
	}

	// Success
	os.Exit(0)
}

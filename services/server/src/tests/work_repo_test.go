package tests

import (
	"os"
	"testing"
	"time"

	serializableModels "github.com/bananocoin/boompow-next/libs/models"
	utils "github.com/bananocoin/boompow-next/libs/utils/testing"
	"github.com/bananocoin/boompow-next/services/server/src/database"
	"github.com/bananocoin/boompow-next/services/server/src/repository"
)

// Test stats repo
func TestStatsRepo(t *testing.T) {
	os.Setenv("MOCK_REDIS", "true")
	mockDb, err := database.NewConnection(&database.Config{
		Host:     os.Getenv("DB_MOCK_HOST"),
		Port:     os.Getenv("DB_MOCK_PORT"),
		Password: os.Getenv("DB_MOCK_PASS"),
		User:     os.Getenv("DB_MOCK_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   "testing",
	})
	utils.AssertEqual(t, nil, err)
	err = database.DropAndCreateTables(mockDb)
	utils.AssertEqual(t, nil, err)
	userRepo := repository.NewUserService(mockDb)
	workRepo := repository.NewWorkService(mockDb, userRepo)

	// Create some users
	err = userRepo.CreateMockUsers()
	utils.AssertEqual(t, nil, err)

	providerEmail := "provider@gmail.com"
	requesterEmail := "requester@gmail.com"
	// Get users
	provider, _ := userRepo.GetUser(nil, &providerEmail)
	requester, _ := userRepo.GetUser(nil, &requesterEmail)

	_, err = workRepo.SaveOrUpdateWorkResult(repository.WorkMessage{
		RequestedByEmail:     requesterEmail,
		ProvidedByEmail:      providerEmail,
		Hash:                 "123",
		Result:               "ac",
		DifficultyMultiplier: 5,
	})
	utils.AssertEqual(t, nil, err)

	workRequest, err := workRepo.GetWorkRecord("123")
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, workRequest.DifficultyMultiplier, 5)
	utils.AssertEqual(t, "ac", workRequest.Result)
	utils.AssertEqual(t, requester.ID, workRequest.RequestedBy)
	utils.AssertEqual(t, provider.ID, workRequest.ProvidedBy)

	// Get other stuff
	workRequests, err := workRepo.GetUnpaidWorks()
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 1, len(workRequests))
	workRequestsByUser, err := workRepo.GetUnpaidWorksForUser(providerEmail)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 1, len(workRequestsByUser))

	// Test the worker
	statsChan := make(chan repository.WorkMessage, 100)
	blockAwardedChan := make(chan serializableModels.ClientMessage, 100)

	// Stats stats processing job
	go workRepo.StatsWorker(statsChan, &blockAwardedChan)

	statsChan <- repository.WorkMessage{
		RequestedByEmail:     requesterEmail,
		ProvidedByEmail:      providerEmail,
		Hash:                 "321",
		Result:               "fe",
		DifficultyMultiplier: 3,
	}

	time.Sleep(1 * time.Second) // Arbitrary time to wait for the worker to process the message
	workRequest, err = workRepo.GetWorkRecord("321")
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, workRequest.DifficultyMultiplier, 3)
	utils.AssertEqual(t, "fe", workRequest.Result)
	utils.AssertEqual(t, requester.ID, workRequest.RequestedBy)
	utils.AssertEqual(t, provider.ID, workRequest.ProvidedBy)
	utils.AssertEqual(t, 1, len(blockAwardedChan))
}

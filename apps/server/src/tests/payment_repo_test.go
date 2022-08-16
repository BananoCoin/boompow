package tests

import (
	"os"
	"strconv"
	"testing"

	"github.com/bananocoin/boompow/apps/server/src/database"
	"github.com/bananocoin/boompow/apps/server/src/repository"
	"github.com/bananocoin/boompow/libs/models"
	"github.com/bananocoin/boompow/libs/utils/number"
	utils "github.com/bananocoin/boompow/libs/utils/testing"
)

// Test payment repo
func TestPaymentrepo(t *testing.T) {
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
	paymentRepo := repository.NewPaymentService(mockDb)

	// Create some users
	err = userRepo.CreateMockUsers()
	utils.AssertEqual(t, nil, err)

	providerEmail := "provider@gmail.com"
	// Get user
	provider, _ := userRepo.GetUser(nil, &providerEmail)

	// Create some payments
	sendRequestsRaw := []models.SendRequest{}

	for i := 0; i < 3; i++ {
		sendRequestsRaw = append(sendRequestsRaw, models.SendRequest{
			BaseRequest: models.SendAction,
			Wallet:      strconv.FormatInt(int64(i), 10),
			Source:      strconv.FormatInt(int64(i), 10),
			Destination: strconv.FormatInt(int64(i), 10),
			AmountRaw:   number.BananoToRaw(float64(i)),
			// Just a unique payment identifier
			ID:     strconv.FormatInt(int64(i), 10),
			PaidTo: provider.ID,
		})
	}
	err = paymentRepo.BatchCreateSendRequests(mockDb, sendRequestsRaw)
	utils.AssertEqual(t, nil, err)

	// Get payments
	payments, err := paymentRepo.GetPendingPayments(mockDb)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 3, len(payments))
	for p := range payments {
		utils.AssertEqual(t, sendRequestsRaw[p].ID, payments[p].ID)
	}

	// Update one payment with block hash
	err = paymentRepo.SetBlockHash(mockDb, "1", "1")
	utils.AssertEqual(t, nil, err)

	// Check that we don't get this payment in the unpaid query
	payments, err = paymentRepo.GetPendingPayments(mockDb)
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 2, len(payments))
	// Assert that we don't get payment ID "1" which has been paid
	for p := range payments {
		utils.AssertEqual(t, true, payments[p].ID == "0" || payments[p].ID == "2")
	}

	// Check out total paid
	totalPaid, err := paymentRepo.GetTotalPaidBanano()
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 3.0+1728016, totalPaid)
}

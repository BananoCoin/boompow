package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/bananocoin/boompow/apps/server/graph/model"
	"github.com/bananocoin/boompow/apps/server/src/database"
	"github.com/bananocoin/boompow/apps/server/src/models"
	"github.com/bananocoin/boompow/apps/server/src/repository"
	utils "github.com/bananocoin/boompow/libs/utils/testing"
)

// Test user repo
func TestUserRepo(t *testing.T) {
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

	// Create user
	banAddress := "ban_3bsnis6ha3m9cepuaywskn9jykdggxcu8mxsp76yc3oinrt3n7gi77xiggtm"
	user, err := userRepo.CreateUser(&model.UserInput{
		Email:      "joe@gmail.com",
		Password:   "password",
		Type:       model.UserType(models.PROVIDER),
		BanAddress: &banAddress,
	}, false)
	utils.AssertEqual(t, nil, err)

	// Get user
	dbUser, err := userRepo.GetUser(&user.ID, nil)
	utils.AssertEqual(t, user.ID, dbUser.ID)
	utils.AssertEqual(t, "joe@gmail.com", dbUser.Email)
	utils.AssertEqual(t, false, dbUser.EmailVerified)
	utils.AssertEqual(t, banAddress, *dbUser.BanAddress)
	utils.AssertEqual(t, false, dbUser.CanRequestWork)
	utils.AssertEqual(t, models.PROVIDER, dbUser.Type)

	// Test confirm emial
	token, err := database.GetRedisDB().GetConfirmationToken(dbUser.Email)
	utils.AssertEqual(t, nil, err)

	userRepo.VerifyEmailToken(&model.VerifyEmailInput{
		Email: dbUser.Email,
		Token: token,
	})

	dbUser, err = userRepo.GetUser(&user.ID, nil)
	utils.AssertEqual(t, true, dbUser.EmailVerified)

	// Test authenticate
	authenticated := userRepo.Authenticate(&model.LoginInput{
		Email:    "joe@gmail.com",
		Password: "password",
	})
	utils.AssertEqual(t, true, authenticated)
	authenticated = userRepo.Authenticate(&model.LoginInput{
		Email:    "joe@gmail.com",
		Password: "wrongPassword",
	})
	utils.AssertEqual(t, false, authenticated)

	// Test delete user
	userRepo.DeleteUser(user.ID)
	dbUser, err = userRepo.GetUser(&user.ID, nil)
	utils.AssertEqual(t, true, err != nil)

	// TEst generate service token
	token = userRepo.GenerateServiceToken()
	utils.AssertEqual(t, 44, len(token))
	utils.AssertEqual(t, true, strings.HasPrefix(token, "service:"))
}

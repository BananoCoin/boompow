package repository

import (
	"errors"
	"fmt"

	"github.com/bananocoin/boompow-next/apps/server/graph/model"
	"github.com/bananocoin/boompow-next/apps/server/src/database"
	"github.com/bananocoin/boompow-next/apps/server/src/email"
	"github.com/bananocoin/boompow-next/apps/server/src/models"
	"github.com/bananocoin/boompow-next/libs/utils/auth"
	"github.com/bananocoin/boompow-next/libs/utils/validation"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepo interface {
	CreateUser(userInput *model.UserInput, doEmail bool) (*models.User, error)
	CreateMockUsers() error
	DeleteUser(id uuid.UUID) error
	GetUser(id *uuid.UUID, email *string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	Authenticate(loginInput *model.LoginInput) bool
	VerifyEmailToken(verifyEmail *model.VerifyEmailInput) (bool, error)
	GenerateServiceToken() string
}

type UserService struct {
	Db *gorm.DB
}

var _ UserRepo = &UserService{}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		Db: db,
	}
}

func (s *UserService) CreateMockUsers() error {
	// Hash password
	hashedPassword, err := auth.HashPassword("password")
	if err != nil {
		return err
	}
	banAddress := "ban_3bsnis6ha3m9cepuaywskn9jykdggxcu8mxsp76yc3oinrt3n7gi77xiggtm"
	provider := &models.User{
		Type:          models.PROVIDER,
		Email:         "provider@gmail.com",
		Password:      hashedPassword,
		EmailVerified: true,
		BanAddress:    &banAddress,
	}

	requester := &models.User{
		Type:           models.REQUESTER,
		Email:          "requester@gmail.com",
		Password:       hashedPassword,
		EmailVerified:  true,
		CanRequestWork: true,
	}

	err = s.Db.Create(&provider).Error

	if err != nil {
		return err
	}

	err = s.Db.Create(&requester).Error

	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) CreateUser(userInput *model.UserInput, doEmail bool) (*models.User, error) {
	// Validate
	if !validation.IsValidEmail(userInput.Email) {
		return nil, errors.New("Invalid email")
	}
	if models.UserType(userInput.Type) == models.PROVIDER && (userInput.BanAddress == nil || !validation.ValidateAddress(*userInput.BanAddress)) {
		return nil, errors.New("Invalid ban_ address")
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(userInput.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:    userInput.Email,
		Password: hashedPassword,
		Type:     models.UserType(userInput.Type),
	}
	if userInput.BanAddress != nil {
		user.BanAddress = userInput.BanAddress
	}
	err = s.Db.Create(&user).Error

	if err != nil {
		return nil, err
	}

	// Generate confirmation token and store in database
	confirmationToken, err := auth.GenerateRandHexString()
	if err != nil {
		return nil, err
	}

	database.GetRedisDB().SetConfirmationToken(userInput.Email, confirmationToken)
	// Send email with confirmation token
	if doEmail {
		email.SendConfirmationEmail(userInput.Email, confirmationToken)
	}
	return user, err
}

func (s *UserService) VerifyEmailToken(verifyEmail *model.VerifyEmailInput) (bool, error) {
	dbVerificationCode, err := database.GetRedisDB().GetConfirmationToken(verifyEmail.Email)
	if err != nil {
		return false, err
	} else if dbVerificationCode != verifyEmail.Token {
		return false, errors.New("Invalid verification code")
	}

	if res := s.Db.Model(&models.User{}).Where("email = ?", verifyEmail.Email).Update("email_verified", true); res.RowsAffected > 0 {
		// Email has been marked verified, delete the token
		database.GetRedisDB().DeleteConfirmationToken(verifyEmail.Email)
		return true, nil
	}
	return false, errors.New("Could not verify email")
}

func (s *UserService) DeleteUser(id uuid.UUID) error {
	user := &models.User{}
	err := s.Db.Delete(user, id).Error
	return err
}

func (s *UserService) GetUser(id *uuid.UUID, email *string) (*models.User, error) {
	if id == nil && email == nil {
		return nil, errors.New("id or email must be provided")
	}
	var err error
	user := &models.User{}
	if id != nil {
		err = s.Db.Where("id = ?", &id).First(user).Error
		return user, err
	}
	err = s.Db.Where("email = ?", &email).First(user).Error
	return user, err
}

func (s *UserService) GetAllUsers() ([]*models.User, error) {
	users := []*models.User{}
	err := s.Db.Find(&users).Error
	return users, err
}

// Compare password to hashed password, return true if match false otherwise
func (s *UserService) Authenticate(loginInput *model.LoginInput) bool {
	user := &models.User{}
	err := s.Db.Where("email = ?", &loginInput.Email).First(user).Error

	if err != nil {
		return false
	}

	return auth.CheckPasswordHash(loginInput.Password, user.Password)
}

// Generate a service token (for services to request work)
func (s *UserService) GenerateServiceToken() string {
	return fmt.Sprintf("service:%s", uuid.New().String())
}

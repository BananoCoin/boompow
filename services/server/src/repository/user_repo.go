package repository

import (
	"errors"

	"github.com/bbedward/boompow-ng/libs/utils/auth"
	"github.com/bbedward/boompow-ng/libs/utils/validation"
	"github.com/bbedward/boompow-ng/services/server/graph/model"
	"github.com/bbedward/boompow-ng/services/server/src/database"
	"github.com/bbedward/boompow-ng/services/server/src/email"
	"github.com/bbedward/boompow-ng/services/server/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepo interface {
	CreateUser(userInput *model.UserInput) (*models.User, error)
	UpdateUser(userInput *model.UserInput, id uuid.UUID) error
	DeleteUser(id uuid.UUID) error
	GetUser(id *uuid.UUID, email *string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	Authenticate(loginInput *model.LoginInput) bool
	VerifyEmailToken(verifyEmail *model.VerifyEmailInput) (bool, error)
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

func (s *UserService) CreateUser(userInput *model.UserInput) (*models.User, error) {
	// Validate
	if !validation.IsValidEmail(userInput.Email) {
		return nil, errors.New("Invalid email")
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
	email.SendConfirmationEmail(userInput.Email, confirmationToken)

	return user, err
}

func (s *UserService) VerifyEmailToken(verifyEmail *model.VerifyEmailInput) (bool, error) {
	dbVerificationCode, err := database.GetRedisDB().GetUserIDForConfirmationToken(verifyEmail.Email)
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

func (s *UserService) UpdateUser(userInput *model.UserInput, id uuid.UUID) error {
	// Validate
	if !validation.IsValidEmail(userInput.Email) {
		return errors.New("Invalid email")
	}

	user := models.User{
		Base: models.Base{
			ID: id,
		},
		Email:    userInput.Email,
		Password: userInput.Password,
	}
	err := s.Db.Model(&user).Where("id = ?", id).Updates(user).Error
	return err
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

package repository

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/bananocoin/boompow/apps/server/graph/model"
	"github.com/bananocoin/boompow/apps/server/src/database"
	"github.com/bananocoin/boompow/apps/server/src/email"
	"github.com/bananocoin/boompow/apps/server/src/models"
	"github.com/bananocoin/boompow/libs/utils/auth"
	"github.com/bananocoin/boompow/libs/utils/validation"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"gorm.io/gorm"
)

type UserRepo interface {
	CreateUser(userInput *model.UserInput, doEmail bool) (*models.User, error)
	SendConfirmEmailEmail(userEmail string, userType models.UserType, actuallyDoEmail bool) error
	CreateMockUsers() error
	DeleteUser(id uuid.UUID) error
	GetUser(id *uuid.UUID, email *string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	Authenticate(loginInput *model.LoginInput) bool
	VerifyEmailToken(verifyEmail *model.VerifyEmailInput) (bool, error)
	VerifyService(verifyService *model.VerifyServiceInput) (bool, error)
	GenerateResetPasswordRequest(resetPasswordInput *model.ResetPasswordInput, doEmail bool) (string, error)
	GenerateServiceToken() string
	GetNumberServices() (int64, error)
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

	// Providers must have ban address
	if models.UserType(userInput.Type) == models.PROVIDER && (userInput.BanAddress == nil || !validation.ValidateAddress(*userInput.BanAddress)) {
		return nil, errors.New("Invalid ban_ address")
	}

	// Requesters must have name and description
	if models.UserType(userInput.Type) == models.REQUESTER && (userInput.ServiceName == nil || userInput.ServiceWebsite == nil) {
		return nil, errors.New("Service name and service website are required")
	}

	if models.UserType(userInput.Type) == models.REQUESTER {
		_, err := url.ParseRequestURI(*userInput.ServiceWebsite)
		if err != nil {
			return nil, errors.New("Invalid website URL")
		}
		if strings.HasPrefix(*userInput.ServiceWebsite, "http://") {
			return nil, errors.New("Only https websites are supported")
		}

		if len(*userInput.ServiceName) < 3 {
			return nil, errors.New("Service name must be at least 3 characters long")
		}
	}

	err := validation.ValidatePassword(userInput.Password)
	if err != nil {
		return nil, err
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(userInput.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:          userInput.Email,
		Password:       hashedPassword,
		Type:           models.UserType(userInput.Type),
		BanAddress:     userInput.BanAddress,
		ServiceName:    userInput.ServiceName,
		ServiceWebsite: userInput.ServiceWebsite,
	}
	if userInput.BanAddress != nil {
		user.BanAddress = userInput.BanAddress
	}
	err = s.Db.Create(&user).Error

	if err != nil {
		if strings.Contains(err.(*pgconn.PgError).Message, "duplicate key value violates unique constraint") {
			return nil, errors.New("Email already exists")
		}
		return nil, errors.New("Unknown error creating user")
	}

	// Send the email
	err = s.SendConfirmEmailEmail(userInput.Email, models.UserType(userInput.Type), doEmail)

	return user, err
}

func (s *UserService) SendConfirmEmailEmail(userEmail string, userType models.UserType, actuallyDoEmail bool) error {
	// Generate confirmation token and store in database
	confirmationToken, err := auth.GenerateRandHexString()
	if err != nil {
		return err
	}

	database.GetRedisDB().SetConfirmationToken(userEmail, confirmationToken)
	// Send email with confirmation token
	if actuallyDoEmail {
		if err = email.SendConfirmationEmail(userEmail, userType, confirmationToken); err != nil {
			return err
		}
	}
	return nil
}

func (s *UserService) GenerateResetPasswordRequest(resetPasswordInput *model.ResetPasswordInput, doEmail bool) (string, error) {
	// Validate
	if !validation.IsValidEmail(resetPasswordInput.Email) {
		return "", errors.New("Invalid email")
	}

	// Get user
	user, err := s.GetUser(nil, &resetPasswordInput.Email)
	if err != nil || user == nil {
		return "", errors.New("No such user")
	}

	// Generate reset password token and store in database
	resetPasswordToken, err := auth.GenerateRandHexString()
	if err != nil {
		return "", err
	}

	database.GetRedisDB().SetResetPasswordToken(resetPasswordInput.Email, resetPasswordToken)
	// Send email with reset password token token
	if doEmail {
		email.SendResetPasswordEmail(resetPasswordInput.Email, resetPasswordToken)
	}
	return resetPasswordToken, err
}

func (s *UserService) VerifyEmailToken(verifyEmail *model.VerifyEmailInput) (bool, error) {
	dbVerificationCode, err := database.GetRedisDB().GetConfirmationToken(verifyEmail.Email)
	if err != nil {
		return false, errors.New("Invalid verification code, it may have expired")
	} else if dbVerificationCode != verifyEmail.Token {
		return false, errors.New("Invalid verification code")
	}

	if res := s.Db.Model(&models.User{}).Where("email = ?", verifyEmail.Email).Update("email_verified", true); res.RowsAffected > 0 {
		// Email has been marked verified, delete the token
		database.GetRedisDB().DeleteConfirmationToken(verifyEmail.Email)
		// Get User
		user, err := s.GetUser(nil, &verifyEmail.Email)
		if err != nil {
			glog.Errorf("Failed to retrieve user after verifying email: %v", err)
			// Don't choke because of this
			return true, nil
		}

		// For requesters, send another email for us to approve them to request work
		if user.Type == models.REQUESTER {
			// Generate token
			approvalToken, err := auth.GenerateRandHexString()
			if err != nil {
				glog.Errorf("Failed generate approvalToken: %v", err)
				// Don't choke because of this
				return true, nil
			}
			// Store token in redis
			database.GetRedisDB().SetApproveServiceToken(verifyEmail.Email, approvalToken)
			// Send email with token
			email.SendAuthorizeServiceEmail(user.Email, *user.ServiceName, *user.ServiceWebsite, approvalToken)
		}

		return true, nil
	}
	return false, errors.New("Could not verify email")
}

func (s *UserService) VerifyService(verifyService *model.VerifyServiceInput) (bool, error) {
	dbVerificationCode, err := database.GetRedisDB().GetApproveServiceToken(verifyService.Email)
	if err != nil {
		return false, errors.New("Invalid verification code, it may have expired")
	} else if dbVerificationCode != verifyService.Token {
		return false, errors.New("Invalid verification code")
	}

	if res := s.Db.Model(&models.User{}).Where("email = ?", verifyService.Email).Update("can_request_work", true); res.RowsAffected > 0 {
		// Email has been marked verified, delete the token
		database.GetRedisDB().DeleteApproveServiceToken(verifyService.Email)
		email.SendServiceApprovedEmail(verifyService.Email)
		return true, nil
	}
	return false, errors.New("Could not verify token")
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

func (s *UserService) GetNumberServices() (int64, error) {
	var count int64
	if err := s.Db.Model(&models.User{}).Where("type = ?", models.REQUESTER).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
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

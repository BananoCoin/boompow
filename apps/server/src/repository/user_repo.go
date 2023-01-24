package repository

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/bananocoin/boompow/apps/server/graph/model"
	"github.com/bananocoin/boompow/apps/server/src/database"
	"github.com/bananocoin/boompow/apps/server/src/email"
	"github.com/bananocoin/boompow/apps/server/src/models"
	"github.com/bananocoin/boompow/libs/utils/auth"
	"github.com/bananocoin/boompow/libs/utils/validation"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

type UserRepo interface {
	CreateUser(userInput *model.UserInput, doEmail bool) (*models.User, error)
	SendConfirmEmailEmail(userEmail string, userType models.UserType, actuallyDoEmail bool) error
	CreateMockUsers() error
	DeleteUser(id uuid.UUID) error
	GetUser(id *uuid.UUID, email *string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	Authenticate(loginInput *model.LoginInput) *models.User
	VerifyEmailToken(verifyEmail *model.VerifyEmailInput) (bool, error)
	VerifyService(verifyService *model.VerifyServiceInput) (bool, error)
	GenerateResetPasswordRequest(resetPasswordInput *model.ResetPasswordInput, doEmail bool) (string, error)
	GenerateServiceToken() string
	CreateService(email string, serviceName string, serviceWebsite string) (string, error)
	GetNumberServices() (int64, error)
	ChangePassword(email string, userInput *model.ChangePasswordInput) error
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

	serviceName := "Service Name"
	serviceWebsite := "https://service.com"

	requester := &models.User{
		Type:           models.REQUESTER,
		Email:          "requester@gmail.com",
		Password:       hashedPassword,
		EmailVerified:  true,
		CanRequestWork: true,
		ServiceName:    &serviceName,
		ServiceWebsite: &serviceWebsite,
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

func (s *UserService) CreateService(email string, serviceName string, serviceWebsite string) (string, error) {
	// Validate
	if !validation.IsValidEmail(email) {
		return "", errors.New("Invalid email")
	}
	// Just gen random password
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ" +
		"abcdefghijklmnopqrstuvwxyzåäö" +
		"0123456789")
	length := 15
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	password := b.String()

	// Hash password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return "", err
	}

	user := &models.User{
		Email:          strings.ToLower(email),
		Password:       hashedPassword,
		Type:           models.UserType(models.PROVIDER),
		ServiceName:    &serviceName,
		ServiceWebsite: &serviceWebsite,
		EmailVerified:  true,
		CanRequestWork: true,
	}
	err = s.Db.Create(&user).Error

	if err != nil {
		if strings.Contains(err.(*pgconn.PgError).Message, "duplicate key value violates unique constraint") {
			return "", errors.New("Email already exists")
		}
		return "", errors.New("Unknown error creating user")
	}

	// Create token
	// Generate token
	token := s.GenerateServiceToken()

	if err := database.GetRedisDB().AddServiceToken(user.ID, token); err != nil {
		return "", fmt.Errorf("error generating token")
	}

	return token, nil
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
		Email:          strings.ToLower(userInput.Email),
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
	err = s.SendConfirmEmailEmail(strings.ToLower(userInput.Email), models.UserType(userInput.Type), doEmail)

	return user, err
}

func (s *UserService) SendConfirmEmailEmail(userEmail string, userType models.UserType, actuallyDoEmail bool) error {
	// Generate confirmation token and store in database
	confirmationToken, err := auth.GenerateRandHexString()
	if err != nil {
		return err
	}

	err = database.GetRedisDB().SetConfirmationToken(userEmail, confirmationToken)
	if err != nil {
		klog.Errorf("Error setting confirmation token: %v", err)
		return err
	}
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
	lower := strings.ToLower(resetPasswordInput.Email)
	user, err := s.GetUser(nil, &lower)
	if err != nil || user == nil {
		return "", errors.New("No such user")
	}

	// Generate reset password token and store in database
	resetPasswordToken, err := auth.GenerateToken(strings.ToLower(user.Email), time.Now)
	if err != nil {
		return "", err
	}
	resetPasswordToken = fmt.Sprintf("resetpassword:%s", resetPasswordToken)

	database.GetRedisDB().SetResetPasswordToken(resetPasswordInput.Email, resetPasswordToken)
	// Send email with reset password token token
	if doEmail {
		email.SendResetPasswordEmail(resetPasswordInput.Email, resetPasswordToken)
	}
	return resetPasswordToken, err
}

func (s *UserService) ChangePassword(email string, userInput *model.ChangePasswordInput) error {
	// Hash password
	hashedPassword, err := auth.HashPassword(userInput.NewPassword)
	if err != nil {
		return err
	}

	if res := s.Db.Model(&models.User{}).Where("email = ?", email).Update("password", hashedPassword); res.RowsAffected > 0 {
		// Email has been marked verified, delete the token
		database.GetRedisDB().DeleteResetPasswordToken(email)

		return nil
	}
	return errors.New("Could not change password")
}

func (s *UserService) VerifyEmailToken(verifyEmail *model.VerifyEmailInput) (bool, error) {
	lowerEmail := strings.ToLower(verifyEmail.Email)
	dbVerificationCode, err := database.GetRedisDB().GetConfirmationToken(lowerEmail)
	if err != nil {
		return false, errors.New("Invalid verification code, it may have expired")
	} else if dbVerificationCode != verifyEmail.Token {
		return false, errors.New("Invalid verification code")
	}

	if res := s.Db.Model(&models.User{}).Where("email = ?", lowerEmail).Update("email_verified", true); res.RowsAffected > 0 {
		// Email has been marked verified, delete the token
		database.GetRedisDB().DeleteConfirmationToken(lowerEmail)
		// Get User
		user, err := s.GetUser(nil, &verifyEmail.Email)
		if err != nil {
			klog.Errorf("Failed to retrieve user after verifying email: %v", err)
			// Don't choke because of this
			return true, nil
		}

		// For requesters, send another email for us to approve them to request work
		if user.Type == models.REQUESTER {
			// Generate token
			approvalToken, err := auth.GenerateRandHexString()
			if err != nil {
				klog.Errorf("Failed generate approvalToken: %v", err)
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
func (s *UserService) Authenticate(loginInput *model.LoginInput) *models.User {
	user := &models.User{}
	emailLower := strings.ToLower(loginInput.Email)
	err := s.Db.Where("lower(email) = ?", &emailLower).First(user).Error

	if err != nil {
		return nil
	}

	if auth.CheckPasswordHash(loginInput.Password, user.Password) {
		return user
	}
	return nil
}

// Generate a service token (for services to request work)
func (s *UserService) GenerateServiceToken() string {
	return fmt.Sprintf("service:%s", uuid.New().String())
}

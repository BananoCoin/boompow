package repository

import (
	"errors"

	"github.com/bbedward/boompow-server-ng/graph/model"
	"github.com/bbedward/boompow-server-ng/src/models"
	"github.com/bbedward/boompow-server-ng/src/utils/auth"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepo interface {
	CreateUser(userInput *model.UserInput) (*models.User, error)
	UpdateUser(userInput *model.UserInput, id uuid.UUID) error
	DeleteUser(id uuid.UUID) error
	GetUser(id *uuid.UUID, username *string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	Authenticate(loginInput *model.Login) bool
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
	// Hash password
	hashedPassword, err := auth.HashPassword(userInput.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username: userInput.Username,
		Password: hashedPassword,
	}
	err = s.Db.Create(&user).Error

	return user, err
}

func (s *UserService) UpdateUser(userInput *model.UserInput, id uuid.UUID) error {
	user := models.User{
		Base: models.Base{
			ID: id,
		},
		Username: userInput.Username,
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

func (s *UserService) GetUser(id *uuid.UUID, username *string) (*models.User, error) {
	if id == nil && username == nil {
		return nil, errors.New("id or username must be provided")
	}
	var err error
	user := &models.User{}
	if id != nil {
		err = s.Db.Where("id = ?", &id).First(user).Error
		return user, err
	}
	err = s.Db.Where("username = ?", &username).First(user).Error
	return user, err
}

func (s *UserService) GetAllUsers() ([]*models.User, error) {
	users := []*models.User{}
	err := s.Db.Find(&users).Error
	return users, err
}

// Compare password to hashed password, return true if match false otherwise
func (s *UserService) Authenticate(loginInput *model.Login) bool {
	user := &models.User{}
	err := s.Db.Where("username = ?", &loginInput.Username).First(user).Error

	if err != nil {
		return false
	}

	return auth.CheckPasswordHash(loginInput.Password, user.Password)
}

package repository

import (
	"github.com/bbedward/boompow-server-ng/graph/model"
	"github.com/bbedward/boompow-server-ng/src/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepo interface {
	CreateUser(userInput *model.UserInput) (*models.User, error)
	UpdateUser(userInput *model.UserInput, id uuid.UUID) error
	DeleteUser(id uuid.UUID) error
	GetOneUser(id uuid.UUID) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
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
	user := &models.User{
		Username: userInput.Username,
		Password: userInput.Password,
	}
	err := s.Db.Create(&user).Error

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

func (s *UserService) GetOneUser(id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := s.Db.Where("id = ?", id).First(user).Error
	return user, err
}

func (s *UserService) GetAllUsers() ([]*models.User, error) {
	users := []*models.User{}
	err := s.Db.Find(&users).Error
	return users, err
}

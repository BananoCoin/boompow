package repository

import (
	"github.com/bananocoin/boompow-next/services/server/src/models"
	"github.com/golang/glog"
	"gorm.io/gorm"
)

type StatsMessage struct {
	RequestedByEmail     string `json:"requestedByEmail"`
	ProvidedByEmail      string `json:"providedByEmail"`
	Hash                 string `json:"hash"`
	Result               string `json:"result"`
	DifficultyMultiplier int    `json:"difficulty_multiplier"`
}

type StatsRepo interface {
	SaveWorkRequest(statsMessage StatsMessage) (*models.WorkRequest, error)
	StatsWorker(statsChan <-chan StatsMessage)
}

type StatsService struct {
	Db       *gorm.DB
	userRepo UserRepo
}

var _ StatsRepo = &StatsService{}

func NewStatsService(db *gorm.DB, userRepo UserRepo) *StatsService {
	return &StatsService{
		Db:       db,
		userRepo: userRepo,
	}
}

func (s *StatsService) SaveWorkRequest(statsMessage StatsMessage) (*models.WorkRequest, error) {
	// Get provider and requester
	provider, err := s.userRepo.GetUser(nil, &statsMessage.ProvidedByEmail)
	if err != nil {
		return nil, err
	}
	// Get requester
	requester, err := s.userRepo.GetUser(nil, &statsMessage.RequestedByEmail)
	if err != nil {
		return nil, err
	}

	// Create work request
	workRequestDb := &models.WorkRequest{
		Hash:                 statsMessage.Hash,
		DifficultyMultiplier: statsMessage.DifficultyMultiplier,
		Result:               statsMessage.Result,
		ProvidedBy:           provider.ID,
		RequestedBy:          requester.ID,
	}

	err = s.Db.Create(&workRequestDb).Error

	if err != nil {
		return nil, err
	}

	return workRequestDb, err
}

func (s *StatsService) StatsWorker(statsChan <-chan StatsMessage) {
	for c := range statsChan {
		_, err := s.SaveWorkRequest(c)
		if err != nil {
			glog.Errorf("Error saving work stats %v", err)
		}
	}
}

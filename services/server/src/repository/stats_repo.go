package repository

import (
	serializableModels "github.com/bananocoin/boompow-next/libs/models"
	"github.com/bananocoin/boompow-next/libs/utils"
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
	GetStatsRecord(hash string) (*models.WorkRequest, error)
	StatsWorker(statsChan <-chan StatsMessage, blockAwardedChan *chan serializableModels.ClientMessage)
	GetUnpaidStatsForUser(email string) ([]*models.WorkRequest, error)
	GetUnpaidStats() ([]*models.WorkRequest, error)
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

func (s *StatsService) GetStatsRecord(hash string) (*models.WorkRequest, error) {
	var workRequest models.WorkRequest
	err := s.Db.Where("hash = ?", hash).First(&workRequest).Error
	if err != nil {
		return nil, err
	}
	return &workRequest, nil
}

func (s *StatsService) GetUnpaidStatsForUser(email string) ([]*models.WorkRequest, error) {
	// Get user
	user, err := s.userRepo.GetUser(nil, &email)
	if err != nil {
		return nil, err
	}
	stats := make([]*models.WorkRequest, 0)
	err = s.Db.Where("provided_by = ?", user.ID).Where("awarded = ?", false).Find(&stats).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *StatsService) GetUnpaidStats() ([]*models.WorkRequest, error) {
	stats := make([]*models.WorkRequest, 0)
	err := s.Db.Where("awarded = ?", false).Find(&stats).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *StatsService) StatsWorker(statsChan <-chan StatsMessage, blockAwardedChan *chan serializableModels.ClientMessage) {
	for c := range statsChan {
		_, err := s.SaveWorkRequest(c)
		if err != nil {
			glog.Errorf("Error saving work stats %v", err)
			continue
		}
		// Process message to send to user
		// Get total unpaid stats
		unpaidStats, err := s.GetUnpaidStats()
		if err != nil {
			glog.Errorf("Error getting unpaid stats %v", err)
		}
		// Get unpaid stats for this user
		unpaidUserStats, err := s.GetUnpaidStatsForUser(c.ProvidedByEmail)
		if err != nil {
			glog.Errorf("Error getting unpaid stats for user %v", err)
		}
		// Get percentage of unpaid stats for this user
		percentageOfPool := float64(len(unpaidUserStats)) / float64(len(unpaidStats)) * 100
		prizePool := utils.GetTotalPrizePool()
		estimatedAward := float64(prizePool) * percentageOfPool / 100
		// Format client message
		blockAwardedMsg := serializableModels.ClientMessage{
			MessageType:    serializableModels.BlockAwarded,
			Hash:           c.Hash,
			PercentOfPool:  percentageOfPool,
			EstimatedAward: estimatedAward,
			ProviderEmail:  c.ProvidedByEmail,
		}

		go func() { *blockAwardedChan <- blockAwardedMsg }()
	}
}

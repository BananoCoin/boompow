package repository

import (
	"errors"
	"time"

	"github.com/bananocoin/boompow/apps/server/src/models"
	serializableModels "github.com/bananocoin/boompow/libs/models"
	"github.com/bananocoin/boompow/libs/utils"
	"github.com/bananocoin/boompow/libs/utils/validation"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkMessage struct {
	RequestedByEmail     string `json:"requestedByEmail"`
	ProvidedByEmail      string `json:"providedByEmail"`
	Hash                 string `json:"hash"`
	Result               string `json:"result"`
	DifficultyMultiplier int    `json:"difficulty_multiplier"`
}

type WorkRepo interface {
	SaveOrUpdateWorkResult(workMessage WorkMessage) (*models.WorkResult, error)
	GetWorkRecord(hash string) (*models.WorkResult, error)
	StatsWorker(statsChan <-chan WorkMessage, blockAwardedChan *chan serializableModels.ClientMessage)
	GetUnpaidWorkSumForUser(email string) (int, error)
	GetUnpaidWorkSum() (int, error)
	RetrieveWorkFromCache(hash string, difficultyMultiplier int) (*models.WorkResult, error)
	GetUnpaidWorkCount(tx *gorm.DB) ([]UnpaidWorkResult, error)
	GetUnpaidWorkCountAndMarkAllPaid(tx *gorm.DB) ([]UnpaidWorkResult, error)
}

type WorkService struct {
	Db       *gorm.DB
	userRepo UserRepo
}

var _ WorkRepo = &WorkService{}

func NewWorkService(db *gorm.DB, userRepo UserRepo) *WorkService {
	return &WorkService{
		Db:       db,
		userRepo: userRepo,
	}
}

func (s *WorkService) SaveOrUpdateWorkResult(workMessage WorkMessage) (*models.WorkResult, error) {
	// Get provider and requester
	provider, err := s.userRepo.GetUser(nil, &workMessage.ProvidedByEmail)
	if err != nil {
		return nil, err
	}
	// Get requester
	requester, err := s.userRepo.GetUser(nil, &workMessage.RequestedByEmail)
	if err != nil {
		return nil, err
	}

	// See if exists
	var workResult models.WorkResult
	var workRequestDb *models.WorkResult
	err = s.Db.Where("hash = ?", workMessage.Hash).First(&workResult).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {

		// Create work request
		workRequestDb = &models.WorkResult{
			Hash:                 workMessage.Hash,
			DifficultyMultiplier: workMessage.DifficultyMultiplier,
			Result:               workMessage.Result,
			ProvidedBy:           provider.ID,
			RequestedBy:          requester.ID,
		}

		err = s.Db.Create(&workRequestDb).Error

		if err != nil {
			return nil, err
		}
	} else if err == nil {
		// Update record
		err = s.Db.Model(&workResult).Updates(map[string]interface{}{"difficulty_multiplier": workMessage.DifficultyMultiplier, "result": workMessage.Result, "provided_by": provider.ID, "requested_by": requester.ID, "awarded": false}).Error
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	// Update timestamps
	err = s.Db.Model(&models.User{}).Where("id = ?", provider.ID).Updates(map[string]interface{}{"last_provided_work_at": time.Now()}).Error
	if err != nil {
		glog.Errorf("Failed to update last_provided_work_at for provider %v", err)
	}
	err = s.Db.Model(&models.User{}).Where("id = ?", requester.ID).Updates(map[string]interface{}{"last_requested_work_at": time.Now()}).Error
	if err != nil {
		glog.Errorf("Failed to update last_requested_work_at for provider %v", err)
	}

	return workRequestDb, err
}

func (s *WorkService) GetWorkRecord(hash string) (*models.WorkResult, error) {
	var workRequest models.WorkResult
	err := s.Db.Where("hash = ?", hash).First(&workRequest).Error
	if err != nil {
		return nil, err
	}
	return &workRequest, nil
}

// Get sum of (difficulty_multiplier * 100), use this to determine payments

type UnpaidSumResult struct {
	DifficultySum int `json:"difficulty_sum"`
}

func (s *WorkService) GetUnpaidWorkSumForUser(email string) (int, error) {
	// Get user
	user, err := s.userRepo.GetUser(nil, &email)
	if err != nil {
		return 0, err
	}
	var result UnpaidSumResult
	err = s.Db.Model(&models.WorkResult{}).Select("sum(difficulty_multiplier*100) as difficulty_sum").Where("awarded = ?", false).Where("provided_by = ?", user.ID).Scan(&result).Error
	if err != nil {
		return 0, err
	}

	return result.DifficultySum, nil
}

// Summate the difficulty of unpaid works for all users
func (s *WorkService) GetUnpaidWorkSum() (int, error) {
	var result UnpaidSumResult
	err := s.Db.Model(&models.WorkResult{}).Select("sum(difficulty_multiplier*100) as difficulty_sum").Where("awarded = ?", false).Scan(&result).Error
	if err != nil {
		return 0, err
	}

	return result.DifficultySum, nil
}

type UnpaidWorkResult struct {
	UnpaidSumResult
	UnpaidCount int       `json:"unpaid_count"`
	ProvidedBy  uuid.UUID `json:"provided_by"`
	BanAddress  string    `json:"ban_address"`
}

func (s *WorkService) GetUnpaidWorkCount(tx *gorm.DB) ([]UnpaidWorkResult, error) {
	var result []UnpaidWorkResult
	// x 100 for more precision
	err := tx.Model(&models.WorkResult{}).Select("COUNT(*) as unpaid_count, provided_by, ban_address, sum(difficulty_multiplier*100) as difficulty_sum").Joins("JOIN users on users.id = work_results.provided_by").Group("provided_by").Group("ban_address").Where("awarded = ?", false).Find(&result).Error
	return result, err
}

func (s *WorkService) GetUnpaidWorkCountAndMarkAllPaid(tx *gorm.DB) ([]UnpaidWorkResult, error) {
	result, err := s.GetUnpaidWorkCount(tx)
	if err != nil {
		return nil, err
	}
	err = tx.Model(&models.WorkResult{}).Where("1=1").Update("awarded", true).Error
	return result, err
}

func (s *WorkService) RetrieveWorkFromCache(hash string, difficultyMultiplier int) (*models.WorkResult, error) {
	var workRequest models.WorkResult
	err := s.Db.Where("hash = ?", hash).First(&workRequest).Error
	if err != nil {
		return nil, err
	}

	// Validate difficulty is valid
	if !validation.IsWorkValid(hash, difficultyMultiplier, workRequest.Result) {
		return nil, gorm.ErrRecordNotFound
	}
	return &workRequest, nil
}

func (s *WorkService) StatsWorker(statsChan <-chan WorkMessage, blockAwardedChan *chan serializableModels.ClientMessage) {
	for c := range statsChan {
		_, err := s.SaveOrUpdateWorkResult(c)
		if err != nil {
			glog.Errorf("Error saving work stats %v", err)
			continue
		}
		// Process message to send to user
		// Get total unpaid stats
		unpaidStats, err := s.GetUnpaidWorkSum()
		if err != nil {
			glog.Errorf("Error getting unpaid stats %v", err)
		}
		// Get unpaid stats for this user
		unpaidUserStats, err := s.GetUnpaidWorkSumForUser(c.ProvidedByEmail)
		if err != nil {
			glog.Errorf("Error getting unpaid stats for user %v", err)
		}
		// Get percentage of unpaid stats for this user
		percentageOfPool := float64(unpaidUserStats) / float64(unpaidStats) * 100
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

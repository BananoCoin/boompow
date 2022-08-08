package repository

import (
	"github.com/bananocoin/boompow-next/apps/server/src/models"
	serializableModels "github.com/bananocoin/boompow-next/libs/models"
	"gorm.io/gorm"
)

type PaymentRepo interface {
	BatchCreateSendRequests(tx *gorm.DB, sendRequests []serializableModels.SendRequest) error
}

type PaymentService struct {
	Db *gorm.DB
}

var _ PaymentRepo = &PaymentService{}

func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{
		Db: db,
	}
}

// Create payments in database
func (s *PaymentService) BatchCreateSendRequests(tx *gorm.DB, sendRequests []serializableModels.SendRequest) error {
	payments := make([]models.Payment, len(sendRequests))

	for i, sendRequest := range sendRequests {
		payments[i] = models.Payment{
			SendId:   sendRequest.ID,
			SendJson: sendRequest,
			PaidTo:   sendRequest.PaidTo,
		}
	}

	return tx.Create(&payments).Error
}

package repository

import (
	"github.com/bananocoin/boompow/apps/server/src/models"
	serializableModels "github.com/bananocoin/boompow/libs/models"
	"github.com/bananocoin/boompow/libs/utils/number"
	"gorm.io/gorm"
)

type PaymentRepo interface {
	BatchCreateSendRequests(tx *gorm.DB, sendRequests []serializableModels.SendRequest) error
	GetPendingPayments(tx *gorm.DB) ([]serializableModels.SendRequest, error)
	SetBlockHash(tx *gorm.DB, sendId string, blockHash string) error
	GetTotalPaidBanano() (float64, error)
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

// Get all payments with null block hash
func (s *PaymentService) GetPendingPayments(tx *gorm.DB) ([]serializableModels.SendRequest, error) {
	var res []serializableModels.SendRequest

	if err := tx.Model(&models.Payment{}).Select("send_json").Where("block_hash is null").Find(&res).Error; err != nil {
		return nil, err
	}

	return res, nil
}

// Update payment with block hash
func (s *PaymentService) SetBlockHash(tx *gorm.DB, sendId string, blockHash string) error {
	return tx.Model(&models.Payment{}).Where("send_id = ?", sendId).Update("block_hash", blockHash).Error
}

// Get total paid sum
func (s *PaymentService) GetTotalPaidBanano() (float64, error) {
	var totalPaid string
	if err := s.Db.Model(&models.Payment{}).Select("sum(cast(send_json->>'amount'as numeric)) as total_raw").Find(&totalPaid).Error; err != nil {
		return -1, err
	}
	asBan, err := number.RawToBanano(totalPaid, true)
	if err != nil {
		return -1, err
	}

	// Magic number, this is what BPoW v1 paid so add it to the total
	asBan += 1728016

	return asBan, nil
}

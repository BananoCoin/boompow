package models

import (
	"github.com/bananocoin/boompow/libs/models"
	"github.com/google/uuid"
)

// Store the payment request to users in the database
type Payment struct {
	Base
	BlockHash *string            `json:"block_hash" gorm:"uniqueIndex"`
	SendId    string             `json:"send_id" gorm:"uniqueIndex;not null"`
	Amount uint `json:"amount"`
	SendJson  models.SendRequest `json:"send_json" gorm:"type:jsonb;not null"`
	PaidTo    uuid.UUID          `json:"user_id" gorm:"not null"`
}

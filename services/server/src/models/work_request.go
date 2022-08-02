package models

import "github.com/google/uuid"

type WorkRequest struct {
	Base
	Hash        string    `json:"hash" gorm:"uniqueIndex;not null"`
	Difficulty  string    `json:"difficulty"`
	Result      string    `json:"result"`
	ProvidedBy  uuid.UUID `json:"providedBy"`
	RequestedBy uuid.UUID `json:"requestedBy"`
}

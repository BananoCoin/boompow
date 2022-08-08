package models

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/google/uuid"
)

// RPC requests
type BaseRequest struct {
	Action string `json:"action"`
}

// send
var SendAction BaseRequest = BaseRequest{Action: "send"}

type SendRequest struct {
	BaseRequest
	Wallet      string    `json:"wallet"`
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	AmountRaw   string    `json:"amount"`
	ID          string    `json:"id"`
	PaidTo      uuid.UUID `json:"-"`
}

type SendResponse struct {
	Block string `json:"block"`
}

// Type of nano payment object as JSONB
func (j SendRequest) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *SendRequest) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

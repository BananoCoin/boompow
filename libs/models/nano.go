package models

// RPC requests
type BaseRequest struct {
	Action string `json:"action"`
}

// send
var SendAction BaseRequest = BaseRequest{Action: "send"}

type SendRequest struct {
	BaseRequest
	Wallet      string `json:"wallet"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	AmountRaw   string `json:"amount"`
	ID          string `json:"id"`
}

type SendResponse struct {
	Block string `json:"block"`
}

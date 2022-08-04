package models

// Work response sent from client -> server
type ClientWorkResponse struct {
	RequestID string `json:"request_id"`
	Hash      string `json:"hash"`
	Result    string `json:"result"`
}

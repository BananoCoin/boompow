package models

// Work response sent from client -> server
type ClientWorkResponse struct {
	Hash   string `json:"hash"`
	Result string `json:"result"`
}

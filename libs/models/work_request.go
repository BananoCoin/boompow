package models

// Work request sent from server -> client
type ClientWorkRequest struct {
	Hash                 string `json:"hash"`
	DifficutlyMultiplier int    `json:"difficulty_multiplier"`
}

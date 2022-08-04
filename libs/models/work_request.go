package models

// Work request sent from server -> client
type ClientWorkRequest struct {
	// We attach a unique request ID to each request, this links it to user requesting work
	RequestID            string `json:"request_id"`
	Hash                 string `json:"hash"`
	DifficultyMultiplier int    `json:"difficulty_multiplier"`
}

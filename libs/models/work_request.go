package models

// Work request sent from server -> client
type ClientRequest struct {
	// Exclude this field from serialization (don't expose requester email to client)
	RequesterEmail string `json:"-"`
	RequestType    string `json:"request_type"`
	// We attach a unique request ID to each request, this links it to user requesting work
	RequestID            string `json:"request_id"`
	Hash                 string `json:"hash"`
	DifficultyMultiplier int    `json:"difficulty_multiplier"`
}

package models

type RequestType string

const (
	WorkGenerate RequestType = "work_generate"
	WorkCancel   RequestType = "work_cancel"
	BlockAwarded RequestType = "block_awarded"
)

// Message request sent from server -> client
type ClientRequest struct {
	// Exclude this field from serialization (don't expose requester email to client)
	RequesterEmail string      `json:"-"`
	RequestType    RequestType `json:"request_type"`
	// We attach a unique request ID to each request, this links it to user requesting work
	RequestID            string `json:"request_id"`
	Hash                 string `json:"hash"`
	DifficultyMultiplier int    `json:"difficulty_multiplier"`
}

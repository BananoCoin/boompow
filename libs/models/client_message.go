package models

type MessageType string

const (
	WorkGenerate MessageType = "work_generate"
	WorkCancel   MessageType = "work_cancel"
	BlockAwarded MessageType = "block_awarded"
)

// Message sent from server -> client
type ClientMessage struct {
	// Exclude this field from serialization (don't expose requester email to client)
	RequesterEmail string      `json:"-"`
	MessageType    MessageType `json:"request_type"`
	// We attach a unique request ID to each request, this links it to user requesting work
	RequestID            string `json:"request_id"`
	Hash                 string `json:"hash"`
	DifficultyMultiplier int    `json:"difficulty_multiplier"`
	// Awarded info
	ProviderEmail  string  `json:"-"`
	PercentOfPool  float64 `json:"percent_of_pool"`
	EstimatedAward float64 `json:"estimated_award"`
}

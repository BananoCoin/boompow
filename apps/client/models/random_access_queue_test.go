package models

import (
	"math/rand"
	"testing"

	serializableModels "github.com/bananocoin/boompow/libs/models"
	utils "github.com/bananocoin/boompow/libs/utils/testing"
)

// Test random access map
func TestRandomAccessMap(t *testing.T) {
	// Seed random for consistency
	rand.Seed(1)
	queue := NewRandomAccessQueue()

	// Add a few items
	queue.Put(serializableModels.ClientMessage{
		RequestID:            "1",
		Hash:                 "1",
		DifficultyMultiplier: 1,
	})
	queue.Put(serializableModels.ClientMessage{
		RequestID:            "2",
		Hash:                 "2",
		DifficultyMultiplier: 2,
	})
	queue.Put(serializableModels.ClientMessage{
		RequestID:            "3",
		Hash:                 "3",
		DifficultyMultiplier: 3,
	})

	// Check that we can access these items
	utils.AssertEqual(t, "1", queue.Get("1").Hash)

	// Check that we can pop a random item
	utils.AssertEqual(t, "3", queue.PopRandom().Hash)

	// Check that popped item is removed
	utils.AssertEqual(t, (*serializableModels.ClientMessage)(nil), queue.Get("3"))

	// Check length
	utils.AssertEqual(t, 2, queue.Len())
}

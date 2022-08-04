package tests

import (
	"math/rand"
	"testing"

	serializableModels "github.com/bbedward/boompow-ng/libs/models"
	utils "github.com/bbedward/boompow-ng/libs/utils/testing"
	"github.com/bbedward/boompow-ng/services/client/models"
)

// Test random access map
func TestRandomAccessMap(t *testing.T) {
	// Seed random for consistency
	rand.Seed(1)
	queue := models.NewRandomAccessQueue()

	// Add a few items
	queue.Put(serializableModels.ClientWorkRequest{
		RequestID:            "1",
		Hash:                 "1",
		DifficutlyMultiplier: 1,
	})
	queue.Put(serializableModels.ClientWorkRequest{
		RequestID:            "2",
		Hash:                 "2",
		DifficutlyMultiplier: 2,
	})
	queue.Put(serializableModels.ClientWorkRequest{
		RequestID:            "3",
		Hash:                 "3",
		DifficutlyMultiplier: 3,
	})

	// Check that we can access these items
	utils.AssertEqual(t, "1", queue.Get("1").Hash)

	// Check that we can pop a random item
	utils.AssertEqual(t, "3", queue.PopRandom().Hash)

	// Check that popped item is removed
	utils.AssertEqual(t, (*serializableModels.ClientWorkRequest)(nil), queue.Get("3"))

	// Check length
	utils.AssertEqual(t, 2, len(queue.Hashes))
}

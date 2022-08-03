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
	accessMap := models.NewRandomAccessMap()

	// Add a few items
	accessMap.Put(serializableModels.ClientWorkRequest{
		Hash:                 "1",
		DifficutlyMultiplier: 1,
	})
	accessMap.Put(serializableModels.ClientWorkRequest{
		Hash:                 "2",
		DifficutlyMultiplier: 2,
	})
	accessMap.Put(serializableModels.ClientWorkRequest{
		Hash:                 "3",
		DifficutlyMultiplier: 3,
	})

	// Check that we can access these items
	utils.AssertEqual(t, "1", accessMap.Get("1").Hash)

	// Check that we can pop a random item
	utils.AssertEqual(t, "3", accessMap.PopRandom().Hash)

	// Check that popped item is removed
	utils.AssertEqual(t, (*serializableModels.ClientWorkRequest)(nil), accessMap.Get("3"))

	// Check length
	utils.AssertEqual(t, 2, len(accessMap.Hashes))
}

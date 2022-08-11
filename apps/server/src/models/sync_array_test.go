package models

import (
	"testing"

	utils "github.com/bananocoin/boompow/libs/utils/testing"
)

// Test sync array
func TestSyncArray(t *testing.T) {
	array := NewSyncArray()

	// Add a few items
	array.Put(ActiveChannelObject{
		RequesterEmail:       "1",
		RequestID:            "1",
		Hash:                 "1",
		DifficultyMultiplier: 1,
		Chan:                 make(chan []byte),
	})
	array.Put(ActiveChannelObject{
		RequesterEmail:       "2",
		RequestID:            "2",
		Hash:                 "2",
		DifficultyMultiplier: 2,
		Chan:                 make(chan []byte),
	})
	array.Put(ActiveChannelObject{
		RequesterEmail:       "3",
		RequestID:            "3",
		Hash:                 "3",
		DifficultyMultiplier: 3,
		Chan:                 make(chan []byte),
	})

	utils.AssertEqual(t, 3, array.Len())
	utils.AssertEqual(t, true, array.Exists("1"))
	utils.AssertEqual(t, "1", array.Get("1").Hash)
	utils.AssertEqual(t, true, array.HashExists("2"))
	array.Delete("1")
	utils.AssertEqual(t, (*ActiveChannelObject)(nil), array.Get("1"))
	utils.AssertEqual(t, 2, array.Len())
	utils.AssertEqual(t, 0, array.IndexOf("3"))
}

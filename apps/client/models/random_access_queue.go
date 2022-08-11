package models

import (
	"math/rand"
	"sync"

	serializableModels "github.com/bananocoin/boompow/libs/models"
)

// RandomAccessQueue provides a data struture that can be access randomly
// This helps workers clears backlogs of work more evenly
// If there is 20 items on 3 workers, each worker will access the next unit of work randomly
type RandomAccessQueue struct {
	mu     sync.Mutex
	hashes []serializableModels.ClientMessage
}

func NewRandomAccessQueue() *RandomAccessQueue {
	return &RandomAccessQueue{
		hashes: []serializableModels.ClientMessage{},
	}
}

// See if element exists
func (r *RandomAccessQueue) exists(hash string) bool {
	for _, v := range r.hashes {
		if v.Hash == hash {
			return true
		}
	}
	return false
}

// Get length - synchronized
func (r *RandomAccessQueue) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.hashes)
}

// Put value into map - synchronized
func (r *RandomAccessQueue) Put(value serializableModels.ClientMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.exists(value.Hash) {
		r.hashes = append(r.hashes, value)
	}
}

// Removes and returns a random value from the map - synchronized
func (r *RandomAccessQueue) PopRandom() *serializableModels.ClientMessage {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.hashes) == 0 {
		return nil
	}
	index := rand.Intn(len(r.hashes))
	ret := r.hashes[index]
	r.hashes = remove(r.hashes, index)

	return &ret
}

// Gets a value from the map - synchronized
func (r *RandomAccessQueue) Get(hash string) *serializableModels.ClientMessage {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.exists(hash) {
		return &r.hashes[r.indexOf(hash)]
	}

	return nil
}

// Removes specified hash - synchronized
func (r *RandomAccessQueue) Delete(hash string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	index := r.indexOf(hash)
	if index > -1 {
		r.hashes = remove(r.hashes, r.indexOf(hash))
	}
}

func (r *RandomAccessQueue) indexOf(hash string) int {
	for i, v := range r.hashes {
		if v.Hash == hash {
			return i
		}
	}
	return -1
}

// NOT thread safe, must be called from within a locked section
func remove(s []serializableModels.ClientMessage, i int) []serializableModels.ClientMessage {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

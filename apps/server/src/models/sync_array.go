package models

import (
	"sync"
)

type ActiveChannelObject struct {
	BlockAward           bool
	RequesterEmail       string
	RequestID            string
	Hash                 string
	DifficultyMultiplier int
	Chan                 chan []byte
}

// SyncArray builds an thread-safe array with some handy methods
type SyncArray struct {
	mu       sync.Mutex
	channels []ActiveChannelObject
}

func NewSyncArray() *SyncArray {
	return &SyncArray{
		channels: []ActiveChannelObject{},
	}
}

// See if element exists
func (r *SyncArray) Exists(requestID string) bool {
	for _, v := range r.channels {
		if v.RequestID == requestID {
			return true
		}
	}
	return false
}

func (r *SyncArray) HashExists(hash string) bool {
	for _, v := range r.channels {
		if v.Hash == hash {
			return true
		}
	}
	return false
}

// Put value into map - synchronized
func (r *SyncArray) Put(value ActiveChannelObject) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.Exists(value.RequestID) {
		r.channels = append(r.channels, value)
	}
}

// Gets a value from the map - synchronized
func (r *SyncArray) Get(requestID string) *ActiveChannelObject {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Exists(requestID) {
		return &r.channels[r.IndexOf(requestID)]
	}

	return nil
}

// Removes specified hash - synchronized
func (r *SyncArray) Delete(requestID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	index := r.IndexOf(requestID)
	if index > -1 {
		r.channels = remove(r.channels, r.IndexOf(requestID))
	}
}

func (r *SyncArray) IndexOf(requestID string) int {
	for i, v := range r.channels {
		if v.RequestID == requestID {
			return i
		}
	}
	return -1
}

func (r *SyncArray) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.channels)
}

// NOT thread safe, must be called from within a locked section
func remove(s []ActiveChannelObject, i int) []ActiveChannelObject {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

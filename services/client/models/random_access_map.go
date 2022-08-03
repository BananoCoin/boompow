package models

import (
	"math/rand"
	"sync"

	serializableModels "github.com/bbedward/boompow-ng/libs/models"
)

// RandomAccessMap provides a data struture that can be access randomly
// This helps workers clears backlogs of work more evenly
// If there is 20 items on 3 workers, each worker will access the next unit of work randomly
type RandomAccessMap struct {
	Data   sync.Map
	mu     sync.Mutex
	Hashes []string
}

func NewRandomAccessMap() *RandomAccessMap {
	return &RandomAccessMap{
		Data:   sync.Map{},
		Hashes: []string{},
	}
}

func (r *RandomAccessMap) Put(key string, value serializableModels.ClientWorkRequest) {
	r.Data.Store(key, value)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Hashes = append(r.Hashes, key)
}

// Removes and returns a random value from the map
func (r *RandomAccessMap) PopRandom() *serializableModels.ClientWorkRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	index := rand.Intn(len(r.Hashes))
	val, ok := r.Data.Load(r.Hashes[index])
	if !ok {
		return nil
	}
	ret := val.(serializableModels.ClientWorkRequest)
	r.Data.Delete(r.Hashes[index])
	r.Hashes = remove(r.Hashes, index)

	return &ret
}

// Gets a value from the map
func (r *RandomAccessMap) Get(hash string) *serializableModels.ClientWorkRequest {
	val, ok := r.Data.Load(hash)
	if !ok {
		return nil
	}
	ret := val.(serializableModels.ClientWorkRequest)

	return &ret
}

// Removes specified hash
func (r *RandomAccessMap) Delete(hash string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Data.Delete(hash)
	index := r.IndexOf(hash)
	if index > -1 {
		r.Hashes = remove(r.Hashes, r.IndexOf(hash))
	}
}

func (r *RandomAccessMap) IndexOf(hash string) int {
	for i, v := range r.Hashes {
		if v == hash {
			return i
		}
	}
	return -1
}

func (r *RandomAccessMap) DataLen() int {
	var i int
	r.Data.Range(func(k, v interface{}) bool {
		i++
		return true
	})
	return i
}

func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

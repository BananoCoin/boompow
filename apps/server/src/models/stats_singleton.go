package models

import (
	"sync"

	"github.com/bananocoin/boompow/apps/server/graph/model"
)

var lock = &sync.Mutex{}

type StatsSingleton struct {
	Stats *model.Stats
}

var statsInstance *StatsSingleton

func GetStatsInstance() *StatsSingleton {
	if statsInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if statsInstance == nil {
			statsInstance = &StatsSingleton{}
		}
	}

	return statsInstance
}

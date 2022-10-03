package graph

import (
	"sync"

	"github.com/bananocoin/boompow/apps/server/src/repository"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	UserRepo    repository.UserRepo
	WorkRepo    repository.WorkRepo
	PaymentRepo repository.PaymentRepo
	PrecacheMap *sync.Map
}

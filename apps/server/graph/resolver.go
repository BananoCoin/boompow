package graph

import "github.com/bananocoin/boompow/apps/server/src/repository"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	UserRepo repository.UserRepo
	WorkRepo repository.WorkRepo
}

package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/bbedward/boompow-server-ng/graph/generated"
	"github.com/bbedward/boompow-server-ng/graph/model"
	"github.com/bbedward/boompow-server-ng/src/utils"
	"github.com/google/uuid"
)

// CreateUser is the resolver for the createUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, input model.UserInput) (*model.User, error) {
	user, err := r.UserRepo.CreateUser(&input)
	userCreated := &model.User{
		Username:  user.Username,
		ID:        user.ID.String(),
		CreatedAt: utils.GenerateISOString(user.CreatedAt),
		UpdatedAt: utils.GenerateISOString(user.UpdatedAt),
	}
	if err != nil {
		return nil, err
	}
	return userCreated, nil
}

// DeleteUser is the resolver for the DeleteUser field.
func (r *mutationResolver) DeleteUser(ctx context.Context, id string) (string, error) {
	uuid, err := uuid.Parse((id))
	if err != nil {
		return "", err
	}

	err = r.UserRepo.DeleteUser(uuid)
	if err != nil {
		return "", err
	}
	successMessage := fmt.Sprintf("deleted %s", id)
	return successMessage, nil
}

// UpdateUser is the resolver for the UpdateUser field.
func (r *mutationResolver) UpdateUser(ctx context.Context, id string, input model.UserInput) (string, error) {
	uuid, err := uuid.Parse((id))
	if err != nil {
		return "", err
	}

	err = r.UserRepo.UpdateUser(&input, uuid)
	if err != nil {
		return "nil", err
	}

	successMessage := fmt.Sprintf("updated %s", id)
	return successMessage, nil
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, input model.Login) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

// RefreshToken is the resolver for the refreshToken field.
func (r *mutationResolver) RefreshToken(ctx context.Context, input model.RefreshTokenInput) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

// GetAllUsers is the resolver for the GetAllUsers field.
func (r *queryResolver) GetAllUsers(ctx context.Context) ([]*model.User, error) {
	users, err := r.UserRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}

	// Create return for GQL using primitive types
	var gqlUsers []*model.User
	for _, user := range users {
		gqlUsers = append(gqlUsers, &model.User{
			ID:        user.ID.String(),
			Username:  user.Username,
			CreatedAt: utils.GenerateISOString(user.CreatedAt),
			UpdatedAt: utils.GenerateISOString(user.UpdatedAt),
		})
	}
	return gqlUsers, nil
}

// GetOneUser is the resolver for the GetOneUser field.
func (r *queryResolver) GetOneUser(ctx context.Context, id string) (*model.User, error) {
	uuid, err := uuid.Parse((id))
	if err != nil {
		return nil, err
	}

	user, err := r.UserRepo.GetOneUser(uuid)
	selectedUser := &model.User{
		ID:        user.ID.String(),
		Username:  user.Username,
		CreatedAt: utils.GenerateISOString(user.CreatedAt),
		UpdatedAt: utils.GenerateISOString(user.UpdatedAt),
	}
	if err != nil {
		return nil, err
	}
	return selectedUser, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	"github.com/bbedward/boompow-server-ng/graph/generated"
	"github.com/bbedward/boompow-server-ng/graph/model"
	"github.com/bbedward/boompow-server-ng/src/middleware"
	"github.com/bbedward/boompow-server-ng/src/models"
	"github.com/bbedward/boompow-server-ng/src/utils/auth"
	utils "github.com/bbedward/boompow-server-ng/src/utils/format"
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
	correct := r.UserRepo.Authenticate(&input)
	if !correct {
		// 1
		return "", errors.New("invalid username or password")
	}
	token, err := auth.GenerateToken(input.Username)
	if err != nil {
		return "", err
	}
	return token, nil
}

// RefreshToken is the resolver for the refreshToken field.
func (r *mutationResolver) RefreshToken(ctx context.Context, input model.RefreshTokenInput) (string, error) {
	username, err := auth.ParseToken(input.Token)
	if err != nil {
		return "", fmt.Errorf("access denied")
	}
	token, err := auth.GenerateToken(username)
	if err != nil {
		return "", err
	}
	return token, nil
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

// GetUser is the resolver for the getUser field.
func (r *queryResolver) GetUser(ctx context.Context, id *string, username *string) (*model.User, error) {
	var err error
	var user *models.User

	// ! TODO - remove me, test for authentication
	user = middleware.ForContext(ctx)
	if user == nil {
		return &model.User{}, fmt.Errorf("access denied")
	}

	if id != nil {
		userID, err := uuid.Parse(*id)
		if err != nil {
			return nil, err
		}
		user, err = r.UserRepo.GetUser(&userID, nil)
	}
	if username != nil {
		user, err = r.UserRepo.GetUser(nil, username)
	}

	if err != nil {
		return nil, err
	}

	selectedUser := &model.User{
		ID:        user.ID.String(),
		Username:  user.Username,
		CreatedAt: utils.GenerateISOString(user.CreatedAt),
		UpdatedAt: utils.GenerateISOString(user.UpdatedAt),
	}
	return selectedUser, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

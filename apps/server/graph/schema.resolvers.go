package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/bananocoin/boompow-next/apps/server/graph/generated"
	"github.com/bananocoin/boompow-next/apps/server/graph/model"
	"github.com/bananocoin/boompow-next/apps/server/src/config"
	"github.com/bananocoin/boompow-next/apps/server/src/controller"
	"github.com/bananocoin/boompow-next/apps/server/src/database"
	"github.com/bananocoin/boompow-next/apps/server/src/middleware"
	"github.com/bananocoin/boompow-next/apps/server/src/models"
	serializableModels "github.com/bananocoin/boompow-next/libs/models"
	"github.com/bananocoin/boompow-next/libs/utils/auth"
	utils "github.com/bananocoin/boompow-next/libs/utils/format"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateUser is the resolver for the createUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, input model.UserInput) (*model.User, error) {
	user, err := r.UserRepo.CreateUser(&input, true)
	if err != nil {
		return nil, err
	}
	userCreated := &model.User{
		Email:      user.Email,
		ID:         user.ID.String(),
		CreatedAt:  utils.GenerateISOString(user.CreatedAt),
		UpdatedAt:  utils.GenerateISOString(user.UpdatedAt),
		Type:       input.Type,
		BanAddress: input.BanAddress,
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

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.LoginResponse, error) {
	correct := r.UserRepo.Authenticate(&input)
	if !correct {
		return nil, errors.New("invalid email or password")
	}
	token, err := auth.GenerateToken(input.Email, time.Now)
	if err != nil {
		return nil, err
	}
	return &model.LoginResponse{
		Token: token,
	}, nil
}

// RefreshToken is the resolver for the refreshToken field.
func (r *mutationResolver) RefreshToken(ctx context.Context, input model.RefreshTokenInput) (string, error) {
	email, err := auth.ParseToken(input.Token)
	if err != nil {
		return "", fmt.Errorf("access denied")
	}
	token, err := auth.GenerateToken(email, time.Now)
	if err != nil {
		return "", err
	}
	return token, nil
}

// WorkGenerate is the resolver for the workGenerate field.
func (r *mutationResolver) WorkGenerate(ctx context.Context, input model.WorkGenerateInput) (string, error) {
	// Require authentication for service
	requester := middleware.AuthorizedServiceToken(ctx)
	if requester == nil {
		return "", fmt.Errorf("access denied")
	}

	reqID := make([]byte, 32)
	if _, err := rand.Read(reqID); err != nil {
		return "", errors.New("server_error:error occured processing request")
	}

	// Check that this request is valid
	_, err := hex.DecodeString(input.Hash)
	if err != nil || len(input.Hash) != 64 {
		return "", errors.New("bad_request:invalid hash")
	}

	// Alter our difficulty to be in a valid range if it isn't
	if input.DifficultyMultiplier < 1 {
		// 1 is NANO receive and banano base difficulty
		input.DifficultyMultiplier = 1
	} else if input.DifficultyMultiplier > config.MAX_WORK_DIFFICULTY_MULTIPLIER {
		input.DifficultyMultiplier = config.MAX_WORK_DIFFICULTY_MULTIPLIER
	}

	// First try to retrieve from cache
	// We only want cached results that meet the required difficulty
	workResult, err := r.WorkRepo.RetrieveWorkFromCache(input.Hash, input.DifficultyMultiplier)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	if workResult != nil {
		return workResult.Result, nil
	}

	workRequest := &serializableModels.ClientMessage{
		RequesterEmail:       requester.User.Email,
		MessageType:          serializableModels.WorkGenerate,
		RequestID:            hex.EncodeToString(reqID),
		Hash:                 input.Hash,
		DifficultyMultiplier: input.DifficultyMultiplier,
	}

	resp, err := controller.BroadcastWorkRequestAndWait(workRequest)
	if err != nil {
		return "", err
	}

	return resp.Result, nil
}

// GenerateServiceToken is the resolver for the generateServiceToken field.
func (r *mutationResolver) GenerateServiceToken(ctx context.Context) (string, error) {
	// Require authentication
	requester := middleware.AuthorizedRequester(ctx)
	if requester == nil {
		return "", fmt.Errorf("access denied")
	}

	// Generate token
	token := r.UserRepo.GenerateServiceToken()

	if err := database.GetRedisDB().AddServiceToken(requester.User.ID, token); err != nil {
		return "", fmt.Errorf("error generating token")
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
			Email:     user.Email,
			CreatedAt: utils.GenerateISOString(user.CreatedAt),
			UpdatedAt: utils.GenerateISOString(user.UpdatedAt),
			Type:      model.UserType(user.Type),
		})
	}
	return gqlUsers, nil
}

// GetUser is the resolver for the getUser field.
func (r *queryResolver) GetUser(ctx context.Context, id *string, email *string) (*model.User, error) {
	var err error
	var user *models.User

	if id != nil {
		userID, err := uuid.Parse(*id)
		if err != nil {
			return nil, err
		}
		user, err = r.UserRepo.GetUser(&userID, nil)
	}
	if email != nil {
		user, err = r.UserRepo.GetUser(nil, email)
	}

	if err != nil {
		return nil, err
	}

	selectedUser := &model.User{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: utils.GenerateISOString(user.CreatedAt),
		UpdatedAt: utils.GenerateISOString(user.UpdatedAt),
	}
	return selectedUser, nil
}

// VerifyEmail is the resolver for the verifyEmail field.
func (r *queryResolver) VerifyEmail(ctx context.Context, input model.VerifyEmailInput) (bool, error) {
	return r.UserRepo.VerifyEmailToken(&input)
}

// Stats is the resolver for the stats field.
func (r *subscriptionResolver) Stats(ctx context.Context) (<-chan *model.Stats, error) {
	msgs := make(chan *model.Stats, 1)

	// Pub stats every 10 seconds
	go func() {
		for {
			nConnectedClients, err := database.GetRedisDB().GetNumberConnectedClients()
			if err == nil {
				msgs <- &model.Stats{ConnectedWorkers: int(nConnectedClients)}
			}
			time.Sleep(10 * time.Second)
		}
	}()
	return msgs, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }

package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bananocoin/boompow/apps/server/graph/generated"
	"github.com/bananocoin/boompow/apps/server/graph/model"
	"github.com/bananocoin/boompow/apps/server/src/config"
	"github.com/bananocoin/boompow/apps/server/src/controller"
	"github.com/bananocoin/boompow/apps/server/src/database"
	"github.com/bananocoin/boompow/apps/server/src/middleware"
	serializableModels "github.com/bananocoin/boompow/libs/models"
	"github.com/bananocoin/boompow/libs/utils/auth"
	utils "github.com/bananocoin/boompow/libs/utils/format"
	"github.com/bananocoin/boompow/libs/utils/validation"
	"github.com/google/uuid"
	"gorm.io/gorm"
	klog "k8s.io/klog/v2"
)

// CreateUser is the resolver for the createUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, input model.UserInput) (*model.User, error) {
	return nil, errors.New("Registrations disabled")
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

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.LoginResponse, error) {
	user := r.UserRepo.Authenticate(&input)
	if user == nil {
		return nil, errors.New("invalid email or password")
	}
	token, err := auth.GenerateToken(strings.ToLower(input.Email), time.Now)
	if err != nil {
		return nil, err
	}
	return &model.LoginResponse{
		Token:          token,
		Type:           model.UserType(user.Type),
		BanAddress:     user.BanAddress,
		ServiceName:    user.ServiceName,
		ServiceWebsite: user.ServiceWebsite,
		EmailVerified:  user.EmailVerified,
		Email:          user.Email,
	}, nil
}

// RefreshToken is the resolver for the refreshToken field.
func (r *mutationResolver) RefreshToken(ctx context.Context, input model.RefreshTokenInput) (string, error) {
	email, err := auth.ParseToken(input.Token)
	if err != nil {
		return "", fmt.Errorf("access denied")
	}
	token, err := auth.GenerateToken(strings.ToLower(email), time.Now)
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
	if workResult != "" {
		return workResult, nil
	}

	workRequest := serializableModels.ClientMessage{
		RequesterEmail:       requester.User.Email,
		BlockAward:           input.BlockAward == nil || *input.BlockAward,
		MessageType:          serializableModels.WorkGenerate,
		RequestID:            uuid.NewString(),
		Hash:                 input.Hash,
		DifficultyMultiplier: input.DifficultyMultiplier,
	}

	resp, err := controller.BroadcastWorkRequestAndWait(workRequest)
	if err != nil {
		return "", err
	}

	r.PrecacheMap.Store(strings.ToUpper(input.Hash), resp.Result)

	return resp.Result, nil
}

// GenerateOrGetServiceToken is the resolver for the generateOrGetServiceToken field.
func (r *mutationResolver) GenerateOrGetServiceToken(ctx context.Context) (string, error) {
	// Require authentication
	requester := middleware.AuthorizedRequester(ctx)
	if requester == nil {
		return "", fmt.Errorf("access denied")
	}

	// Get token
	token, err := database.GetRedisDB().GetServiceTokenForUser(requester.User.ID)
	if err != nil {
		// Generate token
		token = r.UserRepo.GenerateServiceToken()

		if err := database.GetRedisDB().AddServiceToken(requester.User.ID, token); err != nil {
			return "", fmt.Errorf("error generating token")
		}
	}

	return token, nil
}

// ResetPassword is the resolver for the resetPassword field.
func (r *mutationResolver) ResetPassword(ctx context.Context, input model.ResetPasswordInput) (bool, error) {
	return false, errors.New("Password reset disabled")
	r.UserRepo.GenerateResetPasswordRequest(&input, true)

	return true, nil
}

// ResendConfirmationEmail is the resolver for the resendConfirmationEmail field.
func (r *mutationResolver) ResendConfirmationEmail(ctx context.Context, input model.ResendConfirmationEmailInput) (bool, error) {
	return false, errors.New("Email confirmation disabled")
	u, err := r.UserRepo.GetUser(nil, &input.Email)
	if err != nil {
		return false, errors.New("User does not exist")
	}
	if u.EmailVerified {
		return false, errors.New("Email is already verified")
	}

	if err = r.UserRepo.SendConfirmEmailEmail(u.Email, u.Type, true); err != nil {
		return false, err
	}
	return true, nil
}

// SendConfirmationEmail is the resolver for the sendConfirmationEmail field.
func (r *mutationResolver) SendConfirmationEmail(ctx context.Context) (bool, error) {
	return false, errors.New("Email confirmation disabled")
	// Require authentication
	user := middleware.AuthorizedUser(ctx)
	if user == nil {
		return false, fmt.Errorf("access denied")
	}

	if user.User.EmailVerified {
		return false, fmt.Errorf("already verified")
	}

	if err := r.UserRepo.SendConfirmEmailEmail(user.User.Email, user.User.Type, true); err != nil {
		return false, fmt.Errorf("error sending email")
	}

	return true, nil
}

// ChangePassword is the resolver for the changePassword field.
func (r *mutationResolver) ChangePassword(ctx context.Context, input model.ChangePasswordInput) (bool, error) {
	return false, errors.New("Password reset disabled")
	// Require authentication for service
	requester := middleware.AuthorizedChangePassword(ctx)
	if requester == nil {
		return false, fmt.Errorf("access denied")
	}

	// Check that the password is valid
	err := validation.ValidatePassword(input.NewPassword)
	if err != nil {
		return false, err
	}

	// Is valid so update it
	if err := r.UserRepo.ChangePassword(requester.User.Email, &input); err == nil {
		return true, nil
	}

	return false, err
}

// VerifyEmail is the resolver for the verifyEmail field.
func (r *queryResolver) VerifyEmail(ctx context.Context, input model.VerifyEmailInput) (bool, error) {
	return false, errors.New("Email confirmation disabled")
	return r.UserRepo.VerifyEmailToken(&input)
}

// VerifyService is the resolver for the verifyService field.
func (r *queryResolver) VerifyService(ctx context.Context, input model.VerifyServiceInput) (bool, error) {
	return false, errors.New("Service verification disabled")
	return r.UserRepo.VerifyService(&input)
}

// GetUser is the resolver for the getUser field.
func (r *queryResolver) GetUser(ctx context.Context) (*model.GetUserResponse, error) {
	// Require authentication
	user := middleware.AuthorizedUser(ctx)
	if user == nil {
		return nil, fmt.Errorf("access denied")
	}
	return &model.GetUserResponse{
		Type:           model.UserType(user.User.Type),
		BanAddress:     user.User.BanAddress,
		ServiceName:    user.User.ServiceName,
		ServiceWebsite: user.User.ServiceWebsite,
		EmailVerified:  user.User.EmailVerified,
		Email:          user.User.Email,
		CanRequestWork: user.User.CanRequestWork,
	}, nil
}

// Stats is the resolver for the stats field.
func (r *subscriptionResolver) Stats(ctx context.Context) (<-chan *model.Stats, error) {
	msgs := make(chan *model.Stats, 1)

	// Pub stats every 10 seconds
	go func() {
		for {
			// Connected clients
			nConnectedClients, err := database.GetRedisDB().GetNumberConnectedClients()
			if err != nil {
				klog.Infof("Error retrieving connected clients for stats sub %v", err)
				continue
			}
			// Services
			services, err := r.WorkRepo.GetServiceStats()
			if err != nil {
				klog.Infof("Error retrieving services for stats sub %v", err)
				continue
			}
			var serviceStats []*model.StatsServiceType
			for _, service := range services {
				serviceStats = append(serviceStats, &model.StatsServiceType{
					Name:     service.ServiceName,
					Website:  service.ServiceWebsite,
					Requests: service.TotalRequests,
				})
			}
			// Top 10
			top10, err := r.WorkRepo.GetTopContributors(100)
			if err != nil {
				klog.Infof("Error retrieving # services for stats sub %v", err)
				continue
			}
			var top10Contributors []*model.StatsUserType
			for _, u := range top10 {
				top10Contributors = append(top10Contributors, &model.StatsUserType{
					BanAddress:      u.BanAddress,
					TotalPaidBanano: u.TotalBan,
				})
			}
			// Total paid
			totalPaidBan, err := r.PaymentRepo.GetTotalPaidBanano()
			if err == nil {
				msgs <- &model.Stats{ConnectedWorkers: int(nConnectedClients), TotalPaidBanano: fmt.Sprintf("%.2f", totalPaidBan), RegisteredServiceCount: len(services), Top10: top10Contributors, Services: serviceStats}
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

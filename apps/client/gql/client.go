package gql

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Khan/genqlient/graphql"
)

type GQLError string

const (
	InvalidUsernamePasssword GQLError = "Invalid username or password"
	ServerError                       = "Unknown server error, try again later"
)

var client graphql.Client

func InitGQLClient(url string) {
	client = graphql.NewClient(url, http.DefaultClient)
}

func RegisterProvider(ctx context.Context, email string, password string, banAddress string) (*createUserResponse, error) {
	resp, err := createUser(ctx, client, UserInput{
		Email:      email,
		Password:   password,
		Type:       UserTypeProvider,
		BanAddress: banAddress,
	})

	if err != nil {
		fmt.Printf("Error creating user in %v", err)
		return nil, err
	}

	return resp, nil
}

func RegisterService(ctx context.Context, email string, password string, serviceName string, serviceWebsite string) (*createUserResponse, error) {
	resp, err := createUser(ctx, client, UserInput{
		Email:          email,
		Password:       password,
		Type:           UserTypeRequester,
		ServiceName:    serviceName,
		ServiceWebsite: serviceWebsite,
	})

	if err != nil {
		fmt.Printf("Error creating user in %v", err)
		return nil, err
	}

	return resp, nil
}

func Login(ctx context.Context, email string, password string) (*loginUserResponse, GQLError) {
	resp, err := loginUser(ctx, client, LoginInput{
		Email:    email,
		Password: password,
	})

	if err != nil {
		fmt.Printf("Error logging in %v", err)
		if strings.Contains(err.Error(), "invalid email or password") {
			return nil, InvalidUsernamePasssword
		}
		return nil, ServerError
	}

	return resp, ""
}

func RefreshToken(ctx context.Context, token string) (string, error) {
	resp, err := refreshToken(ctx, client, RefreshTokenInput{
		Token: token,
	})

	if err != nil {
		fmt.Printf("\nError refreshing authentication token! You may need to restart the client and re-login %v", err)
		return "", err
	}
	fmt.Printf("\n👮 Refreshed authentication token")

	return resp.RefreshToken, nil
}

func ResendConfirmationEmail(ctx context.Context, email string) (*resendConfirmationEmailResponse, error) {
	resp, err := resendConfirmationEmail(ctx, client, ResendConfirmationEmailInput{
		Email: email,
	})

	if err != nil {
		fmt.Printf("Error resending email %v", err)
		return nil, err
	}

	return resp, nil
}

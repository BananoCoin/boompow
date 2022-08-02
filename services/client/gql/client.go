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

func InitGQLClient() {
	client = graphql.NewClient("http://localhost:8080/graphql", http.DefaultClient)
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
		fmt.Printf("Error refreshing authentication token! You may need to restart the client and re-login %v", err)
		return "", err
	}
	fmt.Printf("👮 Refreshed authentication token\n")

	return resp.RefreshToken, nil
}

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

type authedTransport struct {
	wrapped http.RoundTripper
	token   string
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", t.token)
	return t.wrapped.RoundTrip(req)
}

func InitGQLClient(url string) {
	client = graphql.NewClient(url, http.DefaultClient)
}

func InitGQLClientWithToken(url string, token string) {
	client = graphql.NewClient(url, &http.Client{Transport: &authedTransport{wrapped: http.DefaultTransport, token: token}})
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

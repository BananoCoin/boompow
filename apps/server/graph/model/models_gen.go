// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
)

type ChangePasswordInput struct {
	NewPassword string `json:"newPassword"`
}

type GetUserResponse struct {
	Email          string   `json:"email"`
	Type           UserType `json:"type"`
	BanAddress     *string  `json:"banAddress"`
	ServiceName    *string  `json:"serviceName"`
	ServiceWebsite *string  `json:"serviceWebsite"`
	EmailVerified  bool     `json:"emailVerified"`
	CanRequestWork bool     `json:"canRequestWork"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token          string   `json:"token"`
	Email          string   `json:"email"`
	Type           UserType `json:"type"`
	BanAddress     *string  `json:"banAddress"`
	ServiceName    *string  `json:"serviceName"`
	ServiceWebsite *string  `json:"serviceWebsite"`
	EmailVerified  bool     `json:"emailVerified"`
}

type RefreshTokenInput struct {
	Token string `json:"token"`
}

type ResendConfirmationEmailInput struct {
	Email string `json:"email"`
}

type ResetPasswordInput struct {
	Email string `json:"email"`
}

type Stats struct {
	ConnectedWorkers       int                 `json:"connectedWorkers"`
	TotalPaidBanano        string              `json:"totalPaidBanano"`
	RegisteredServiceCount int                 `json:"registeredServiceCount"`
	Top10                  []*StatsUserType    `json:"top10"`
	Services               []*StatsServiceType `json:"services"`
}

type StatsServiceType struct {
	Name     string `json:"name"`
	Website  string `json:"website"`
	Requests int    `json:"requests"`
}

type StatsUserType struct {
	BanAddress      string `json:"banAddress"`
	TotalPaidBanano string `json:"totalPaidBanano"`
}

type User struct {
	ID         string   `json:"id"`
	Email      string   `json:"email"`
	CreatedAt  string   `json:"createdAt"`
	UpdatedAt  string   `json:"updatedAt"`
	Type       UserType `json:"type"`
	BanAddress *string  `json:"banAddress"`
}

type UserInput struct {
	Email          string   `json:"email"`
	Password       string   `json:"password"`
	Type           UserType `json:"type"`
	BanAddress     *string  `json:"banAddress"`
	ServiceName    *string  `json:"serviceName"`
	ServiceWebsite *string  `json:"serviceWebsite"`
}

type VerifyEmailInput struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

type VerifyServiceInput struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

type WorkGenerateInput struct {
	Hash                 string `json:"hash"`
	DifficultyMultiplier int    `json:"difficultyMultiplier"`
	BlockAward           *bool  `json:"blockAward"`
}

type UserType string

const (
	UserTypeProvider  UserType = "PROVIDER"
	UserTypeRequester UserType = "REQUESTER"
)

var AllUserType = []UserType{
	UserTypeProvider,
	UserTypeRequester,
}

func (e UserType) IsValid() bool {
	switch e {
	case UserTypeProvider, UserTypeRequester:
		return true
	}
	return false
}

func (e UserType) String() string {
	return string(e)
}

func (e *UserType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = UserType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid UserType", str)
	}
	return nil
}

func (e UserType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

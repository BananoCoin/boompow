package auth

import (
	"testing"

	utils "github.com/bananocoin/boompow/libs/utils/testing"
)

func TestHashPassword(t *testing.T) {
	password := "password"
	hash, _ := HashPassword(password)
	utils.AssertEqual(t, true, CheckPasswordHash(password, hash))
	utils.AssertEqual(t, false, CheckPasswordHash("password1", hash))
}

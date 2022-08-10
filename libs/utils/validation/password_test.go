package validation

import (
	"strings"
	"testing"

	utils "github.com/bananocoin/boompow/libs/utils/testing"
	"github.com/stretchr/testify/assert"
)

func TestPasswordValidation(t *testing.T) {
	goodPass := "Password123!"

	utils.AssertEqual(t, nil, ValidatePassword(goodPass))
	// All lower/upper are invalid
	assert.EqualErrorf(t, ValidatePassword(strings.ToLower(goodPass)), "password must have at least one upper case character", "Error should be: %v, got: %v", "password must have at least one upper case character", ValidatePassword(strings.ToLower(goodPass)))
	assert.EqualErrorf(t, ValidatePassword(strings.ToUpper(goodPass)), "password must have at least one lower case character", "Error should be: %v, got: %v", "password must have at least one lower case character", ValidatePassword(strings.ToUpper(goodPass)))
	// No number is invalid
	assert.EqualErrorf(t, ValidatePassword("Password!"), "password must have at least one numeric character", "Error should be: %v, got: %v", "password must have at least one numeric character", ValidatePassword("Password!"))
	// No special character is invalid
	assert.EqualErrorf(t, ValidatePassword("Password123"), "password must have at least one special character", "Error should be: %v, got: %v", "password must have at least one special character", ValidatePassword("Password123"))
	// 8 in length
	assert.EqualErrorf(t, ValidatePassword("Passwor"), "password must be at least 8 characters long", "Error should be: %v, got: %v", "password must be at least 8 characters long", ValidatePassword("Password"))
}

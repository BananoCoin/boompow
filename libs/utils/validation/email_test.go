package validation

import (
	"testing"

	utils "github.com/bananocoin/boompow/libs/utils/testing"
)

func TestEmailValidation(t *testing.T) {
	utils.AssertEqual(t, true, IsValidEmail("helloworld@gmail.com"))
	utils.AssertEqual(t, false, IsValidEmail("helloworldgmail.com"))
}

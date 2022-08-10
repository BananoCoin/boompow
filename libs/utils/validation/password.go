package validation

import (
	"fmt"
	"unicode"
)

// 8 characters long, at least one lowercase letter, one uppercase letter, one number, one special character
func ValidatePassword(p string) error {
	if len(p) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
next:
	for name, classes := range map[string][]*unicode.RangeTable{
		"upper case": {unicode.Upper, unicode.Title},
		"lower case": {unicode.Lower},
		"numeric":    {unicode.Number, unicode.Digit},
		"special":    {unicode.Space, unicode.Symbol, unicode.Punct, unicode.Mark},
	} {
		for _, r := range p {
			if unicode.IsOneOf(classes, r) {
				continue next
			}
		}
		return fmt.Errorf("password must have at least one %s character", name)
	}
	return nil
}

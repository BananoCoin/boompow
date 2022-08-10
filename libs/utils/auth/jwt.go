package auth

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"github.com/bananocoin/boompow/libs/utils"
	"github.com/dgrijalva/jwt-go"
)

// secret key being used to sign tokens
var (
	SecretKey = utils.GetJwtKey()
)

// GenerateToken generates a jwt token and assign a email to it's claims and return it
func GenerateToken(email string, nowFunc func() time.Time) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	/* Create a map to store our claims */
	claims := token.Claims.(jwt.MapClaims)
	/* Set token claims */
	claims["email"] = email
	claims["exp"] = nowFunc().Add(time.Hour * 24).Unix()
	tokenString, err := token.SignedString(SecretKey)
	if err != nil {
		log.Fatal("Error in Generating key")
		return "", err
	}
	return tokenString, nil
}

// ParseToken parses a jwt token and returns the email in it's claims
func ParseToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		email := claims["email"].(string)
		return email, nil
	} else {
		return "", err
	}
}

// Generate random 32-byte hex string
func GenerateRandHexString() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

package auth

import (
	"encoding/hex"
	"os"
	"testing"
	"time"

	utils "github.com/bananocoin/boompow-next/libs/utils/testing"
)

var now = func() time.Time {
	return time.Unix(0, 0)
}

func TestGenerateToken(t *testing.T) {
	os.Setenv("PRIV_KEY", "value")
	defer os.Unsetenv("PRIV_KEY")
	token, _ := GenerateToken("joe@gmail.com", now)
	utils.AssertEqual(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImpvZUBnbWFpbC5jb20iLCJleHAiOjg2NDAwfQ.4EWNyTndi4_6yT8JlA9RWjVIC6p2BiKAJx3BHGA4qYM", token)
}

func TestParseToken(t *testing.T) {
	os.Setenv("PRIV_KEY", "value")
	defer os.Unsetenv("PRIV_KEY")
	token, _ := GenerateToken("joe@gmail.com", time.Now)
	parsed, _ := ParseToken(token)
	utils.AssertEqual(t, "joe@gmail.com", parsed)
}

func TestGenerateRandHexString(t *testing.T) {
	gen, _ := GenerateRandHexString()
	parsed, err := hex.DecodeString(gen)

	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, 64, len(gen))
	utils.AssertEqual(t, 32, len(parsed))
}

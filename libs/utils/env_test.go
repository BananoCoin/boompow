package utils

import (
	"os"
	"testing"

	utils "github.com/bananocoin/boompow-next/libs/utils/testing"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("MY_ENV", "value")
	defer os.Unsetenv("MY_ENV")

	utils.AssertEqual(t, "value", GetEnv("MY_ENV", "default"))
	utils.AssertEqual(t, "default", GetEnv("MY_ENV_UNKNOWN", "default"))
}

func TestGetJwtKey(t *testing.T) {
	os.Unsetenv("PRIV_KEY")
	utils.AssertEqual(t, []byte("badKey"), GetJwtKey())

	os.Setenv("PRIV_KEY", "X")
	defer os.Unsetenv("PRIV_KEY")
	utils.AssertEqual(t, []byte("X"), GetJwtKey())
}

func TestGetSmtpConnInformation(t *testing.T) {
	os.Setenv("SMTP_SERVER", "")
	os.Setenv("SMTP_PORT", "-1")
	os.Setenv("SMTP_USERNAME", "")
	os.Setenv("SMTP_PASSWORD", "")
	defer os.Unsetenv("SMTP_SERVER")
	defer os.Unsetenv("SMTP_PORT")
	defer os.Unsetenv("SMTP_USERNAME")
	defer os.Unsetenv("SMTP_PASSWORD")

	utils.AssertEqual(t, (*SmtpConnInformation)(nil), GetSmtpConnInformation())

	os.Setenv("SMTP_SERVER", "abc.com")
	os.Setenv("SMTP_PORT", "1234")
	os.Setenv("SMTP_USERNAME", "joe")
	os.Setenv("SMTP_PASSWORD", "jeff")

	connInfo := GetSmtpConnInformation()
	utils.AssertEqual(t, "abc.com", connInfo.Server)
	utils.AssertEqual(t, 1234, connInfo.Port)
	utils.AssertEqual(t, "joe", connInfo.Username)
	utils.AssertEqual(t, "jeff", connInfo.Password)
}

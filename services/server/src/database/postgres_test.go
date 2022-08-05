package database

import (
	"os"
	"testing"
)

// Test that it panics when we try to call Newconnection with mock=true and not testing db
func TestNewConnectionPanicsWhenNameWrongAndMock(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	// The following is the code under test
	NewConnection(&Config{
		Host:     os.Getenv("DB_MOCK_HOST"),
		Port:     os.Getenv("DB_MOCK_PORT"),
		Password: os.Getenv("DB_MOCK_PASS"),
		User:     os.Getenv("DB_MOCK_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   "production",
	}, true)
}

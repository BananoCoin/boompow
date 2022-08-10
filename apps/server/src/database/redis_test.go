package database

import (
	"os"
	"testing"

	utils "github.com/bananocoin/boompow/libs/utils/testing"
	"github.com/google/uuid"
)

// Test that it panics when we try to call Newconnection with mock=true and not testing db
func TestMockRedis(t *testing.T) {
	os.Setenv("MOCK_REDIS", "true")

	redis := GetRedisDB()
	utils.AssertEqual(t, true, redis.Mock)
}

func TestRedis(t *testing.T) {
	os.Setenv("MOCK_REDIS", "true")

	redis := GetRedisDB()

	// Confirmation token bits
	if err := redis.SetConfirmationToken("email", "token"); err != nil {
		t.Errorf("Error setting confirmation token: %s", err)
	}
	token, err := redis.GetConfirmationToken("email")
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, "token", token)

	ret, err := redis.DeleteConfirmationToken("email")
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, int64(1), ret)

	// Connected clients bits
	if err := redis.AddConnectedClient("1"); err != nil {
		t.Errorf("Error adding client: %s", err)
	}
	ret, err = redis.GetNumberConnectedClients()
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, int64(1), ret)
	if err := redis.RemoveConnectedClient("1"); err != nil {
		t.Errorf("Error removing client: %s", err)
	}
	ret, err = redis.GetNumberConnectedClients()
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, int64(0), ret)
	// Add a couple clients
	if err := redis.AddConnectedClient("1"); err != nil {
		t.Errorf("Error adding client: %s", err)
	}
	if err := redis.AddConnectedClient("2"); err != nil {
		t.Errorf("Error adding client: %s", err)
	}
	ret, err = redis.GetNumberConnectedClients()
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, int64(2), ret)
	ret, err = redis.WipeAllConnectedClients()
	utils.AssertEqual(t, nil, err)
	ret, err = redis.GetNumberConnectedClients()
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, int64(0), ret)

	// Service token bits
	uid := uuid.New()
	if err := redis.AddServiceToken(uid, "token"); err != nil {
		t.Errorf("Error adding service token: %s", err)
	}
	uidStr, err := redis.GetServiceTokenUser("token")
	utils.AssertEqual(t, nil, err)
	utils.AssertEqual(t, uid.String(), uidStr)
	_, err = redis.GetServiceTokenUser("nonexistentoken")
	utils.AssertEqual(t, true, err != nil)
}

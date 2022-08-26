package database

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/bananocoin/boompow/apps/server/src/config"
	"github.com/bananocoin/boompow/libs/utils"
	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
	"k8s.io/klog/v2"
)

var ctx = context.Background()

// Prefix for all keys
const keyPrefix = "boompow"

// Singleton to keep assets loaded in memory
type redisManager struct {
	Client *redis.Client
	Mock   bool
}

var singleton *redisManager
var once sync.Once

func GetRedisDB() *redisManager {
	once.Do(func() {
		if utils.GetEnv("MOCK_REDIS", "false") == "true" {
			klog.Infof("Using mock redis client because MOCK_REDIS=true is set in environment")
			mr, _ := miniredis.Run()
			client := redis.NewClient(&redis.Options{
				Addr: mr.Addr(),
			})
			singleton = &redisManager{
				Client: client,
				Mock:   true,
			}
		} else {
			redis_port, err := strconv.Atoi(utils.GetEnv("REDIS_PORT", "6379"))
			if err != nil {
				panic("Invalid REDIS_PORT specified")
			}
			redis_db, err := strconv.Atoi(utils.GetEnv("REDIS_DB", "0"))
			if err != nil {
				panic("Invalid REDIS_DB specified")
			}
			client := redis.NewClient(&redis.Options{
				Addr: fmt.Sprintf("%s:%d", utils.GetEnv("REDIS_HOST", "localhost"), redis_port),
				DB:   redis_db,
			})
			singleton = &redisManager{
				Client: client,
				Mock:   false,
			}
		}
	})
	return singleton
}

// del - Redis DEL
func (r *redisManager) Del(key string) (int64, error) {
	val, err := r.Client.Del(ctx, key).Result()
	return val, err
}

// get - Redis GET
func (r *redisManager) Get(key string) (string, error) {
	val, err := r.Client.Get(ctx, key).Result()
	return val, err
}

// set - Redis SET
func (r *redisManager) Set(key string, value string, expiry time.Duration) error {
	err := r.Client.Set(ctx, key, value, expiry).Err()
	return err
}

// hlen - Redis HLEN
func (r *redisManager) Hlen(key string) (int64, error) {
	val, err := r.Client.HLen(ctx, key).Result()
	return val, err
}

// hget - Redis HGET
func (r *redisManager) Hget(key string, field string) (string, error) {
	val, err := r.Client.HGet(ctx, key, field).Result()
	return val, err
}

// hgetall - Redis HGETALL
func (r *redisManager) Hgetall(key string) (map[string]string, error) {
	val, err := r.Client.HGetAll(ctx, key).Result()
	return val, err
}

// hset - Redis HSET
func (r *redisManager) Hset(key string, field string, values interface{}) error {
	err := r.Client.HSet(ctx, key, field, values).Err()
	return err
}

// hdel - Redis HDEL
func (r *redisManager) Hdel(key string, field string) error {
	err := r.Client.HDel(ctx, key, field).Err()
	return err
}

// Set email confirmation token
func (r *redisManager) SetConfirmationToken(email string, token string) error {
	// Expire in 24H
	return r.Set(fmt.Sprintf("emailconfirmation:%s", email), token, config.EMAIL_CONFIRMATION_TOKEN_VALID_MINUTES*time.Minute)
}

// Get token for given email
func (r *redisManager) GetConfirmationToken(email string) (string, error) {
	return r.Get(fmt.Sprintf("emailconfirmation:%s", email))
}

// Delete conf token
func (r *redisManager) DeleteConfirmationToken(email string) (int64, error) {
	return r.Del(fmt.Sprintf("emailconfirmation:%s", email))
}

// Set reset password token
func (r *redisManager) SetResetPasswordToken(email string, token string) error {
	// Expire in 24H
	return r.Set(fmt.Sprintf("passwordreset:%s", email), token, config.EMAIL_CONFIRMATION_TOKEN_VALID_MINUTES*time.Minute)
}

// Get token for given email
func (r *redisManager) GetResetPasswordToken(email string) (string, error) {
	return r.Get(fmt.Sprintf("passwordreset:%s", email))
}

// Delete conf token
func (r *redisManager) DeleteResetPasswordToken(email string) (int64, error) {
	return r.Del(fmt.Sprintf("passwordreset:%s", email))
}

// Set token
func (r *redisManager) SetApproveServiceToken(email string, token string) error {
	// Expire in 2 weeks
	return r.Set(fmt.Sprintf("approveservice:%s", email), token, 336*time.Hour)
}

// Get token for given email
func (r *redisManager) GetApproveServiceToken(email string) (string, error) {
	return r.Get(fmt.Sprintf("approveservice:%s", email))
}

// Delete conf token
func (r *redisManager) DeleteApproveServiceToken(email string) (int64, error) {
	return r.Del(fmt.Sprintf("approveservice:%s", email))
}

// Functions for keeping track of connected clients
func (r *redisManager) AddConnectedClient(clientID string) error {
	return r.Hset("clients", clientID, "1")
}

func (r *redisManager) RemoveConnectedClient(clientID string) error {
	return r.Hdel("clients", clientID)
}

func (r *redisManager) GetNumberConnectedClients() (int64, error) {
	return r.Hlen("clients")
}

func (r *redisManager) WipeAllConnectedClients() (int64, error) {
	return r.Del("clients")
}

// For service tokens
func (r *redisManager) AddServiceToken(userID uuid.UUID, token string) error {
	userIdStr := userID.String()
	return r.Hset("servicetokens", token, userIdStr)
}

func (r *redisManager) GetServiceTokenUser(serviceToken string) (string, error) {
	user, err := r.Hget("servicetokens", serviceToken)
	if err != nil {
		return "", err
	}
	return user, nil
}

func (r *redisManager) GetServiceTokenForUser(userID uuid.UUID) (string, error) {
	userIdStr := userID.String()
	ret, err := r.Hgetall("servicetokens")
	if err != nil {
		return "", err
	}

	for k, v := range ret {
		if v == userIdStr {
			return k, nil
		}
	}
	return "", errors.New("No Token")
}

// For caching work
func (r *redisManager) CacheWork(hash string, result string) error {
	// 5 minute cache
	return r.Set(fmt.Sprintf("cache:%s", hash), result, 5*time.Minute)
}

func (r *redisManager) GetCachedWork(hash string) (string, error) {
	return r.Get(fmt.Sprintf("cache:%s", hash))
}

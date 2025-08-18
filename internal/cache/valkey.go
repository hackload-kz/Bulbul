package cache

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type ValkeyClient struct {
	client      *redis.Client
	usersHashKey string
}

func NewValkeyClient() (*ValkeyClient, error) {
	addr := os.Getenv("VALKEY_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	password := os.Getenv("VALKEY_PASSWORD")
	usersHashKey := os.Getenv("VALKEY_USERS_HASH_KEY")
	if usersHashKey == "" {
		usersHashKey = "users:auth"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		DialTimeout:  5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Valkey: %w", err)
	}

	return &ValkeyClient{
		client:       rdb,
		usersHashKey: usersHashKey,
	}, nil
}

func (v *ValkeyClient) GetUserIDByAuth(ctx context.Context, email, passwordHash string) (int64, error) {
	authString := fmt.Sprintf("%s:%s", email, passwordHash)
	cacheKey := base64.StdEncoding.EncodeToString([]byte(authString))

	userIDStr, err := v.client.HGet(ctx, v.usersHashKey, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("user not found in cache")
		}
		return 0, fmt.Errorf("cache lookup error: %w", err)
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID in cache: %w", err)
	}

	return userID, nil
}

func (v *ValkeyClient) Close() error {
	return v.client.Close()
}
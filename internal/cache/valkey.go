package cache

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/redis/rueidis"
)

type ValkeyClient struct {
	client         rueidis.Client
	usersHashKey   string
	authCacheTTL   time.Duration
	eventsCacheTTL time.Duration
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

	// Parse cache TTL configurations
	authCacheTTL := getEnvDuration("VALKEY_AUTH_CACHE_TTL_MIN", 10*time.Minute)
	eventsCacheTTL := getEnvDuration("VALKEY_EVENTS_CACHE_TTL_MIN", 15*time.Minute)

	// Parse cache size in MB
	cacheSizeMB := getEnvInt("VALKEY_CLIENT_CACHE_SIZE_MB", 128)
	cacheSizeBytes := int64(cacheSizeMB) * (1 << 20) // Convert MB to bytes

	// Create rueidis client with comprehensive connection pool settings
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{addr},
		Password:    password,
		SelectDB:    0,

		// Client-side caching configuration
		CacheSizeEachConn: int(cacheSizeBytes),

		// Connection pool configuration to prevent blocking/exhaustion
		ConnWriteTimeout:      5 * time.Second, // Detect hanging connections quickly
		BlockingPoolSize:      64,              // Increase pool size for concurrent requests
		PipelineMultiplex:     8,               // Limit concurrent pipelined commands
		DisableAutoPipelining: true,            // Avoid Head-of-Line blocking under load
		AlwaysRESP2:           true,            // Use RESP2 for better compatibility

		// Note: Connection timeouts are handled by ConnWriteTimeout above
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create rueidis client: %w", err)
	}

	// Test connection with ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Do(ctx, client.B().Ping().Build()).Error(); err != nil {
		return nil, fmt.Errorf("failed to connect to Valkey: %w", err)
	}

	return &ValkeyClient{
		client:         client,
		usersHashKey:   usersHashKey,
		authCacheTTL:   authCacheTTL,
		eventsCacheTTL: eventsCacheTTL,
	}, nil
}

// Helper functions for parsing environment variables
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return time.Duration(intValue) * time.Minute
		}
	}
	return defaultValue
}

func (v *ValkeyClient) GetUserIDByAuth(ctx context.Context, email, passwordHash string) (int64, error) {
	authString := fmt.Sprintf("%s:%s", email, passwordHash)
	cacheKey := base64.StdEncoding.EncodeToString([]byte(authString))

	// Use client-side caching for auth lookups (they rarely change)
	result := v.client.DoCache(ctx,
		v.client.B().Hget().Key(v.usersHashKey).Field(cacheKey).Cache(),
		v.authCacheTTL,
	)

	userIDStr, err := result.ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
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
	v.client.Close()
	return nil
}

// LogConnectionPoolStats logs current connection pool statistics for monitoring
func (v *ValkeyClient) LogConnectionPoolStats(ctx context.Context) {
	// Log pool configuration and usage for debugging
	slog.Info("Valkey connection pool status - Client configured with auto-pipelining disabled")
}

func (v *ValkeyClient) generateEventsListCacheKey(page, pageSize int) string {
	return fmt.Sprintf("events:list:page:%d:size:%d", page, pageSize)
}

func (v *ValkeyClient) SetEventsList(ctx context.Context, page, pageSize int, events interface{}) error {
	cacheKey := v.generateEventsListCacheKey(page, pageSize)

	eventData, err := json.Marshal(events)
	if err != nil {
		return fmt.Errorf("failed to marshal events data: %w", err)
	}

	// Use background context with timeout for cache SET operations (non-blocking)
	// This prevents cache writes from affecting response times
	cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use regular Do() for SET operations (not cacheable)
	serverTTL := v.eventsCacheTTL
	err = v.client.Do(cacheCtx,
		v.client.B().Set().Key(cacheKey).Value(string(eventData)).Ex(serverTTL).Build(),
	).Error()

	if err != nil {
		slog.Error("failed to set events cache", "error", err)
	}

	return nil
}

func (v *ValkeyClient) GetEventsList(ctx context.Context, page, pageSize int, result interface{}) error {
	cacheKey := v.generateEventsListCacheKey(page, pageSize)

	// Create a timeout context for cache operations (2 seconds max)
	cacheCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Use client-side caching for GET operations
	resp := v.client.DoCache(cacheCtx,
		v.client.B().Get().Key(cacheKey).Cache(),
		v.eventsCacheTTL,
	)

	cachedData, err := resp.ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return fmt.Errorf("cache miss")
		}
		// Check for context timeout/cancellation
		if ctx.Err() != nil {
			return fmt.Errorf("cache operation cancelled: %w", ctx.Err())
		}
		return fmt.Errorf("cache lookup error: %w", err)
	}

	err = json.Unmarshal([]byte(cachedData), result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal cached events data: %w", err)
	}

	return nil
}

// GetEventsListRaw returns raw JSON bytes from cache without unmarshaling
// This avoids the overhead of JSON unmarshaling when serving cached responses directly
// Uses client-side caching for maximum performance with fallback strategy
func (v *ValkeyClient) GetEventsListRaw(ctx context.Context, page, pageSize int) ([]byte, error) {
	cacheKey := v.generateEventsListCacheKey(page, pageSize)

	// Create a timeout context for cache operations (1 second max for reads)
	cacheCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	// Use client-side caching with raw bytes for maximum performance
	resp := v.client.DoCache(cacheCtx,
		v.client.B().Get().Key(cacheKey).Cache(),
		v.eventsCacheTTL,
	)

	// Check if served from client-side cache
	if resp.IsCacheHit() {
		// Ultra-fast path: served directly from client-side memory
		slog.Info("Cache HIT (client-side) for events list", "page", page, "page_size", pageSize)
	}

	cachedData, err := resp.ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return nil, fmt.Errorf("cache miss")
		}
		// Check for context timeout/cancellation - this is the likely culprit!
		if cacheCtx.Err() != nil {
			slog.Warn("Cache operation TIMEOUT for events list", "page", page, "page_size", pageSize, "error", cacheCtx.Err())
			return nil, fmt.Errorf("cache operation timeout: %w", cacheCtx.Err())
		}
		slog.Error("Cache lookup ERROR for events list", "page", page, "page_size", pageSize, "error", err)
		return nil, fmt.Errorf("cache lookup error: %w", err)
	}

	return []byte(cachedData), nil
}

package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-dev-frame/sponge/pkg/cache"
	"github.com/go-dev-frame/sponge/pkg/encoding"

	"user-server-go/internal/database"
)

const (
	// cache prefix key, must end with a colon
	userTokenCachePrefixKey = "user:token:"
	// UserTokenExpireTime expire time
	UserTokenExpireTime = 24 * time.Hour
)

var _ UserTokenCache = (*userTokenCache)(nil)

// UserTokenCache cache interface
type UserTokenCache interface {
	Set(ctx context.Context, id uint64, token string, duration time.Duration) error
	Get(ctx context.Context, id uint64) (string, error)
	Del(ctx context.Context, id uint64) error
}

type userTokenCache struct {
	cache cache.Cache
}

// NewUserTokenCache create a new cache
func NewUserTokenCache(cacheType *database.CacheType) UserTokenCache {
	newObject := func() interface{} {
		return ""
	}
	cachePrefix := ""
	jsonEncoding := encoding.JSONEncoding{}

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, newObject)
		return &userTokenCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, newObject)
		return &userTokenCache{cache: c}
	}

	panic(fmt.Sprintf("unsupported cache type='%s'", cacheType.CType))
}

// cache key
func (c *userTokenCache) getUserTokenCacheKey(id uint64) string {
	return fmt.Sprintf("%s%v", userTokenCachePrefixKey, id)
}

// Set cache
func (c *userTokenCache) Set(ctx context.Context, id uint64, token string, duration time.Duration) error {
	cacheKey := c.getUserTokenCacheKey(id)
	return c.cache.Set(ctx, cacheKey, &token, duration)
}

// Get cache
func (c *userTokenCache) Get(ctx context.Context, id uint64) (string, error) {
	var token string
	cacheKey := c.getUserTokenCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &token)
	if err != nil {
		return token, err
	}
	return token, nil
}

// Del delete cache
func (c *userTokenCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.getUserTokenCacheKey(id)
	return c.cache.Del(ctx, cacheKey)
}

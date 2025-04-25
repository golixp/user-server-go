package cache

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-dev-frame/sponge/pkg/cache"
	"github.com/go-dev-frame/sponge/pkg/encoding"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"user-server-go/internal/database"
	"user-server-go/internal/model"
)

const (
	// cache prefix key, must end with a colon
	userCachePrefixKey = "user:"
	// UserExpireTime expire time
	UserExpireTime = 5 * time.Minute
)

var _ UserCache = (*userCache)(nil)

// UserCache cache interface
type UserCache interface {
	Set(ctx context.Context, id uint64, data *model.User, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.User, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.User, error)
	MultiSet(ctx context.Context, data []*model.User, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool

	SetByUsername(ctx context.Context, email string, data *model.User, duration time.Duration) error
	GetByUsername(ctx context.Context, email string) (*model.User, error)
	DelByUsername(ctx context.Context, email string) error
	SetUsernamePlaceholder(ctx context.Context, email string) error
}

// userCache define a cache struct
type userCache struct {
	cache cache.Cache
}

// NewUserCache new a cache
func NewUserCache(cacheType *database.CacheType) UserCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.User{}
		})
		return &userCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.User{}
		})
		return &userCache{cache: c}
	}

	return nil // no cache
}

// GetUserCacheKey cache key
func (c *userCache) GetUserCacheKey(id uint64) string {
	return userCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *userCache) Set(ctx context.Context, id uint64, data *model.User, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetUserCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *userCache) Get(ctx context.Context, id uint64) (*model.User, error) {
	var data *model.User
	cacheKey := c.GetUserCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *userCache) MultiSet(ctx context.Context, data []*model.User, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetUserCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *userCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.User, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetUserCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.User)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.User)
	for _, id := range ids {
		val, ok := itemMap[c.GetUserCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *userCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetUserCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *userCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetUserCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *userCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}

// Username cache key
func (c *userCache) getUserUsernameCacheKey(username string) string {
	return userCachePrefixKey + username
}

// SetByUsername write to cache
func (c *userCache) SetByUsername(ctx context.Context, username string, data *model.User, duration time.Duration) error {
	if data == nil || username == "" {
		return nil
	}
	cacheKey := c.getUserUsernameCacheKey(username)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// GetByUsername cache value
func (c *userCache) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var data *model.User
	cacheKey := c.getUserUsernameCacheKey(username)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// DelByUsername delete cache
func (c *userCache) DelByUsername(ctx context.Context, username string) error {
	cacheKey := c.getUserUsernameCacheKey(username)
	return c.cache.Del(ctx, cacheKey)
}

// SetUsernamePlaceholder set empty cache
func (c *userCache) SetUsernamePlaceholder(ctx context.Context, username string) error {
	cacheKey := c.getUserUsernameCacheKey(username)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

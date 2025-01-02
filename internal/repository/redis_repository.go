package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	goredis "github.com/go-redis/redis/v8"
	"strings"
	"time"
	"weather-api/internal/models"
)

type RedisRepository struct {
	redis *goredis.Client
}

func NewRedisCacheRepository(rd *goredis.Client) *RedisRepository {
	return &RedisRepository{
		redis: rd,
	}
}

func (r *RedisRepository) SetCachedRepository(ctx context.Context, response *models.WeatherModel) error {
	response.Cached = true
	key := strings.ToLower(response.City)

	str, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return r.redis.Set(ctx, key, str, 1*time.Minute).Err()
}

func (r *RedisRepository) GetCachedRepository(ctx context.Context, key string, response *models.WeatherModel) error {
	val, err := r.redis.Get(ctx, strings.ToLower(key)).Result()
	if err != nil {
		return errors.New(fmt.Sprintf("redis.Get key=%s err = %s", key, err.Error()))
	}

	return json.Unmarshal([]byte(val), &response)
}

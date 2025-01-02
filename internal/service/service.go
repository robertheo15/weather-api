package service

import "weather-api/internal/repository"

type Service struct {
	redisRepository *repository.RedisRepository
	apiUsage        int
	cacheHits       int
}

func NewService(redisRepository *repository.RedisRepository) *Service {
	return &Service{
		redisRepository: redisRepository,
	}
}

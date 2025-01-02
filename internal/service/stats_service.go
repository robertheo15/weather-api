package service

import (
	"weather-api/internal/models"
)

func (s *Service) IncrementAPIUsage() {
	s.apiUsage++
}

func (s *Service) IncrementCacheHits() {
	s.cacheHits++
}

func (s *Service) GetStats() *models.Stats {
	stats := &models.Stats{
		APIUsage:  s.apiUsage,
		CacheHits: s.cacheHits,
	}

	return stats
}

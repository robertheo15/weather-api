package models

type Stats struct {
	APIUsage  int `json:"api_usage"`
	CacheHits int `json:"cache_hits"`
}

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"weather-api/internal/models"
)

func (s *Service) GetOpenWeatherAPI(ctx context.Context, city string) (*models.WeatherModel, error) {
	apiURL := os.Getenv("OPENWEATHER_API")
	apiKey := os.Getenv("OPENWEATHERMAP_KEY")
	if apiURL == "" || apiKey == "" {
		return nil, errors.New("API URL or key not set in environment variables")
	}

	url := fmt.Sprintf("%s/weather?q=%s&appid=%s", apiURL, city, apiKey)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("API returned non-200 status code: %d, body: %s", response.StatusCode, string(body))
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var weatherData models.OpenWeatherDataResponse
	if err := json.Unmarshal(responseData, &weatherData); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	return models.OpenWeatherDataResponseToWeather(weatherData), nil
}

func (s *Service) GetWeatherAPI(ctx context.Context, city string) (*models.WeatherModel, error) {
	apiURL := os.Getenv("WEATHER_API")
	apiKEY := os.Getenv("WEATHERAPI_KEY")
	if apiURL == "" || apiKEY == "" {
		return nil, errors.New("API URL or key not set in environment variables")
	}

	url := fmt.Sprintf("%s/current.json?q=%s&key=%s", apiURL, city, apiKEY)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("API returned non-200 status code: %d, body: %s", response.StatusCode, string(body))
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var weatherData models.WeatherAPIResponse
	if err := json.Unmarshal(responseData, &weatherData); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	return models.WeatherAPIToWeather(weatherData), nil
}

func (s *Service) FetchWeatherFromAPIs(ctx context.Context, city string) (*models.WeatherModel, error) {
	cachedResult, err := s.getWeatherFromCache(ctx, city)
	if err == nil {
		s.IncrementCacheHits()
		return cachedResult, nil
	}

	s.IncrementAPIUsage()

	result, err := s.fetchWeatherConcurrently(ctx, city)
	if err != nil {
		return nil, err
	}

	if err := s.cacheWeatherResult(ctx, result); err != nil {
		log.Printf("Failed to cache weather result: %v", err)
	}

	return result, nil
}

func (s *Service) getWeatherFromCache(ctx context.Context, city string) (*models.WeatherModel, error) {
	var result models.WeatherModel
	err := s.redisRepository.GetCachedRepository(ctx, city, &result)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve from cache: %w", err)
	}
	return &result, nil
}

func (s *Service) fetchWeatherConcurrently(ctx context.Context, city string) (*models.WeatherModel, error) {
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	apis := []func(context.Context, string) (*models.WeatherModel, error){
		s.GetOpenWeatherAPI,
		s.GetWeatherAPI,
	}

	resultCh := make(chan *models.WeatherModel, len(apis))
	errorCh := make(chan error, len(apis))

	for _, apiFunc := range apis {
		wg.Add(1)
		go func(apiFunc func(context.Context, string) (*models.WeatherModel, error)) {
			defer wg.Done()
			if data, err := apiFunc(ctx, city); err != nil {
				log.Printf("API fetch error: %v", err)
				errorCh <- err
			} else {
				resultCh <- data
			}
		}(apiFunc)
	}

	go func() {
		wg.Wait()
		close(resultCh)
		close(errorCh)
	}()

	return s.collectWeatherResults(ctx, resultCh, errorCh)
}

func (s *Service) collectWeatherResults(ctx context.Context, resultCh <-chan *models.WeatherModel, errorCh <-chan error) (*models.WeatherModel, error) {
	var result *models.WeatherModel
	var err error

	for {
		select {
		case res, ok := <-resultCh:
			if ok {
				return res, nil
			}
		case apiErr, ok := <-errorCh:
			if ok {
				err = apiErr
			}
		case <-ctx.Done():
			return nil, errors.New("timeout exceeded while fetching weather data")
		}

		if len(resultCh) == 0 && len(errorCh) == 0 {
			break
		}
	}

	if result == nil && err != nil {
		return nil, fmt.Errorf("all API calls failed: %w", err)
	}
	return result, nil
}

func (s *Service) cacheWeatherResult(ctx context.Context, result *models.WeatherModel) error {
	err := s.redisRepository.SetCachedRepository(ctx, result)
	if err != nil {
		return fmt.Errorf("failed to set cache weather data: %w", err)
	}
	return nil
}

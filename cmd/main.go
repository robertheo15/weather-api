package main

import (
	"context"
	"github.com/go-chi/chi"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"weather-api/internal/config"
	"weather-api/internal/http"
	"weather-api/internal/repository"
	"weather-api/internal/service"
)

func main() {
	locJakarta, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Fatalf("config: failed to load Asia/Jakarta location error=%s", err)
	}
	ctx := context.Background()
	config.LoadEnvFile()

	clientRedis := config.NewRedis(ctx)
	repo := repository.NewRedisCacheRepository(clientRedis)
	svc := service.NewService(repo)

	time.Local = locJakarta

	router := chi.NewRouter()
	log.Println("Server running on port : " + config.Port)
	server := http.NewServer(router, svc, ctx)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Run(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-sigs

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}

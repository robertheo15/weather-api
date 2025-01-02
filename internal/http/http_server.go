package http

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"log"
	"net/http"
	"time"
	"weather-api/internal/config"
	"weather-api/internal/service"
)

type Server struct {
	router     *chi.Mux
	service    *service.Service
	ctx        context.Context
	httpServer *http.Server
}

func NewServer(router *chi.Mux, service *service.Service, ctx context.Context) *Server {
	s := &Server{
		router:  router,
		service: service,
		ctx:     ctx,
	}

	s.httpServer = &http.Server{
		Addr:           ":" + config.Port,
		Handler:        s.router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return s
}

func (s *Server) Run() error {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "HEAD", "OPTION"},
		AllowedHeaders:   []string{"User-Agent", "Content-Type", "Accept", "Accept-Encoding", "Accept-Language", "Cache-Control", "Connection", "DNT", "Host", "Origin", "Pragma", "Referer"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	s.router.Use(httprate.LimitByIP(60, time.Duration(1)*time.Minute))

	s.router.Mount("/api/v1", s.router)

	// weather
	s.router.Get("/weathers", s.GetWeatherByCity)

	// stats
	s.router.Get("/stats", s.GetStats)

	log.Printf("Starting server on port %s", config.Port)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}

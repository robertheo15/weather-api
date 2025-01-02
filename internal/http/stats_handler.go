package http

import (
	"errors"
	"net/http"
	pkgHttp "weather-api/pkg/http"
)

func (s *Server) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := s.service.GetStats()
	if stats == nil {
		pkgHttp.SetError(w, http.StatusInternalServerError, errors.New("failed to fetch stats"))
		return
	}

	pkgHttp.SetResponse(w, http.StatusOK, stats, "Stats fetched successfully", true)

}

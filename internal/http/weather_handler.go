package http

import (
	"net/http"
	pkgHttp "weather-api/pkg/http"
)

func (s *Server) GetWeatherByCity(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	weather, err := s.service.FetchWeatherFromAPIs(s.ctx, city)
	if err != nil {
		pkgHttp.SetError(w, http.StatusInternalServerError, err)
		return
	}

	data := map[string]interface{}{
		"weather": weather,
	}
	pkgHttp.SetResponse(w, http.StatusOK, data, "Weather fetched successfully", true)
}

package server

import (
	"configuration-management/internal/models"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (s *Server) HeaderBelongsToConfig(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		config, ok := c.Get("config").(*models.Config)
		if !ok {
			log.Println("Missing config")
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		headerID, idErr := uuid.Parse(c.Param("headerId"))
		if idErr != nil {
			log.Fatalf("Invalid header id: %e", idErr)
			return echo.NewHTTPError(http.StatusBadRequest, "invalid header id")
		}

		header, err := s.db.GetHeaderReplacement(headerID)
		if err != nil {
			log.Printf("failed to get header replacement: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError)

		}

		if header.ConfigID != config.ID {
			log.Println("header does not belong to the config")
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		c.Set("header", header)
		return next(c)
	}
}

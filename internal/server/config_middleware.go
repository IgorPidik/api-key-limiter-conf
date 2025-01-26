package server

import (
	"configuration-management/internal/models"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (s *Server) ConfigBelongToProject(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		project, ok := c.Get("project").(*models.Project)
		if !ok {
			log.Println("Missing project")
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		configID, idErr := uuid.Parse(c.Param("configId"))
		if idErr != nil {
			log.Fatalf("Invalid config id: %e", idErr)
			return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
		}

		config, err := s.db.GetConfig(configID)
		if err != nil {
			log.Printf("failed to get config: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError)

		}

		if config.ProjectID != project.ID {
			log.Println("config does not belong to the project")
			return echo.NewHTTPError(http.StatusBadRequest)
		}

		c.Set("config", config)
		return next(c)
	}
}

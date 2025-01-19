package server

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (s *Server) ProjectBelongsToLoggedUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, ok := c.Get("userID").(string)
		if !ok {
			log.Println("Missing user")
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		projectID, idErr := uuid.Parse(c.Param("id"))
		if idErr != nil {
			log.Fatalf("Invalid project id: %e", idErr)
			return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
		}
		project, err := s.db.GetProject(projectID)
		if err != nil {
			log.Printf("failed to get project: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError)

		}

		if project.UserID != uuid.MustParse(userID) {
			log.Println("project does not belong to the logged user")
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		return next(c)
	}
}

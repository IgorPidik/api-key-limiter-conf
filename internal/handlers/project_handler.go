package handlers

import (
	"configuration-management/internal/database"
	"configuration-management/web/projects_components"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ProjectHandler struct {
	db *database.DatabaseService
}

func NewProjectHandler(db *database.DatabaseService) *ProjectHandler {
	return &ProjectHandler{db}
}

func (p *ProjectHandler) ListProjects(c echo.Context) error {
	projects, err := p.db.ListProjects()
	if err != nil {
		log.Fatalf("Error fetching projects: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	component := projects_components.ListProjects(projects)
	renderErr := component.Render(c.Request().Context(), c.Response().Writer)
	if renderErr != nil {
		log.Fatalf("Error rendering in HelloWebHandler: %e", renderErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

package handlers

import (
	"configuration-management/internal/database"
	"configuration-management/internal/utils"
	"configuration-management/web/projects_components"
	"log"
	"net/http"

	"github.com/google/uuid"
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

	component := projects_components.Projects(projects)
	renderErr := component.Render(c.Request().Context(), c.Response().Writer)
	if renderErr != nil {
		log.Fatalf("Error rendering in HelloWebHandler: %e", renderErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

func (p *ProjectHandler) CreateProject(c echo.Context) error {
	if c.Request().ParseForm() != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	name := c.Request().FormValue("project-name")
	accessKey := utils.GenerateToken(126)

	userId := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	project, projectErr := p.db.CreateProject(name, accessKey, userId)
	if projectErr != nil {
		log.Fatalf("Error creating project: %e", projectErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	component := projects_components.ProjectDetails(*project, true)
	if err := component.Render(c.Request().Context(), c.Response().Writer); err != nil {
		log.Fatalf("Error rendering created project: %e", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

func (p *ProjectHandler) DeleteProject(c echo.Context) error {
	projectId, idErr := uuid.Parse(c.Param("id"))
	if idErr != nil {
		log.Fatalf("Invalid project id: %e", idErr)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}

	if deleteErr := p.db.DeleteProject(projectId); deleteErr != nil {
		log.Fatalf("Failed to delete project: %e", deleteErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

func (p *ProjectHandler) CreateConfig(c echo.Context) error {
	projectID, idErr := uuid.Parse(c.Param("id"))
	if idErr != nil {
		log.Fatalf("Invalid project id: %e", idErr)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}

	if err := c.Request().ParseForm(); err != nil {
		log.Fatalf("failed to parse form for creating config: %e", idErr)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	name := c.Request().FormValue("name")
	header := c.Request().FormValue("header-name")
	value := c.Request().FormValue("header-value")
	host := c.Request().FormValue("host")
	config, configErr := p.db.CreateConfig(projectID, name, host, header, value)
	if configErr != nil {
		log.Fatalf("Failed to create config: %e", configErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	component := projects_components.ConfigDetails(projectID, *config)
	if err := component.Render(c.Request().Context(), c.Response().Writer); err != nil {
		log.Fatalf("Error rendering created config: %e", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

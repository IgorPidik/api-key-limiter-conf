package handlers

import (
	"configuration-management/internal/database"
	"configuration-management/internal/forms"
	"configuration-management/internal/utils"
	"configuration-management/web/projects_components"
	"log"
	"net/http"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type CreateProjectForm struct {
	Name string `form:"project-name" validate:"required"`
}

type CreateConfigForm struct {
	Name             string `form:"name" validate:"required"`
	HeaderName       string `form:"header-name" validate:"required"`
	HeaderValue      string `form:"header-value" validate:"required"`
	NumberOfRequests int    `form:"num-of-requests" validate:"required"`
	Per              string `form:"requests-per" validate:"required,oneof=second minute hour day"`
}

type ProjectHandler struct {
	db       *database.DatabaseHandler
	decoder  *form.Decoder
	validate *validator.Validate
}

func NewProjectHandler(db *database.DatabaseHandler) *ProjectHandler {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return &ProjectHandler{db, form.NewDecoder(), validate}
}

func (p *ProjectHandler) ListProjects(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		log.Println("Missing user")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	projects, err := p.db.ListProjects(uuid.MustParse(userID))
	if err != nil {
		log.Fatalf("Error fetching projects: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	component := projects_components.Projects(projects)
	renderErr := component.Render(c.Request().Context(), c.Response().Writer)
	if renderErr != nil {
		log.Fatalf("Error rendering in ListProjects: %e", renderErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

func (p *ProjectHandler) CreateProject(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		log.Println("Missing user")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if c.Request().ParseForm() != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	var createProjectForm CreateProjectForm
	if err := p.decoder.Decode(&createProjectForm, c.Request().Form); err != nil {
		log.Fatalf("Error decoding CreateProjectForm: %e", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if validationErr := p.validate.Struct(createProjectForm); validationErr != nil {
		errors := make(forms.FormErrors)
		for _, err := range validationErr.(validator.ValidationErrors) {
			errors[err.Field()] = err.Tag()
		}

		c.Response().Header().Set("HX-Reswap", "outerHTML")
		c.Response().Header().Set("HX-Retarget", "#create-project-form")
		component := projects_components.CreateProject(errors)
		if err := component.Render(c.Request().Context(), c.Response().Writer); err != nil {
			log.Fatalf("Error rendering created project: %e", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}

	accessKey := utils.GenerateToken(126)

	project, projectErr := p.db.CreateProject(createProjectForm.Name, accessKey, uuid.MustParse(userID))
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

	if c.Request().ParseForm() != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	var createConfigForm CreateConfigForm
	if err := p.decoder.Decode(&createConfigForm, c.Request().Form); err != nil {
		log.Fatalf("Error decoding CreateConfigForm: %e", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if validationErr := p.validate.Struct(createConfigForm); validationErr != nil {
		errors := make(forms.FormErrors)
		for _, err := range validationErr.(validator.ValidationErrors) {
			errors[err.Field()] = err.Tag()
		}

		c.Response().Header().Set("HX-Reswap", "outerHTML")
		c.Response().Header().Set("HX-Retarget", "#"+projects_components.GetCreateConfigFormID(projectID))
		component := projects_components.CreateConfigForm(projectID, errors)
		if err := component.Render(c.Request().Context(), c.Response().Writer); err != nil {
			log.Fatalf("Error rendering created config: %e", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}

	config, configErr := p.db.CreateConfig(projectID, createConfigForm.Name, createConfigForm.NumberOfRequests, createConfigForm.Per)
	if configErr != nil {
		log.Fatalf("Failed to create config: %e", configErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	headeReplacement, headerErr := p.db.CreateHeaderReplacement(config.ID, createConfigForm.HeaderName, createConfigForm.HeaderValue)
	if headerErr != nil {
		log.Fatalf("Failed to create header replacement: %e", headerErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	config.HeaderReplacements = append(config.HeaderReplacements, *headeReplacement)

	component := projects_components.ConfigDetails(*config)
	if err := component.Render(c.Request().Context(), c.Response().Writer); err != nil {
		log.Fatalf("Error rendering created config: %e", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

func (p *ProjectHandler) DeleteConfig(c echo.Context) error {
	projectID, projectIDErr := uuid.Parse(c.Param("id"))
	if projectIDErr != nil {
		log.Fatalf("Invalid project id: %e", projectIDErr)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	}

	configID, configIDErr := uuid.Parse(c.Param("configId"))
	if configIDErr != nil {
		log.Fatalf("Invalid config id: %e", configIDErr)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid config id")
	}

	if deleteErr := p.db.DeleteConfig(projectID, configID); deleteErr != nil {
		log.Fatalf("Failed to delete config: %e", deleteErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

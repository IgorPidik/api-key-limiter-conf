package handlers

import (
	"configuration-management/internal/database"
	"configuration-management/internal/forms"
	"configuration-management/internal/models"
	"configuration-management/internal/utils"
	"configuration-management/web/projects_components"
	"log"
	"net/http"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CreateProjectForm struct {
	Name        string `form:"name" validate:"required"`
	Description string `form:"name"`
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
	user, ok := c.Get("user").(*models.User)
	if !ok {
		log.Println("Missing user")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	projects, err := p.db.ListProjects(user.ID)
	if err != nil {
		log.Fatalf("Error fetching projects: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	component := projects_components.Projects(user, projects)
	renderErr := component.Render(c.Request().Context(), c.Response().Writer)
	if renderErr != nil {
		log.Fatalf("Error rendering in ListProjects: %e", renderErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

func (p *ProjectHandler) processCreateForm(c echo.Context) (*CreateProjectForm, forms.FormErrors, error) {
	if c.Request().ParseForm() != nil {
		return nil, nil, echo.NewHTTPError(http.StatusBadRequest)
	}
	var createProjectForm CreateProjectForm
	if err := p.decoder.Decode(&createProjectForm, c.Request().Form); err != nil {
		log.Fatalf("Error decoding CreateProjectForm: %e", err)
		return nil, nil, echo.NewHTTPError(http.StatusInternalServerError)
	}

	if validationErr := p.validate.Struct(createProjectForm); validationErr != nil {
		errors := make(forms.FormErrors)
		for _, err := range validationErr.(validator.ValidationErrors) {
			errors[err.Field()] = err.Tag()
		}

		return nil, errors, nil
	}
	return &createProjectForm, nil, nil
}

func (p *ProjectHandler) CreateProject(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		log.Println("Missing user")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	createProjectForm, formErrors, processingErr := p.processCreateForm(c)
	if processingErr != nil {
		return processingErr
	}
	if formErrors != nil {
		c.Response().Header().Set("HX-Reswap", "outerHTML")
		c.Response().Header().Set("HX-Retarget", "#create-project-form")
		component := projects_components.CreateProject(formErrors)
		if err := component.Render(c.Request().Context(), c.Response().Writer); err != nil {
			log.Fatalf("Error rendering created project: %e", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil

	}

	accessKey, err := utils.GenerateEncryptedToken(32)
	if err != nil {
		log.Fatalf("Error creating encrypted access key: %e", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	project, projectErr := p.db.CreateProject(createProjectForm.Name, createProjectForm.Description, accessKey, user.ID)
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
	project, ok := c.Get("project").(*models.Project)
	if !ok {
		log.Println("Missing project instance in the context")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if deleteErr := p.db.DeleteProject(project.ID); deleteErr != nil {
		log.Fatalf("Failed to delete project: %e", deleteErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

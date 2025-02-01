package handlers

import (
	"configuration-management/internal/database"
	"configuration-management/internal/forms"
	"configuration-management/internal/models"
	"configuration-management/internal/utils"
	"configuration-management/web/projects_components"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CreateConfigForm struct {
	Name             string `form:"name" validate:"required"`
	HeaderName       string `form:"header-name" validate:"required"`
	HeaderValue      string `form:"header-value" validate:"required"`
	NumberOfRequests int    `form:"num-of-requests" validate:"required"`
	Per              string `form:"requests-per" validate:"required,oneof=second minute hour day week month year forever"`
}

type ConfigHandler struct {
	db       *database.DatabaseHandler
	decoder  *form.Decoder
	validate *validator.Validate
}

func NewConfigHandler(db *database.DatabaseHandler) *ConfigHandler {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return &ConfigHandler{db, form.NewDecoder(), validate}
}

func (ch *ConfigHandler) processCreateConfigForm(c echo.Context) (*CreateConfigForm, forms.FormErrors, error) {
	if c.Request().ParseForm() != nil {
		return nil, nil, echo.NewHTTPError(http.StatusBadRequest)
	}
	var createConfigForm CreateConfigForm
	if err := ch.decoder.Decode(&createConfigForm, c.Request().Form); err != nil {
		log.Fatalf("Error decoding CreateConfigForm: %e", err)
		return nil, nil, echo.NewHTTPError(http.StatusInternalServerError)
	}

	if validationErr := ch.validate.Struct(createConfigForm); validationErr != nil {
		errors := make(forms.FormErrors)
		for _, err := range validationErr.(validator.ValidationErrors) {
			errors[err.Field()] = err.Tag()
		}
		return nil, errors, nil
	}
	return &createConfigForm, nil, nil
}

func (ch *ConfigHandler) CreateConfig(c echo.Context) error {
	project, ok := c.Get("project").(*models.Project)
	if !ok {
		log.Println("Missing project instance in the context")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	createConfigForm, formErrs, processingErr := ch.processCreateConfigForm(c)
	if processingErr != nil {
		return processingErr
	}

	if formErrs != nil {
		c.Response().Header().Set("HX-Reswap", "outerHTML")
		c.Response().Header().Set("HX-Retarget", "#"+projects_components.GetCreateConfigFormID(project.ID))
		component := projects_components.CreateConfigForm(project.ID, formErrs)
		if err := component.Render(c.Request().Context(), c.Response().Writer); err != nil {
			log.Fatalf("Error rendering created config: %e", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}

	config, configErr := ch.db.CreateConfig(project.ID, createConfigForm.Name, createConfigForm.NumberOfRequests, createConfigForm.Per)
	if configErr != nil {
		log.Fatalf("Failed to create config: %e", configErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	encryptedValue, encryptErr := utils.EncryptData(createConfigForm.HeaderValue)
	if encryptErr != nil {
		log.Fatalf("Failed to encrypt header value: %e", encryptErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	headeReplacement, headerErr := ch.db.CreateHeaderReplacement(config.ID, createConfigForm.HeaderName, encryptedValue)
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

func (ch *ConfigHandler) DeleteConfig(c echo.Context) error {
	config, ok := c.Get("config").(*models.Config)
	if !ok {
		log.Println("Missing config instance in the context")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if deleteErr := ch.db.DeleteConfig(config.ID); deleteErr != nil {
		log.Fatalf("Failed to delete config: %e", deleteErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

func (ch *ConfigHandler) GetConfigConnection(c echo.Context) error {
	config, ok := c.Get("config").(*models.Config)
	if !ok {
		log.Println("Missing config instance in the context")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	project, ok := c.Get("project").(*models.Project)
	if !ok {
		log.Println("Missing project")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	host := os.Getenv("PROXY_HOST")
	connectionString := fmt.Sprintf("https://%s:%s:%s@%s", config.ID, project.ID, project.AccessKey, host)
	component := projects_components.ConfigConnectionString(connectionString)
	if err := component.Render(c.Request().Context(), c.Response().Writer); err != nil {
		log.Fatalf("Error rendering connection string: %e", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

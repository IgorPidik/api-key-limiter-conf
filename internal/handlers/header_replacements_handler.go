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

type CreateHeaderReplacementForm struct {
	HeaderName  string `form:"header-name" validate:"required"`
	HeaderValue string `form:"header-value" validate:"required"`
}

type HeaderReplacementsHandler struct {
	db       *database.DatabaseHandler
	decoder  *form.Decoder
	validate *validator.Validate
}

func NewHeaderReplacementsHandler(db *database.DatabaseHandler) *HeaderReplacementsHandler {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return &HeaderReplacementsHandler{db, form.NewDecoder(), validate}
}

func (h *HeaderReplacementsHandler) processForm(c echo.Context) (*CreateHeaderReplacementForm, forms.FormErrors, error) {
	if c.Request().ParseForm() != nil {
		return nil, nil, echo.NewHTTPError(http.StatusBadRequest)
	}

	var headerForm CreateHeaderReplacementForm
	if err := h.decoder.Decode(&headerForm, c.Request().Form); err != nil {
		log.Fatalf("Error decoding CreateHeaderReplacementForm: %e", err)
		return nil, nil, echo.NewHTTPError(http.StatusInternalServerError)
	}

	if validationErr := h.validate.Struct(headerForm); validationErr != nil {
		errors := make(forms.FormErrors)
		for _, err := range validationErr.(validator.ValidationErrors) {
			errors[err.Field()] = err.Tag()
		}
		return nil, errors, nil

	}

	return &headerForm, nil, nil
}

func (h *HeaderReplacementsHandler) CreateHeaderReplacement(c echo.Context) error {
	project, ok := c.Get("project").(*models.Project)
	if !ok {
		log.Println("Missing project instance in the context")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	config, ok := c.Get("config").(*models.Config)
	if !ok {
		log.Println("Missing config instance in the context")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	headerForm, formErrs, processingErr := h.processForm(c)
	if processingErr != nil {
		return processingErr
	}

	if formErrs != nil {
		c.Response().Header().Set("HX-Reswap", "outerHTML")
		c.Response().Header().Set("HX-Retarget", "#"+projects_components.GetCreateHeaderFormID(config.ID))
		component := projects_components.CreateHeaderReplacement(project.ID, config.ID, formErrs)
		if err := component.Render(c.Request().Context(), c.Response().Writer); err != nil {
			log.Fatalf("Error rendering created header replacement: %e", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		return nil
	}

	encryptedValue, encryptErr := utils.EncryptData(headerForm.HeaderValue)
	if encryptErr != nil {
		log.Fatalf("Failed to encrypt header value: %e", encryptErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	replacement, replacementErr := h.db.CreateHeaderReplacement(config.ID, headerForm.HeaderName, encryptedValue)
	if replacementErr != nil {
		log.Fatalf("Failed to create headerReplacement: %e", replacementErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	component := projects_components.HeaderReplacement(project.ID, *replacement)
	if err := component.Render(c.Request().Context(), c.Response().Writer); err != nil {
		log.Fatalf("Error rendering created header replacement: %e", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

func (h *HeaderReplacementsHandler) DeleteHeaderReplacement(c echo.Context) error {
	header, ok := c.Get("header").(*models.HeaderReplacement)
	if !ok {
		log.Println("Missing header replacement instance in the config")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if deleteErr := h.db.DeleteHeaderReplacement(header.ID); deleteErr != nil {
		log.Fatalf("Failed to delete header: %e", deleteErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

func (h *HeaderReplacementsHandler) GetHeaderReplacementValue(c echo.Context) error {
	header, ok := c.Get("header").(*models.HeaderReplacement)
	if !ok {
		log.Println("Missing header replacement instance in the context")
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	decryptedHeaderValue, err := utils.DecryptData(header.HeaderValue)
	if err != nil {
		log.Fatalf("failed to decrypt header value: %e", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.String(http.StatusOK, decryptedHeaderValue)
}

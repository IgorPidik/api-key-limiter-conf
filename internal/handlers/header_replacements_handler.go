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

func (h *HeaderReplacementsHandler) CreateHeaderReplacement(c echo.Context) error {
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

	if c.Request().ParseForm() != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	var headerForm CreateHeaderReplacementForm
	if err := h.decoder.Decode(&headerForm, c.Request().Form); err != nil {
		log.Fatalf("Error decoding CreateHeaderReplacementForm: %e", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if validationErr := h.validate.Struct(headerForm); validationErr != nil {
		errors := make(forms.FormErrors)
		for _, err := range validationErr.(validator.ValidationErrors) {
			errors[err.Field()] = err.Tag()
		}

		c.Response().Header().Set("HX-Reswap", "outerHTML")
		c.Response().Header().Set("HX-Retarget", "#"+projects_components.GetCreateHeaderFormID(configID))
		component := projects_components.CreateHeaderReplacement(projectID, configID, errors)
		if err := component.Render(c.Request().Context(), c.Response().Writer); err != nil {
			log.Fatalf("Error rendering created header replacement: %e", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}

	encryptedValue, encryptErr := utils.EncryptData(headerForm.HeaderValue)
	if encryptErr != nil {
		log.Fatalf("Failed to encrypt header value: %e", encryptErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	replacement, replacementErr := h.db.CreateHeaderReplacement(configID, headerForm.HeaderName, encryptedValue)
	if replacementErr != nil {
		log.Fatalf("Failed to create headerReplacement: %e", replacementErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	component := projects_components.HeaderReplacement(projectID, *replacement)
	if err := component.Render(c.Request().Context(), c.Response().Writer); err != nil {
		log.Fatalf("Error rendering created header replacement: %e", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

func (h *HeaderReplacementsHandler) DeleteHeaderReplacement(c echo.Context) error {
	configID, configIDErr := uuid.Parse(c.Param("configId"))
	if configIDErr != nil {
		log.Fatalf("Invalid config id: %e", configIDErr)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid config id")
	}

	headerID, headerIDErr := uuid.Parse(c.Param("headerId"))
	if headerIDErr != nil {
		log.Fatalf("Invalid header id: %e", headerIDErr)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid header id")
	}

	if deleteErr := h.db.DeleteHeaderReplacement(configID, headerID); deleteErr != nil {
		log.Fatalf("Failed to delete header: %e", deleteErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

func (h *HeaderReplacementsHandler) GetHeaderReplacementValue(c echo.Context) error {
	headerID, headerIDErr := uuid.Parse(c.Param("headerId"))
	if headerIDErr != nil {
		log.Fatalf("Invalid header id: %e", headerIDErr)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid header id")
	}

	header, err := h.db.GetHeaderReplacement(headerID)
	if err != nil {
		log.Fatalf("failed to fetch header replacement data: %e", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	decryptedHeaderValue, err := utils.DecryptData(header.HeaderValue)
	if err != nil {
		log.Fatalf("failed to decrypt header value: %e", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.String(http.StatusOK, decryptedHeaderValue)
}

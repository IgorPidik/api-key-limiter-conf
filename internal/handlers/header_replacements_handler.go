package handlers

import (
	"configuration-management/internal/database"
	"configuration-management/internal/forms"
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

func (h *HeaderReplacementsHandler) ListHeaderReplacements(c echo.Context) error {
	// projectID, projectIDErr := uuid.Parse(c.Param("id"))
	// if projectIDErr != nil {
	// 	log.Fatalf("Invalid project id: %e", projectIDErr)
	// 	return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	// }

	// configID, configIDErr := uuid.Parse(c.Param("configId"))
	// if configIDErr != nil {
	// 	log.Fatalf("Invalid config id: %e", configIDErr)
	// 	return echo.NewHTTPError(http.StatusBadRequest, "invalid config id")
	// }
	//
	// headers, listErr := h.db.ListHeaderReplacements(configID)
	// if listErr != nil {
	// 	log.Fatalf("Failed to list header replacements: %e", listErr)
	// 	return echo.NewHTTPError(http.StatusInternalServerError)
	// }

	// component := projects_components.ListHeaderReplacements(headers)
	// renderErr := component.Render(c.Request().Context(), c.Response().Writer)
	// if renderErr != nil {
	// 	log.Fatalf("Error rendering in ListProjects: %e", renderErr)
	// 	return echo.NewHTTPError(http.StatusInternalServerError)
	// }

	return nil
}

func (h *HeaderReplacementsHandler) CreateHeaderReplacement(c echo.Context) error {
	// projectID, idErr := uuid.Parse(c.Param("id"))
	// if idErr != nil {
	// 	log.Fatalf("Invalid project id: %e", idErr)
	// 	return echo.NewHTTPError(http.StatusBadRequest, "invalid project id")
	// }

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
		// TODO: handle form errors

		// c.Response().Header().Set("HX-Reswap", "outerHTML")
		// c.Response().Header().Set("HX-Retarget", "#"+projects_components.GetCreateConfigFormID(projectID))
		// component := projects_components.CreateHeaderReplacementForm(projectID, errors)
		// if err := component.Render(c.Request().Context(), c.Response().Writer); err != nil {
		// 	log.Fatalf("Error rendering created config: %e", err)
		// 	return echo.NewHTTPError(http.StatusInternalServerError)
		// }
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}

	replacement, replacementErr := h.db.CreateHeaderReplacement(configID, headerForm.HeaderName, headerForm.HeaderValue)
	if replacementErr != nil {
		log.Fatalf("Failed to create headerReplacement: %e", replacementErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	component := projects_components.HeaderReplacement(*replacement)
	if err := component.Render(c.Request().Context(), c.Response().Writer); err != nil {
		log.Fatalf("Error rendering created header replacement: %e", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

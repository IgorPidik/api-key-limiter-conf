package projects_components

import (
	"configuration-management/internal/forms"
	"github.com/google/uuid"
	"strings"
)

func GetModalId(projectID uuid.UUID) string {
	return "create_config_modal_" + strings.Replace(projectID.String(), "-", "", -1)
}

func GetModalFormId(projectID uuid.UUID) string {
	return "close_modal_" + strings.Replace(projectID.String(), "-", "", -1)
}

func GetDetailsTabID(projectID uuid.UUID) string {
	return "details_tab_" + strings.Replace(projectID.String(), "-", "", -1)
}

func GetCreateConfigFormID(projectID uuid.UUID) string {
	return "create_config_form" + strings.Replace(projectID.String(), "-", "", -1)
}

func GetInputClass(fieldName string, errors forms.FormErrors, additionalClasses string) string {
	classes := "input input-bordered w-full"
	if _, ok := errors[fieldName]; ok {
		classes += " input-error"
	}

	if additionalClasses != "" {
		classes += " " + additionalClasses
	}

	return classes
}

func GetListHeaderReplacementID(configID uuid.UUID) string {
	return "list_headers" + strings.Replace(configID.String(), "-", "", -1)
}

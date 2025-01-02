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

func GetInputClass(fieldName string, errors forms.FormErrors) string {
	classes := "input input-bordered w-full"
	if _, ok := errors[fieldName]; ok {
		classes += " input-error"
	}

	return classes
}

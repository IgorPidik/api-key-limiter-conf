package projects_components

import (
	"strings"

	"github.com/google/uuid"
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

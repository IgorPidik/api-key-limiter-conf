package database

import (
	"configuration-management/internal/models"
	"fmt"

	"github.com/google/uuid"
)

func (s *DatabaseHandler) ListProjects(userID uuid.UUID) ([]models.Project, error) {
	query := `
		SELECT id, name, access_key
		FROM projects
		WHERE user_id = $1
		ORDER BY timestamp DESC
	`

	rows, err := s.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %v", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		if err := rows.Scan(&project.ID, &project.Name, &project.AccessKey); err != nil {
			return nil, fmt.Errorf("failed to scan project row: %v", err)
		}
		// list configs
		configs, configsErr := s.ListConfigs(project.ID)
		if configsErr != nil {
			return nil, fmt.Errorf("failed to list configs for projectID: %s:  %v", project.ID.String(), configsErr)
		}
		project.Configs = configs

		projects = append(projects, project)
	}

	return projects, nil
}

func (s *DatabaseHandler) GetProject(projectID uuid.UUID) (*models.Project, error) {
	query := `
		SELECT id, name, access_key, user_id
		FROM projects
		WHERE id = $1
	`

	var project models.Project
	if err := s.DB.QueryRow(query, projectID).Scan(&project.ID, &project.Name,
		&project.AccessKey, &project.UserID); err != nil {
		return nil, fmt.Errorf("failed to query project: %v", err)
	}

	return &project, nil
}

func (s *DatabaseHandler) CreateProject(name string, accessKey string, userID uuid.UUID) (*models.Project, error) {
	query := `
		INSERT into projects (name, access_key, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, name, access_key 
	`

	var project models.Project
	err := s.DB.QueryRow(query, name, accessKey, userID).Scan(&project.ID, &project.Name, &project.AccessKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %v", err)
	}

	return &project, nil
}

func (s *DatabaseHandler) DeleteProject(projectID uuid.UUID) error {
	query := `
		DELETE FROM projects WHERE id=$1
	`
	_, err := s.DB.Exec(query, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %v", err)
	}

	return nil
}

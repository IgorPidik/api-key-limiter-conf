package database

import (
	"configuration-management/internal/models"
	"fmt"

	"github.com/google/uuid"
)

func (s *DatabaseHandler) GetConfig(configID uuid.UUID) (*models.Config, error) {
	query := `
		SELECT id, project_id, name, limit_requests_count, limit_duration
		FROM configs
		WHERE id = $1
	`
	var config models.Config
	if err := s.DB.QueryRow(query, configID).Scan(
		&config.ID, &config.ProjectID, &config.Name,
		&config.LimitNumberOfRequests, &config.LimitPer,
	); err != nil {
		return nil, fmt.Errorf("failed to scan config: %v", err)
	}

	return &config, nil
}

func (s *DatabaseHandler) ListConfigs(projectID uuid.UUID) ([]models.Config, error) {
	query := `
		SELECT id, project_id, name, limit_requests_count, limit_duration
		FROM configs
		WHERE project_id = $1
	`

	rows, err := s.DB.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %v", err)
	}
	defer rows.Close()

	var configs []models.Config
	for rows.Next() {
		var config models.Config
		if err := rows.Scan(
			&config.ID, &config.ProjectID, &config.Name,
			&config.LimitNumberOfRequests, &config.LimitPer,
		); err != nil {
			return nil, fmt.Errorf("failed to scan config row: %v", err)
		}
		// list header replacements
		replacements, replacementsErr := s.ListHeaderReplacements(config.ID)
		if replacementsErr != nil {
			return nil, fmt.Errorf("failed to list header replacements for configID: %s:  %v", config.ID.String(), replacementsErr)
		}
		config.HeaderReplacements = replacements
		configs = append(configs, config)
	}

	return configs, nil
}

func (s *DatabaseHandler) CreateConfig(projectID uuid.UUID, name string,
	numberOfRequests int, per string) (*models.Config, error) {
	query := `
		INSERT into configs (project_id, name, limit_requests_count, limit_duration)
		VALUES ($1, $2, $3, $4) 
		RETURNING id, project_id, name, limit_requests_count, limit_duration
	`
	var config models.Config
	if err := s.DB.QueryRow(query, projectID, name, numberOfRequests, per).Scan(
		&config.ID, &config.ProjectID, &config.Name,
		&config.LimitNumberOfRequests, &config.LimitPer,
	); err != nil {
		return nil, fmt.Errorf("failed to create config: %v", err)
	}

	return &config, nil
}

func (s *DatabaseHandler) DeleteConfig(configID uuid.UUID) error {
	query := `
		DELETE FROM configs WHERE id=$1
	`
	_, err := s.DB.Exec(query, configID)
	if err != nil {
		return fmt.Errorf("failed to delete config: %v", err)
	}

	return nil
}

package database

import (
	"configuration-management/internal/models"
	"fmt"

	"github.com/google/uuid"
)

func (s *DatabaseHandler) ListHeaderReplacements(configID uuid.UUID) ([]models.HeaderReplacement, error) {
	query := `
		SELECT id, config_id, header_name, header_value
		FROM header_replacements
		WHERE config_id = $1
	`
	rows, err := s.DB.Query(query, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to query header replacements: %v", err)
	}
	defer rows.Close()

	var replacements []models.HeaderReplacement
	for rows.Next() {
		var replacement models.HeaderReplacement
		if err := rows.Scan(
			&replacement.ID, &replacement.ConfigID, &replacement.HeaderName, &replacement.HeaderValue,
		); err != nil {
			return nil, fmt.Errorf("failed to scan config row: %v", err)
		}

		replacements = append(replacements, replacement)
	}

	return replacements, nil
}

func (s *DatabaseHandler) GetHeaderReplacement(headerID uuid.UUID) (*models.HeaderReplacement, error) {
	query := `
		SELECT id, config_id, header_name, header_value
		FROM header_replacements
		WHERE id = $1
	`
	var replacement models.HeaderReplacement
	if err := s.DB.QueryRow(query, headerID).Scan(
		&replacement.ID, &replacement.ConfigID, &replacement.HeaderName, &replacement.HeaderValue); err != nil {
		return nil, fmt.Errorf("failed to get header replacement: %v", err)
	}

	return &replacement, nil
}

func (s *DatabaseHandler) CreateHeaderReplacement(configID uuid.UUID, name string, value string) (*models.HeaderReplacement, error) {
	query := `
		INSERT INTO header_replacements (config_id, header_name, header_value)
		VALUES ($1, $2, $3)
		RETURNING id, config_id, header_name, header_value
	`
	var replacement models.HeaderReplacement
	if err := s.DB.QueryRow(query, configID, name, value).Scan(
		&replacement.ID, &replacement.ConfigID, &replacement.HeaderName, &replacement.HeaderValue,
	); err != nil {
		return nil, fmt.Errorf("failed to create header replacement: %v", err)
	}

	return &replacement, nil
}

func (s *DatabaseHandler) DeleteHeaderReplacement(headerID uuid.UUID) error {
	query := `
		DELETE FROM header_replacements WHERE id=$1
	`
	_, err := s.DB.Exec(query, headerID)
	if err != nil {
		return fmt.Errorf("failed to delete header: %v", err)
	}

	return nil
}

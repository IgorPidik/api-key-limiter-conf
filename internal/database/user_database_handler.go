package database

import (
	"configuration-management/internal/models"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func (s *DatabaseHandler) CreateUser(oauth2ID int, name string, avatarUrl string) (*models.User, error) {
	query := `
		INSERT INTO users (oauth2_id, name, avatarUrl)
		VALUES ($1, $2, $3)
		ON CONFLICT (oauth2_id) DO NOTHING;
	`

	if _, err := s.DB.Exec(query, oauth2ID, name, avatarUrl); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return s.GetUserByOAuth2ID(oauth2ID)
}

func (s *DatabaseHandler) GetUserByOAuth2ID(oauth2ID int) (*models.User, error) {
	query := `
		SELECT id, oauth2_id, name, avatarUrl FROM users WHERE oauth2_id=$1;
	`

	var user models.User
	if err := s.DB.QueryRow(query, oauth2ID).Scan(
		&user.ID, &user.OAuth2ID, &user.Name, &user.AvatarUrl); err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}

func (s *DatabaseHandler) GetUser(userID uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, oauth2_id, name, avatarUrl FROM users WHERE id=$1;
	`

	var user models.User
	if err := s.DB.QueryRow(query, userID).Scan(
		&user.ID, &user.OAuth2ID, &user.Name, &user.AvatarUrl); err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}

func (s *DatabaseHandler) CreateUserSession(userID uuid.UUID) (*models.Session, error) {
	query := `
		INSERT INTO user_sessions (user_id) VALUES ($1)
		RETURNING id, user_id, created_at
	`

	var session models.Session
	if err := s.DB.QueryRow(query, userID).Scan(&session.ID, &session.UserID, &session.CreatedAt); err != nil {
		return nil, fmt.Errorf("failed to create user session: %v", err)
	}

	return &session, nil
}

func (s *DatabaseHandler) GetUserSession(sessionID uuid.UUID) (*models.Session, error) {
	query := `
		SELECT id, user_id, created_at FROM user_sessions WHERE id = $1 
	`

	var session models.Session
	if err := s.DB.QueryRow(query, sessionID).Scan(&session.ID, &session.UserID, &session.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user session: %v", err)
	}

	return &session, nil
}

package database

import (
	"configuration-management/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

type DatabaseHandler struct {
	DB *sql.DB
}

var (
	database   = os.Getenv("DB_DATABASE")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_PORT")
	host       = os.Getenv("DB_HOST")
	schema     = os.Getenv("DB_SCHEMA")
	dbInstance *DatabaseHandler
)

func New() *DatabaseHandler {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &DatabaseHandler{
		DB: db,
	}
	return dbInstance
}

func (s *DatabaseHandler) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	err := s.DB.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.DB.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

func (s *DatabaseHandler) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.DB.Close()
}

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

func (s *DatabaseHandler) DeleteConfig(projectID uuid.UUID, configID uuid.UUID) error {
	query := `
		DELETE FROM configs WHERE id=$1 AND project_id = $2
	`
	_, err := s.DB.Exec(query, configID, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete config: %v", err)
	}

	return nil
}

func (s *DatabaseHandler) ListHeaderReplacements(configID uuid.UUID) ([]models.HeaderReplacement, error) {
	query := `
		SELECT id, config_id, header_name, header_value
		FROM header_replacements
		WHERE config_id = $1
	`
	rows, err := s.DB.Query(query, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %v", err)
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

func (s *DatabaseHandler) DeleteHeaderReplacement(configID uuid.UUID, headerID uuid.UUID) error {
	query := `
		DELETE FROM header_replacements WHERE id=$1 AND config_id = $2
	`
	_, err := s.DB.Exec(query, headerID, configID)
	if err != nil {
		return fmt.Errorf("failed to delete header: %v", err)
	}

	return nil
}

func (s *DatabaseHandler) CreateUser(oauth2ID int, name string) (*models.User, error) {
	query := `
		INSERT INTO users (oauth2_id, name)
		VALUES ($1, $2)
		ON CONFLICT (oauth2_id) DO NOTHING;
	`

	if _, err := s.DB.Exec(query, oauth2ID, name); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return s.GetUser(oauth2ID)
}

func (s *DatabaseHandler) GetUser(oauth2ID int) (*models.User, error) {
	query := `
		SELECT id, oauth2_id, name FROM users WHERE oauth2_id=$1;
	`

	var user models.User
	if err := s.DB.QueryRow(query, oauth2ID).Scan(
		&user.ID, &user.OAuth2ID, &user.Name); err != nil {
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

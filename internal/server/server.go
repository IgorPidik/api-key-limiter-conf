package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"configuration-management/internal/database"
	"configuration-management/internal/handlers"
)

type Server struct {
	port int

	db              *database.DatabaseHandler
	projectsHandler *handlers.ProjectHandler
	configHandler   *handlers.ConfigHandler
	headersHandler  *handlers.HeaderReplacementsHandler
	loginHandler    *handlers.LoginHandler
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	db := database.New()
	NewServer := &Server{
		port:            port,
		projectsHandler: handlers.NewProjectHandler(db),
		headersHandler:  handlers.NewHeaderReplacementsHandler(db),
		loginHandler:    handlers.NewLoginHandler(db),
		configHandler:   handlers.NewConfigHandler(db),
		db:              db,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf("localhost:%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

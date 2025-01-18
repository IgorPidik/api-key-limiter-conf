package server

import (
	"net/http"

	"configuration-management/web"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	fileServer := http.FileServer(http.FS(web.Files))
	e.GET("/assets/*", echo.WrapHandler(fileServer))

	e.GET("/login", s.loginHandler.Login)
	e.GET("/auth/github", s.loginHandler.LoginWithGitlab)
	e.GET("/auth/github/callback", s.loginHandler.GithubCallback)

	e.GET("/projects", s.projectsHandler.ListProjects, s.UserAuth)
	e.POST("/projects", s.projectsHandler.CreateProject)
	e.DELETE("/projects/:id", s.projectsHandler.DeleteProject)

	e.POST("/projects/:id/configs", s.projectsHandler.CreateConfig)
	e.DELETE("/projects/:id/configs/:configId", s.projectsHandler.DeleteConfig)

	e.POST("/projects/:id/configs/:configId/headers", s.headersHandler.CreateHeaderReplacement)
	e.DELETE("/projects/:id/configs/:configId/headers/:headerId", s.headersHandler.DeleteHeaderReplacement)

	e.GET("/health", s.healthHandler)

	return e
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}

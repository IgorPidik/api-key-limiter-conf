package server

import (
	"net/http"
	"os"

	"configuration-management/web"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))))
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
	e.GET("/health", s.healthHandler)

	e.GET("/login", s.loginHandler.Login)
	e.GET("/logout", s.loginHandler.Logout)
	e.GET("/auth/github", s.loginHandler.LoginWithGitlab)
	e.GET("/auth/github/callback", s.loginHandler.GithubCallback)

	projectsGroup := e.Group("/projects", s.UserAuth)
	projectsGroup.GET("", s.projectsHandler.ListProjects, s.UserAuth)
	projectsGroup.POST("", s.projectsHandler.CreateProject)

	projectActionsGroup := projectsGroup.Group("/:id", s.ProjectBelongsToLoggedUser)
	projectActionsGroup.DELETE("", s.projectsHandler.DeleteProject)

	projectActionsGroup.POST("/configs", s.projectsHandler.CreateConfig)

	configsGroup := projectActionsGroup.Group("/configs/:configId", s.ConfigBelongToProject)
	configsGroup.DELETE("", s.projectsHandler.DeleteConfig)
	configsGroup.GET("/connection", s.projectsHandler.GetConfigConnection)

	configsGroup.POST("/headers", s.headersHandler.CreateHeaderReplacement)
	configsGroup.DELETE("/headers/:headerId", s.headersHandler.DeleteHeaderReplacement)
	configsGroup.GET("/headers/:headerId/value", s.headersHandler.GetHeaderReplacementValue)

	return e
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}

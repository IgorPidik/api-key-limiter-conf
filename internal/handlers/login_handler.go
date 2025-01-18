package handlers

import (
	"configuration-management/internal/database"
	"configuration-management/internal/models"
	"configuration-management/web/login_components"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type LoginHandler struct {
	conf *oauth2.Config
	db   *database.DatabaseHandler
}

func NewLoginHandler(db *database.DatabaseHandler) *LoginHandler {
	conf := &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{},
		Endpoint:     github.Endpoint,
	}
	return &LoginHandler{conf, db}
}

func (l *LoginHandler) Login(c echo.Context) error {
	component := login_components.Login()
	renderErr := component.Render(c.Request().Context(), c.Response().Writer)
	if renderErr != nil {
		log.Fatalf("Error rendering in ListProjects: %e", renderErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

func (l *LoginHandler) LoginWithGitlab(c echo.Context) error {
	// Generate a random state for CSRF protection and set it in a cookie.
	state, err := randString(16)
	if err != nil {
		panic(err)
	}

	cookie := &http.Cookie{
		Name:     "state",
		Value:    state,
		Path:     "/",
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   c.Request().TLS != nil,
		HttpOnly: true,
	}
	redirectURL := l.conf.AuthCodeURL(state)
	http.SetCookie(c.Response().Writer, cookie)
	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (l *LoginHandler) GithubCallback(c echo.Context) error {
	state, err := c.Request().Cookie("state")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "State not found")
	}
	if c.Request().URL.Query().Get("state") != state.Value {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid state")
	}

	code := c.Request().URL.Query().Get("code")
	tok, err := l.conf.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("error exchanging token: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	client := l.conf.Client(context.Background(), tok)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		log.Printf("error fetching user data from github: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	var githubUser models.GithubUser
	json.NewDecoder(resp.Body).Decode(&githubUser)
	user, userErr := l.db.CreateUser(githubUser.Id, githubUser.Name)
	if userErr != nil {
		log.Printf("failed to create a user: %v\n", userErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	userSession, sessionErr := l.db.CreateUserSession(user.ID)
	if sessionErr != nil {
		log.Printf("failed to create user session: %v\n", sessionErr)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	// create session
	sess, err := session.Get("session", c)
	if err != nil {
		log.Printf("failed to create session cookie: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}

	sess.Values["session_id"] = userSession.ID.String()
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		log.Printf("failed to update session in request: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.Redirect(http.StatusPermanentRedirect, "/projects")
}

func randString(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

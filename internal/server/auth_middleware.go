package server

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func (s *Server) UserAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			log.Printf("failed to read user session: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		sessionIDString, ok := sess.Values["session_id"]
		if !ok {
			log.Println("no session id")
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		sessionID, parseErr := uuid.Parse(sessionIDString.(string))
		if parseErr != nil {
			log.Println("unable to parse session id")
			return echo.NewHTTPError(http.StatusInternalServerError)

		}

		userSession, userSessionErr := s.db.GetUserSession(sessionID)
		if userSessionErr != nil {
			log.Printf("failed to get user session: %v\n", userSessionErr)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		if userSession == nil {
			return c.Redirect(http.StatusTemporaryRedirect, "/login")

		}

		c.Set("userID", userSession.UserID.String())
		return next(c)
	}
}

package sessmngt

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func SessionMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the session
		sess, err := session.Get("session", c)
		if err != nil {
			c.Logger().Error("Could not get the session")
			return c.String(http.StatusInternalServerError, "Internal Server Error")
		}

		// Get the database from the context
		db := c.Get("db").(*sql.DB)

		// Check if a session exists, redirect to the login page if not
		sentUsrIdInterface := sess.Values["userId"]
		sentExpiryTimeUnixInterface := sess.Values["expiryTimeUnix"]
		if sentUsrIdInterface == nil || sentExpiryTimeUnixInterface == nil {
			return c.Redirect(http.StatusMovedPermanently, "/logout")
		}

		// Try to convert the session values to the proper types
		sentUsrId, okId := sentUsrIdInterface.(int)
		sentExpiryTimeUnix, okTime := sentExpiryTimeUnixInterface.(int64)
		if !okId || !okTime {
			return c.Redirect(http.StatusMovedPermanently, "/logout")
		}

		// Validate if the session exists in the database,
		// has a valid expiry time, user id, and session id
		usrSess, err := GetSession(db, sess.ID)
		if err != nil || usrSess.UserId != sentUsrId || usrSess.ExpiryTimeUnix != sentExpiryTimeUnix || usrSess.SessId != sess.ID {
			return c.Redirect(http.StatusMovedPermanently, "/login")
		}

		// If all of the above goes ok, we can proceeed to the next middleware function
		return next(c)
	}
}

package sessmngt

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func SessionMiddleware(nextContext echo.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the session
		_, err := session.Get("session", c)
		if err != nil {
			c.Logger().Error("Could not get the session")
			return c.String(http.StatusInternalServerError, "Internal Server Error")
		}

		// Get the database from the context
		// db := c.Get("db").(*sql.DB)

		// Validate if the session exists in the database,
		// has a valid expiry time, and has a user id

		// userSession := Get

		return nil
	}
}

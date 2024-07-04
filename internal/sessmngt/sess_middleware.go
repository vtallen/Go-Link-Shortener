package sessmngt

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

/*
* Function: SessionMiddleware
*
* Parameters: next echo.HandlerFunc - The next middleware function to call in the chain of registered functions
*
* Returns: echo.HandlerFunc - The closure function that validates user sessions
*
* Description: A middleware function that can be applied to routes that will ensure that
*              a session cookie exists and all data within it matches what is stored in the
*              database. Failed authentication will lead to the user being sent to the /logout
*              endpoint which invalidates their session cookie and sends them to the /login endpoint
*
 */
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

		// Check if a session exists in the current session cookie, redirect to the login page if not
		sentUsrIdInterface := sess.Values["userId"]
		sentExpiryTimeUnixInterface := sess.Values["expiryTimeUnix"]
		sentSessIdInterface := sess.Values["sessId"]
		if sentUsrIdInterface == nil || sentExpiryTimeUnixInterface == nil || sentSessIdInterface == nil {
			return c.Redirect(http.StatusMovedPermanently, "/logout")
		}

		// // Try to convert the session cookie values to the proper types
		sentUsrId, okId := sentUsrIdInterface.(int)
		sentExpiryTimeUnix, okTime := sentExpiryTimeUnixInterface.(int64)
		sentSessId, okSessId := sentSessIdInterface.(int64)
		// Type casts return bool not error
		if !okId || !okTime || !okSessId {
			return c.Redirect(http.StatusMovedPermanently, "/logout")
		}

		// // Validate if the session exists in the database,
		// // has a valid expiry time, user id, and session id
		usrSess, err := GetSessionStruct(db, sentSessId)
		if err != nil || usrSess == nil {
			return c.Redirect(http.StatusMovedPermanently, "/logout")
		}

		if usrSess.UserId != sentUsrId || usrSess.ExpiryTimeUnix != sentExpiryTimeUnix || usrSess.SessId != sentSessId {
			return c.Redirect(http.StatusMovedPermanently, "/logout")
		}

		// If all of the above goes ok, we can proceeed to the next middleware function
		// Now we can assume in all further function calls that the session is valid
		// As this function gets called each time the user attempts to visit a page
		return next(c)
	}
}

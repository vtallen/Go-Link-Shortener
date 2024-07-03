package sessmngt

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// TODO : make it so expired sessions get removed from the database, check each time this function is run
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
			// return c.Redirect(http.StatusMovedPermanently, "/logout")
			return c.String(http.StatusInternalServerError, "values do not exist in cookie")
		}

		// // Try to convert the session cookie values to the proper types
		sentUsrId, okId := sentUsrIdInterface.(int)
		sentExpiryTimeUnix, okTime := sentExpiryTimeUnixInterface.(int64)
		sentSessId, okSessId := sentSessIdInterface.(int64)
		// Type casts return bool not error
		if !okId || !okTime || !okSessId {
			return c.String(http.StatusInternalServerError, "unable to convert values")
			// return c.Redirect(http.StatusMovedPermanently, "/logout")
		}

		// // Validate if the session exists in the database,
		// // has a valid expiry time, user id, and session id
		usrSess, err := GetSessionStruct(db, sentSessId)
		if err != nil || usrSess == nil {
			return c.String(http.StatusInternalServerError, "err in getting session struct "+err.Error())
		}

		if usrSess.UserId != sentUsrId || usrSess.ExpiryTimeUnix != sentExpiryTimeUnix || usrSess.SessId != sentSessId {
			// debug := `sentUsrId = ` + string(sentUsrId) + "\nusrSess.UserId = " + string(usrSess.UserId)
			return c.String(http.StatusInternalServerError, "values do not match")
			return c.Redirect(http.StatusMovedPermanently, "/logout") // TODO : should this redirect to /logout? Check after valiadating everything else
		}

		// If all of the above goes ok, we can proceeed to the next middleware function
		// Now we can assume in all further function calls that the session is valid
		// As this function gets called each time the user attempts to visit a page
		return next(c)
	}
}

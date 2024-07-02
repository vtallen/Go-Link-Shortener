package sessmngt

import (
	"crypto/rand"
	"math"
	"math/big"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/vtallen/go-link-shortener/internal/conf"
	"golang.org/x/crypto/bcrypt"
)

// Apparently the session ID for gorilla sessions does not get populated
// if using a normal cookie, see: https://github.com/gorilla/sessions/issues/224
// This code was found and modified from: https://reintech.io/blog/generating-random-numbers-in-go
func GenSessionId() (int64, error) {
	max := big.NewInt(math.MaxInt64)
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}

	return randomNumber.Int64(), nil
}

func InvalidateSession(sess *sessions.Session, c echo.Context) error {
	// Invalidates the session so it gets deleted
	sess.Options.MaxAge = -1

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}

// This function should take the session and the supplied UserSession struct
// Then store those values in the session and save it
func SetSessionCookie(sess *sessions.Session, user *UserLogin, config *conf.Config) (*UserSession, error) {
	sess.Options = &sessions.Options{
		MaxAge:   86400 * config.Auth.CookieMaxAgeDays,
		HttpOnly: true,
	}

	// Create the user session struct
	var userSession UserSession
	userSession.StoreExpiryTime(86400 * int64(config.Auth.CookieMaxAgeDays))
	userSession.UserId = user.Id
	sessId, err := GenSessionId()
	if err != nil {
		return nil, err
	}
	userSession.SessId = sessId

	// Set the user session values into the session cookie
	sess.Values["sessId"] = userSession.SessId
	sess.Values["expiryTimeUnix"] = userSession.ExpiryTimeUnix
	sess.Values["userId"] = user.Id

	return &userSession, nil
}

/*
* Function: HashPassword
*
* Parameters: password string - The password to hash
*
* Description: This function takes a password and hashes it.
*
* Returns: string - The hashed password
*          error - If there is an error hashing the password, the error is returned.
 */

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

/*
* Function: CheckPassword
*
* Parameters: hash string - The hashed password to compare to
*             password string - The password to compare
*
* Description: This function takes a hashed password and a password and compares them.
*
* Returns: error - If the passwords do not match, the error is returned.
 */
func CheckPassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

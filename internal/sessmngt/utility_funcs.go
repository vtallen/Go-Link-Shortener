package sessmngt

import (
	"crypto/rand"
	"errors"
	"math"
	"math/big"
	"net"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/meyskens/go-hcaptcha"
	"github.com/vtallen/go-link-shortener/internal/conf"
	"golang.org/x/crypto/bcrypt"
)

// Apparently the session ID for gorilla sessions does not get populated
// if using a normal cookie, see: https://github.com/gorilla/sessions/issues/224
// This code was found and modified from: https://reintech.io/blog/generating-random-numbers-in-go
/*
* Function: GenSessionId
*
* Parameters: None
*
* Returns: int64 - The generated session id
*          error - returned if the call to rand.Int fials
*
* Description: Generates a random int64 to be used as a session ID
*
 */
func GenSessionId() (int64, error) {
	max := big.NewInt(math.MaxInt64)
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}

	return randomNumber.Int64(), nil
}

/*
* Function: CheckCaptcha
*
* Parameters: c echo.Context - The context of the current request
*
* Returns:  error - returns nil if the captcha was verified
*
* Description: Validates that the captcha from a page that has one has been
*              successfully filled out
*
 */
func CheckCaptcha(c echo.Context, secretkey string) error {
	captchaResponse := c.FormValue("h-captcha-response")

	// Create a captcha object
	hc := hcaptcha.New(secretkey) //lint:ignore

	// Get the remote address of the request
	ip, _, err := net.SplitHostPort(c.Request().RemoteAddr)
	if err != nil {
		return err
	}

	// Verify the captcha response with hCaptcha's servers
	resp, err := hc.Verify(captchaResponse, ip)
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New("captcha failed")
	}

	return nil
}

/*
* Function: CookieExists
*
* Parameters: sess *sessions.Session - The gorilla sessions session to check for the existence of a cookie
*
* Returns: bool - true if the session cookie exists
*
* Description: Checks only if the values for a session cookie exist in the session and are not nil
*
 */
func CookieExists(sess *sessions.Session) bool {
	return sess.Values["sessId"] != nil || sess.Values["expiryTimeUnix"] != nil || sess.Values["userId"] != nil
}

/*
* Function: InvalidateSession
*
* Parameters: sess *sessions.Session - The gorilla sessions session to invalidate
*             c echo.Context - The context of the current request
*
* Returns: error
*
* Description: Sets all values of the session cookie to nil and saves it
*
 */
func InvalidateSession(sess *sessions.Session, c echo.Context) error {
	// Invalidates the session so it gets deleted
	sess.Options.MaxAge = -1
	sess.Values["sessId"] = nil
	sess.Values["expiryTimeUnix"] = nil
	sess.Values["userId"] = nil

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}

/*
* Function: SetSessionCookie
*
* Parameters: sess *sessions.Session - The gorilla sessions session to invalidate
*             user *UserLogin - The user which is being logged in
*             config *conf.Config - The configuration struct for the whole server
*
* Returns: *UserSession - A filled out session struct
*          error - Returns an error only if the call to GenSessionId fails
*
* Description: Sets all needed values in the session cookie and returns a filled out UserSesssion struct.
*              This function does not save and send the cookie back to the client
*
 */
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
* Returns: string - The hashed password
*          error - If there is an error hashing the password, the error is returned.
*
* Description: This function takes a password and hashes it.
*
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
* Returns: error - If the passwords do not match, the error is returned.
*
* Description: This function takes a hashed password and a password and compares them.
*
 */
func CheckPassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

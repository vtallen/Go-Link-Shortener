package sessmngt

import (
	"github.com/gorilla/sessions"
	"github.com/vtallen/go-link-shortener/internal/conf"
	"golang.org/x/crypto/bcrypt"
)

// This function should take the session and the supplied UserSession struct
// Then store those values in the session and save it
func CreateSessionCookie(sess *sessions.Session, userSession *UserSession, config conf.Config) {
	// // Validate that user is not logged in already
	// if sess.Values["userId"] != nil {
	// 	data.HasError = true
	// 	data.AlreadyLoggedIn = true
	// 	data.ErrorText = "User already logged in"
	// 	c.Logger().Info("User already logged in, email: " + email)
	// 	return c.Render(200, "login-form", data)
	// }

	// 86400 is the number of seconds in a day
	sess.Options = &sessions.Options{
		MaxAge:   86400 * config.Auth.CookieMaxAgeDays,
		HttpOnly: true,
	}

	sess.Values["userId"] = userSession.UserId
	sess.Values[""] = sess.ID

	// if err := sess.Save(c.Request(), c.Response()); err != nil {
	// 	data.HasError = true
	// 	data.ErrorText = "Error saving session"
	// 	data.LoginForm.Email = email
	// 	c.Logger().Info("Error saving session: " + err.Error() + " | email: " + email)
	// 	return c.Render(200, "login-form", data)
	// }
}

func TeardownSession() {
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

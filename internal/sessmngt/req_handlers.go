package sessmngt

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/vtallen/go-link-shortener/internal/conf"
	"github.com/vtallen/go-link-shortener/internal/pagestructs"
)

func HandleLoginPage(c echo.Context, data *pagestructs.LoginData, config *conf.Config) error {
	return c.Render(200, "login", data)
}

func HandleLoginSession(c echo.Context, db *sql.DB, data *pagestructs.LoginData, config *conf.Config) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	// TODO: validate email is in correct format, prevent sql injection. mail.ParseAddress I think?

	// Check if the user exists
	user, err := GetUserByEmail(db, email)
	if err != nil {
		data.HasError = true
		data.ErrorText = "Either the user does not exist or the password is incorrect"
		data.LoginForm.Email = email
		c.Logger().Info("failed login, user does not exist: " + email)
		return c.Render(200, "login-form", data)
	}

	// Check if the password is correct
	err = CheckPassword(user.Password, password)
	if err != nil {
		data.HasError = true
		data.ErrorText = "Either the user does not exist or the password is incorrect"
		data.LoginForm.Email = email
		c.Logger().Info("failed login for user: " + user.Email)
		return c.Render(200, "login-form", data)
	}

	// get the session
	sess, err := session.Get("session", c)
	if err != nil {
		data.HasError = true
		data.ErrorText = "Error creating session"
		data.LoginForm.Email = email
		c.Logger().Warn("Error creating session: " + err.Error() + " | email: " + email)
		return c.Render(200, "login-form", data)
	}

	// Validate that user is not logged in already
	if sess.Values["userId"] != nil {
		data.HasError = true
		data.AlreadyLoggedIn = true
		data.ErrorText = "User already logged in"
		c.Logger().Info("User already logged in, email: " + email)
		return c.Render(200, "login-form", data)
	}

	// create a user session in the database

	// sess.Options = &sessions.Options{
	// 	MaxAge:   86400 * config.Auth.CookieMaxAgeDays,
	// 	HttpOnly: true,
	// }

	// sess.Values["userId"] = user.Id

	userSession, err := SetSessionCookie(sess, user, config)
	if err != nil {
		data.HasError = true
		data.ErrorText = "Error creating session 1"
		c.Logger().Info("Error generating session ID: " + err.Error() + " | email: " + email)
		return c.Render(200, "login-form", data)
	}

	fmt.Println(userSession)

	err = userSession.StoreDB(db)
	if err != nil {
		data.HasError = true
		data.ErrorText = "Error creating session 2"
		c.Logger().Info("Error storing session in DB: " + err.Error() + " | email: " + email)
		return c.Render(200, "login-form", data)
	}

	// print("\n\n\n" + err.Error() + "\n\n\n")
	// store the session in the database

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		data.HasError = true
		data.ErrorText = "Error saving session 3"
		data.LoginForm.Email = email
		c.Logger().Info("Error saving session: " + err.Error() + " | email: " + email)
		return c.Render(200, "login-form", data)
	}

	data.HasError = false

	c.Response().Header().Set("HX-Redirect", "/user")
	return c.String(http.StatusMovedPermanently, "redirecting")
	// c.Response().Header().Set("HX-Push-URL", "/user")
	// return c.Redirect(http.StatusMovedPermanently, "/user")
}

func HandleLogout(c echo.Context, config *conf.Config) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return c.Render(http.StatusMovedPermanently, "error-page", pagestructs.ErrorPageData{ErrorText: "Error getting session, could not log out"})
	}

	// sess.Values["userId"] = nil
	// sess.Values["expiryTimeUnix"] = nil

	// if err := sess.Save(c.Request(), c.Response()); err != nil {
	// 	return c.Render(http.StatusMovedPermanently, "error-page", pagestructs.ErrorPageData{ErrorText: "Error saving session, could not log out"})
	// }

	err = InvalidateSession(sess, c)
	if err != nil {
		return c.Render(http.StatusMovedPermanently, "error-page", pagestructs.ErrorPageData{ErrorText: "Error saving session, could not log out"})
	}

	// Feels kind of hacky, but I could not find a better/more reliable solution.
	// Using the HTTP headers lead to some kind of race condition that messed up the session store.
	// Or I just don't know how to do it properly (the more likely reason)
	return c.String(200, `<script>window.location.href="/login"</script>`)
}

func HandleRegisterPage(c echo.Context, data *pagestructs.RegisterData, config *conf.Config) error {
	return c.Render(200, "register", data)
}

func HandleRegisterSession(c echo.Context, db *sql.DB, data *pagestructs.RegisterData, config *conf.Config) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	// validate email format
	_, err := mail.ParseAddress(email)
	if err != nil {
		data.HasError = true
		data.ErrorText = "Invalid email"
		return c.Render(200, "register-form", data)
	}

	splitEmail := strings.Split(email, "@")
	if len(splitEmail) != 2 {
		data.HasError = true
		data.ErrorText = "Invalid email"
		return c.Render(200, "register-form", data)
	}

	username := splitEmail[0]

	// Check if the user exists
	usr, err := GetUserByEmail(db, email)
	// fmt.Println(err + "\n\n\n")
	if err == nil {
		data.HasError = true
		data.ErrorText = "User already exists"
		c.Logger().Info("Register attempted with user that already exists, email: " + usr.Email)
		return c.Render(200, "register-form", data)
	}

	// Hash the password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		data.HasError = true
		data.ErrorText = "Error creating account"
		c.Logger().Warn("Error hashing password for new user " + email + ": " + err.Error())
		return c.Render(200, "register-form", data)
	}

	// Add the user to the database
	err = AddUser(db, email, username, hashedPassword, "user")
	if err != nil {
		data.HasError = true
		data.ErrorText = "Error adding user"
		c.Logger().Warn(err.Error())
		return c.Render(200, "register-form", data)
	}

	c.Logger().Info("Added user " + email + " to the database")

	// create the session
	// sess, err := session.Get("session", c)
	// if err != nil {
	// 	data.HasError = true
	// 	data.ErrorText = "Error creating session"
	// 	c.Logger().Warn("Error creating session: " + err.Error() + " | email: " + email)
	// 	return c.Render(200, "register-form", data)
	// }

	// id, err := GetUserId(db, email)
	// if err != nil {
	// 	data.HasError = true
	// 	data.ErrorText = "Error creating session"
	// 	c.Logger().Info("Error getting user id while creating session: " + err.Error() + " | email: " + email)
	// 	return c.Render(200, "register-form", data)
	// }

	// sess.Values["userId"] = id

	// sess.Options = &sessions.Options{
	// 	MaxAge:   86400 * config.Auth.CookieMaxAgeDays,
	// 	HttpOnly: true,
	// }

	// if err := sess.Save(c.Request(), c.Response()); err != nil {
	// 	data.HasError = true
	// 	data.ErrorText = "Error creating session"
	// 	c.Logger().Info("Error saving session: " + err.Error() + " | email: " + email)
	// 	return err
	// }

	data.HasError = false
	data.Success = true

	return c.Render(200, "register-form", data)
}

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

/*
* Function:
*
* Parameters: c echo.Context - The context for the current request
*             data *pagestructs.LoginData - The needed paged data for the template/functionality
*             config *conf.Config - The configuration struct for the server
*
* Returns: error
*
* Description: Handles serving the login page found in views/login.html from a GET request
*
 */
func HandleLoginPage(c echo.Context, data *pagestructs.LoginData, config *conf.Config) error {
	return c.Render(200, "login", data)
}

/*
* Function:
*
* Parameters: c echo.Context - The context for the current request
*             data *pagestructs.LoginData - The needed paged data for the template/functionality
*             config *conf.Config - The configuration struct for the server
*
* Returns: error
*
* Description: Handles a POST request made to /login to create the session cookie and redirect the user to the
*              user homepage
*
 */
func HandleLoginSession(c echo.Context, data *pagestructs.LoginData, config *conf.Config) error {
	db := c.Get("db").(*sql.DB)

	email := c.FormValue("email")
	password := c.FormValue("password")

	_, err := mail.ParseAddress(email)
	if err != nil {
		data.HasError = true
		data.ErrorText = "Invalid email"
		return c.Render(200, "login-form", data)
	}

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
		data.IsLoggedIn = true
		data.ErrorText = "User already logged in"
		c.Logger().Info("User already logged in, email: " + email)
		return c.Render(200, "login-form", data)
	}

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
}

/*
* Function:
*
* Parameters: c echo.Context - The context for the current request
*             config *conf.Config - The configuration struct for the server
*
* Returns: error
*
* Description: Handles a GET request made to the /logout endpoint. It invalidates
*              the session cookie if it exists then redirects the user to the /login endpoint
*
 */

func HandleLogout(c echo.Context, config *conf.Config) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return c.Render(http.StatusMovedPermanently, "error-page", pagestructs.ErrorPageData{ErrorText: "Error getting session, could not log out"})
	}

	err = InvalidateSession(sess, c)
	if err != nil {
		return c.Render(http.StatusMovedPermanently, "error-page", pagestructs.ErrorPageData{ErrorText: "Error saving session, could not log out"})
	}

	return c.Redirect(http.StatusFound, "/login")
}

/*
* Function: HandleRegisterPage
*
* Parameters: c echo.Context - The context for the current request
*             data *pagestructs.RegisterData - The needed paged data for the template/functionality
*             config *conf.Config - The configuration struct for the server
*
* Returns: error
*
* Description: Handles serving the register page html on endpoint /register
*
 */

func HandleRegisterPage(c echo.Context, data *pagestructs.RegisterData, config *conf.Config) error {
	sess, err := session.Get("session", c)
	if err != nil {
		data.HasError = true
		data.ErrorText = "Could not get the session"
		return c.Render(200, "register", data)
	}
	// This statement ensures that a user can only make an account if they are logged out
	if CookieExists(sess) {
		data.IsLoggedIn = true
		return c.Render(200, "register", data)
	}

	return c.Render(200, "register", data)
}

/*
* Function:
*
* Parameters: c echo.Context - The context for the current request
*             data *pagestructs.RegisterData - The needed paged data for the template/functionality
*             config *conf.Config - The configuration struct for the server
*
* Returns: error
*
* Description: Handles a POST request made to the /register endpoint from the register form
*
 */

func HandleRegisterSession(c echo.Context, data *pagestructs.RegisterData, config *conf.Config) error {
	db := c.Get("db").(*sql.DB)

	sess, err := session.Get("session", c)
	if err != nil {
		data.HasError = true
		data.ErrorText = "Could not get the session"
		return c.Render(200, "register-form", data)
	}
	if CookieExists(sess) {
		data.IsLoggedIn = true
		return c.Render(200, "register-form", data)
	}

	email := c.FormValue("email")
	password := c.FormValue("password")

	// validate email format
	_, err = mail.ParseAddress(email)
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

	data.HasError = false
	data.Success = true

	return c.Render(200, "register-form", data)
}

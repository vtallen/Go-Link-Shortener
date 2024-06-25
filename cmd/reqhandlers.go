package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/vtallen/go-link-shortener/pkg/codegen"
)

func HandleRedirect(c echo.Context, db *sql.DB, config *Config) error {
	shortcode := c.Param("shortcode")

	// Figure out the id
	id := codegen.UniverseToBaseTen(shortcode, config.Shortcodes.Universe)
	// Query the db for the link
	link, err := GetLink(db, id)
	if err != nil {
		errData := ErrorPageData{ErrorText: "404, link does not exist"}
		return c.Render(404, "error-page", errData) // Show the not found page if link does not exist
	}

	return c.Redirect(302, link.Url) // If a url exists, redirect the user to it
}

func HandleAddLink(c echo.Context, db *sql.DB, config *Config, data *IndexData) error {
	URL := c.FormValue("url")
	if URL != "" {
		// Validate the URL
		_, err := url.Parse(URL)
		if err != nil {
			data.ShortcodeForm.URL = URL
			data.ShortcodeForm.HasError = true
			return c.Render(200, "shortcode-form", data)
		}

		// Generate a random id for the table
		id := codegen.GenRandID(config.Shortcodes.Universe, config.Shortcodes.ShortcodeLength)
		// Create a shortcode from the id
		shortcode := codegen.BaseTenToUniverse(id, config.Shortcodes.Universe)
		// Put the link in the database
		err = AddLink(db, id, shortcode, URL)
		if err != nil {
			log.Println(err.Error())
			return c.String(500, "Error adding link to database")
		}

		// Set all of the data for the form to be displayed
		data.ShortcodeForm.Result = shortcode
		data.ShortcodeForm.URL = ""
		data.ShortcodeForm.HasError = false

		fmt.Printf("Added link id: %d | shortcode: %s | url: %s\n", id, shortcode, URL)

		return c.Render(200, "shortcode-form", data)
	}

	// This can only be hit if /create is visited without the form being filled out
	// This should not be possible because of client side validataion
	data.ShortcodeForm.URL = URL
	data.ShortcodeForm.Result = ""
	return c.Render(200, "shortcode-form", data)
}

func HandleLoginPage(c echo.Context, data *LoginData, config *Config) error {
	return c.Render(200, "login", data)
}

func HandleLoginSession(c echo.Context, db *sql.DB, data *LoginData, config *Config) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

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

	sess.Options = &sessions.Options{
		MaxAge:   86400 * config.Auth.CookieMaxAgeDays,
		HttpOnly: true,
	}

	sess.Values["userId"] = user.Id

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		data.HasError = true
		data.ErrorText = "Error saving session"
		data.LoginForm.Email = email
		c.Logger().Info("Error saving session: " + err.Error() + " | email: " + email)
		return c.Render(200, "login-form", data)
	}

	data.HasError = false
	// A kind of hacky solution for a redirect I found here: https://stackoverflow.com/questions/70618200/how-to-implement-a-redirect-with-htmx/70618201#70618201

	c.Response().Header().Set("HX-Redirect", "/user")
	return c.String(http.StatusMovedPermanently, "redirecting")
	// c.Response().Header().Set("HX-Push-URL", "/user")
	// return c.Redirect(http.StatusMovedPermanently, "/user")
}

func HandleLogout(c echo.Context, config *Config) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return c.Render(http.StatusMovedPermanently, "error-page", ErrorPageData{ErrorText: "Error getting session, could not log out"})
	}

	sess.Values["userId"] = nil

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return c.Render(http.StatusMovedPermanently, "error-page", ErrorPageData{ErrorText: "Error saving session, could not log out"})
	}

	// c.Response().Header().Set("HX-Redirect", "/login")
	// c.Response().Header().Set("HX-Refresh", "true")
	// c.Response().Header().Set("HX-Replace-Url", "/login")
	// return c.Redirect(http.StatusMovedPermanently, "/login")

	return c.String(200, `<script>window.location.href="/login"</script>`)
}

func HandleRegisterPage(c echo.Context, data *RegisterData, config *Config) error {
	return c.Render(200, "register", data)
}

func HandleRegisterSession(c echo.Context, db *sql.DB, data *RegisterData, config *Config) error {
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

	c.Logger().Warn("Added user " + email + " to the database")

	// create the session
	sess, err := session.Get("session", c)
	if err != nil {
		data.HasError = true
		data.ErrorText = "Error creating session"
		c.Logger().Warn("Error creating session: " + err.Error() + " | email: " + email)
		return c.Render(200, "register-form", data)
	}

	sess.Options = &sessions.Options{
		MaxAge:   86400 * config.Auth.CookieMaxAgeDays,
		HttpOnly: true,
	}

	id, err := GetUserId(db, email)
	if err != nil {
		data.HasError = true
		data.ErrorText = "Error creating session"
		c.Logger().Info("Error getting user id while creating session: " + err.Error() + " | email: " + email)
		return c.Render(200, "register-form", data)
	}

	sess.Values["userId"] = id

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		data.HasError = true
		data.ErrorText = "Error creating session"
		c.Logger().Info("Error saving session: " + err.Error() + " | email: " + email)
		return err
	}

	data.HasError = false
	data.Success = true

	return c.Render(200, "register-form", data)
}

func HandleUserPage(c echo.Context, db *sql.DB, data *UserPageData, config *Config) error {
	return c.Render(200, "user-homepage", data)
}

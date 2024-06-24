package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/mail"
	"net/url"
	"strings"

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
		return c.Render(404, "404", nil) // Show the not found page if link does not exist
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
	fmt.Print(email + password)

	// Check if the user exists
	user, err := GetUserByEmail(db, email)
	if err != nil {
		data.HasError = true
		data.ErrorText = "User does not exist"
		return c.Render(200, "login", data)
	}
	fmt.Println(user)

	// TODO finish

	return nil
}

func HandleRegisterPage(c echo.Context, data *RegisterData, config *Config) error {
	return c.Render(200, "register", data)
}

func HandleRegisterSession(c echo.Context, db *sql.DB, data *RegisterData, config *Config) error {
	email := c.FormValue("email")
	// fmt.Println(email)
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
		c.Logger().Error("Error hashing password for new user " + email + ": " + err.Error())
		return c.Render(200, "register-form", data)
	}

	// Add the user to the database
	err = AddUser(db, email, username, hashedPassword, "user")
	if err != nil {
		data.HasError = true
		data.ErrorText = "Error adding user"
		c.Logger().Error(err.Error())
		return c.Render(200, "register-form", data)
	}

	data.HasError = false
	data.Success = true
	c.Logger().Info("Added user " + email + " to the database")

	return c.Render(200, "register-form", data)
}

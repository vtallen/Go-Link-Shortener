package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/vtallen/go-link-shortener/internal/conf"
	"github.com/vtallen/go-link-shortener/internal/pagestructs"
	"github.com/vtallen/go-link-shortener/pkg/codegen"
)

func HandleRedirect(c echo.Context, db *sql.DB, config *conf.Config) error {
	shortcode := c.Param("shortcode")

	// Figure out the id
	id := codegen.UniverseToBaseTen(shortcode, config.Shortcodes.Universe)
	// Query the db for the link
	link, err := GetLink(db, id)
	if err != nil {
		errData := pagestructs.ErrorPageData{ErrorText: "404, link does not exist"}
		return c.Render(404, "error-page", errData) // Show the not found page if link does not exist
	}

	return c.Redirect(302, link.Url) // If a url exists, redirect the user to it
}

func HandleAddLink(c echo.Context, db *sql.DB, config *conf.Config, data *pagestructs.IndexData) error {
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

func HandleUserPage(c echo.Context, db *sql.DB, data *pagestructs.UserPageData, config *conf.Config) error {
	return c.Render(200, "user-homepage", data)
}

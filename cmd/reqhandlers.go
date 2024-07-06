package main

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/vtallen/go-link-shortener/internal/conf"
	"github.com/vtallen/go-link-shortener/internal/globalstructs"
	"github.com/vtallen/go-link-shortener/internal/sessmngt"
	"github.com/vtallen/go-link-shortener/pkg/codegen"
)

func HandleRedirect(c echo.Context, db *sql.DB, config *conf.Config) error {
	shortcode := c.Param("shortcode")

	// Figure out the id
	id := codegen.UniverseToBaseTen(shortcode, config.Shortcodes.Universe)
	// Query the db for the link
	link, err := GetLink(db, id)
	if err != nil {
		errData := globalstructs.ErrorPageData{ErrorText: "404, link does not exist"}
		return c.Render(http.StatusNotFound, "error-page", errData) // Show the not found page if link does not exist
	}

	return c.Redirect(http.StatusMovedPermanently, link.Url) // If a url exists, redirect the user to it
}

func HandleAddLink(c echo.Context, db *sql.DB, config *conf.Config, data *globalstructs.IndexData) error {
	URL := c.FormValue("url")
	if URL != "" {
		// Check the captcha if the user is not logged in
		if !data.IsLoggedIn {
			err := sessmngt.CheckCaptcha(c, config.HCaptcha.SecretKey)
			if err != nil {
				data.ShortcodeForm.HasError = true
				data.ShortcodeForm.ErrorText = "Please answer the captcha"
				return c.Render(http.StatusOK, "shortcode-form", data)
			}
		}

		// Validate the URL
		_, err := url.Parse(URL)
		if err != nil {
			data.ShortcodeForm.URL = URL
			data.ShortcodeForm.HasError = true
			data.ShortcodeForm.ErrorText = "There was an error submitting the URL, please try again"
			return c.Render(http.StatusOK, "shortcode-form", data)
		}

		// Generate a random id for the table
		id := codegen.GenRandID(config.Shortcodes.Universe, config.Shortcodes.ShortcodeLength)
		// Create a shortcode from the id
		shortcode := codegen.BaseTenToUniverse(id, config.Shortcodes.Universe)

		// Put the link in the database, tag it with the user ID if the user is logged in
		if data.IsLoggedIn {
			sess, err := session.Get("session", c)
			if err != nil {
				data.ShortcodeForm.HasError = true
				data.ShortcodeForm.ErrorText = "Could not get the session"
				return c.Render(http.StatusOK, "shortcode-form", data)
			}
			userId, ok := sess.Values["userId"].(int)
			if !ok {
				data.ShortcodeForm.HasError = true
				data.ShortcodeForm.ErrorText = "Internal Server Error"
				return c.Render(http.StatusOK, "shortcode-form", data)
			}
			err = AddLink(db, id, shortcode, URL, userId)
		} else {
			err = AddLink(db, id, shortcode, URL, -1)
		}

		if err != nil {
			log.Println(err.Error())
			return c.String(http.StatusInternalServerError, "Error adding link to database")
		}

		// Set all of the data for the form to be displayed
		data.ShortcodeForm.Result = shortcode
		data.ShortcodeForm.URL = ""
		data.ShortcodeForm.HasError = false

		return c.Render(http.StatusOK, "shortcode-form", data)
	}

	// This can only be hit if /create is visited without the form being filled out
	// This should not be possible because of client side validataion
	data.ShortcodeForm.URL = URL
	data.ShortcodeForm.Result = ""
	return c.Render(http.StatusOK, "shortcode-form", data)
}

func HandleUserPage(c echo.Context, db *sql.DB, data *globalstructs.UserPageData, config *conf.Config) error {
	data.LinksData = []globalstructs.Link{
		{ID: 1, Shortcode: "abcd", Url: "https://statpage.com"},
		{ID: 1, Shortcode: "abcd", Url: "https://statpage.com"},
		{ID: 1, Shortcode: "abcd", Url: "https://statpage.com"},
		{ID: 1, Shortcode: "abcd", Url: "https://statpage.com"},
		{ID: 1, Shortcode: "abcd", Url: "https://statpage.com"},
		{ID: 1, Shortcode: "abcd", Url: "https://statpage.com"},
	}
	return c.Render(http.StatusOK, "user-homepage", data)
}

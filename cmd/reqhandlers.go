package main

import (
	"database/sql"
	"fmt"

	"github.com/labstack/echo"
	"github.com/vtallen/go-link-shortener/pkg/codegen"
)

func HandleRedirect(c echo.Context, db *sql.DB, config *Config) error {
	shortcode := c.Param("shortcode")

	id := codegen.UniverseToBaseTen(shortcode, config.Shortcodes.Universe)

	var url string
	err := db.QueryRow("SELECT url FROM links WHERE id = ?", id).Scan(&url)
	if err != nil {
		return c.Render(404, "404", nil) // Show the not found page
	}

	return c.Redirect(302, url) // If a url exists, redirect the user to it
}

func HandleAddLink(c echo.Context, db *sql.DB, config *Config, data *IndexData) error {
	url := c.FormValue("url")
	if url != "" {
		fmt.Println(url)
		codegen.GenRandID(config.Shortcodes.Universe, config.Shortcodes.ShortcodeLength)

		data.ShortcodeForm.Result = "abcd"
		// return c.String(200, "shortcode: abcd")
		data.ShortcodeForm.URL = ""
		return c.Render(200, "shortcode-form", data)
		// return c.Render(200, "index", CreateShortcodeData{Shortcode: "test", Title: "This is from /create"})
	}

	// return c.String(200, "url is empty")
	data.ShortcodeForm.URL = url
	data.ShortcodeForm.Result = ""
	return c.Render(200, "shortcode-form", data)
}

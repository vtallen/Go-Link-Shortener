/*
* File cmd/req_handlers.go
*
* Description: This file contains all the request handlers for the web server that are not involved in
*              session management or authentication
*
 */

package main

import (
	"database/sql"
	"net/http"
	"net/url"
	"strconv"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/vtallen/go-link-shortener/internal/conf"
	"github.com/vtallen/go-link-shortener/internal/globalstructs"
	"github.com/vtallen/go-link-shortener/internal/sessmngt"
	"github.com/vtallen/go-link-shortener/pkg/codegen"
)

/*
* Function: HandleRedirect
*
* Parameters: c      echo.Context - The context of the request
*            config *conf.Config - The configuration for the application
*
* Returns: error - If there is an error redirecting the user
*
* Description: This function handles the redirecting of the user to the correct URL based on the shortcode in the url
*
 */
func HandleRedirect(c echo.Context, config *conf.Config) error {
	db, ok := c.Get("db").(*sql.DB)
	if !ok {
		c.Logger().Errorf("Could not get db from context, failed to convert to *sql.DB")
		return c.String(http.StatusInternalServerError, "Internal server error")
	}
	shortcode := c.Param("shortcode")

	// Figure out the id
	id := codegen.UniverseToBaseTen(shortcode, config.Shortcodes.Universe)
	// Query the db for the link
	link, err := GetLink(db, id)
	if err != nil {
		errData := globalstructs.ErrorPageData{ErrorText: "404, link does not exist"}
		return c.Render(http.StatusNotFound, "error-page", errData) // Show the not found page if link does not exist
	}

	// Increment the click counter for the link
	err = IncrementLinkClickCount(db, id)
	if err != nil {
		c.Logger().Errorf("Could not increment click count for link id: %d", id)
	}

	return c.Redirect(http.StatusMovedPermanently, link.Url) // If a url exists, redirect the user to it
}

/*
* Function: HandleAddLink
*
* Parameters: c      echo.Context             - The context of the request
*             config *conf.Config             - The configuration for the application
*             data   *globalstructs.IndexData - The data to pass to the template
*
* Returns: error - If there is an error adding the link to the database
*
* Description: This function handles the adding of a link to the database from a POST request to /create
*
 */
func HandleAddLink(c echo.Context, config *conf.Config, data *globalstructs.IndexData) error {
	db, ok := c.Get("db").(*sql.DB)
	if !ok {
		c.Logger().Errorf("Could not get db from context, failed to convert to *sql.DB")
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

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
			if err != nil {
				c.Logger().Errorf("Could not add link to database: %s", err.Error())
				return c.String(http.StatusInternalServerError, "Error adding link to database")
			}

		} else {
			err = AddLink(db, id, shortcode, URL, -1)
			if err != nil {
				c.Logger().Errorf("Could not add link to database: %s", err.Error())
				return c.String(http.StatusInternalServerError, "Error adding link to database")
			}
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

/*
* Function: HandleDeleteLink
*
* Parameters: c echo.Context - The context of the request
*
* Returns: error - If there is an error deleting the link from the database
*
* Description: This function handles the deletion of a link from the database from a POST request to /delete from
*              the user page
*
 */
func HandleDeleteLink(c echo.Context) error {
	db, ok := c.Get("db").(*sql.DB)
	if !ok {
		c.Logger().Errorf("Could not get db from context, failed to convert to *sql.DB")
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	id, err := strconv.Atoi(c.FormValue("link-id"))
	if err != nil {
		c.Logger().Errorf("Could not convert link-id to int: %s\n", err.Error())
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	err = DeleteLink(db, id)
	if err != nil {
		c.Logger().Errorf("Could not delete link with id: %d from database with error: %s", id, err.Error())
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	// This return of nil will ensure that the calling row from the user page will be deleted from the screen
	return nil
}

/*
* Function: HandleUserPage
*
* Parameters: c    echo.Context                - The context of the request
*             data *globalstructs.UserPageData - The data to pass to the template
*I            config *conf.Config              - The configuration for the application
*
* Returns: error - If there is an error getting the user links from the database
*
* Description: This function handles the rendering of the user page and getting the user links
*
 */
func HandleUserPage(c echo.Context, data *globalstructs.UserPageData, config *conf.Config) error {
	db, ok := c.Get("db").(*sql.DB)
	if !ok {
		c.Logger().Errorf("Could not get db from context, failed to convert to *sql.DB\n")
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	sess, err := session.Get("session", c)
	if err != nil {
		c.Logger().Errorf("Could not get session from context: %s\n", err.Error())
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	userId, ok := sess.Values["userId"].(int)
	if !ok {
		c.Logger().Errorf("Could not convert the session userId to int.\n")
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	data.LinksData, err = GetUserLinks(db, userId)
	if err != nil {
		c.Logger().Error("Could not get user links from database. Error:%s\n ", err.Error())
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	if len(data.LinksData) == 0 {
		data.LinksDataEmpty = true
	} else {
		data.LinksDataEmpty = false
	}

	return c.Render(http.StatusOK, "user-homepage", data)
}

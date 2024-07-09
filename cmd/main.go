/*
* File: cmd/main.go
*
* Description: The main entry point for the application, this file is responsible for setting up the web server, database, and routing
*
 */

package main

import (
	"database/sql"
	"html/template"
	"io"
	"os"
	"strconv"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vtallen/go-link-shortener/internal/conf"
	"github.com/vtallen/go-link-shortener/internal/globalstructs"
	"github.com/vtallen/go-link-shortener/internal/sessmngt"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

/*
* Struct: Templates
*
* Description: This struct is used to store the html templates for the web server that get ingested at startup
*
 */
type Templates struct {
	templates *template.Template
}

/*
* Function: newTemplate
*
* Parameters: None
*
* Returns: *Templates - A pointer to the Templates struct containign all the html templates in the views folder
*
* Description: This function is used to load all the html templates in the views folder into memory
*
 */
func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

/*
* Function: Templates.Render
*
* Parameters: w    io.Writer   - The writer to write the rendered html to
*             name string      - The name of the template to render
*             data interface{} - The data to pass to the template
*
* Returns: error - Any error that occured during the rendering of the template
*
* Description: This function is used to render the html templates to the client
 */
func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

/*
* Function: stringToLogLevel
*
* Parameters: llstring string - The string representation of the log level
*
* Returns: log.Lvl - The log level that corresponds to the string
*
* Description: This function is used to convert a string representation of a log level from the config file to the log.Lvl type
*
 */
func stringToLogLevel(llstring string) log.Lvl {
	switch llstring {
	case "INFO":
		return log.INFO
	case "WARN":
		return log.WARN
	case "ERROR":
		return log.ERROR
	case "DEBUG":
		return log.DEBUG
	default:
		return log.INFO
	}
}

func main() {
	config, err := conf.LoadConfig("config.yaml")
	if err != nil {
		panic("Could not load configuration file config.yaml, Error: " + err.Error())
	}

	// Setup database connection
	db, err := sql.Open("sqlite3", config.Database.Path)
	if err != nil {
		panic("Could not setup database, Error: " + err.Error())
	}
	defer db.Close()

	e := echo.New() // Create the web server

	// Initalize tables in the database
	SetupDB(db, e)

	file, err := os.OpenFile(config.Logging.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("Filed to open log file: " + config.Logging.LogFile + " Error: " + err.Error())
	}

	// Create the multiwriter that will output to the standard out and the log file
	multiWriter := io.MultiWriter(os.Stdout, file)

	e.Logger.SetOutput(multiWriter)
	e.Logger.SetLevel(stringToLogLevel(config.Logging.LogLevel))

	// Setup the logger middleware
	e.Use(middleware.Logger())

	// Setup middleware
	e.Use(middleware.Logger())
	e.Use(dbMiddleware(db)) // Injects the database variable into the request context
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(config.Auth.CookieSecret))))

	e.Static("/images", "images")
	e.Static("/css", "css")

	e.Renderer = newTemplate() // Load the templates

	// Setup data structs for the different pages
	indexData := globalstructs.IndexData{}
	indexData.Server = &config.Server
	errorPageData := globalstructs.ErrorPageData{ErrorText: "No error"}

	// Serve the index page
	e.GET("/", func(c echo.Context) error {
		indexData.ShortcodeForm.URL = ""
		indexData.ShortcodeForm.Result = ""
		indexData.ShortcodeForm.HasError = false
		indexData.HCaptchaSiteKey = config.HCaptcha.SiteKey

		// The navbar changes based on if a user is logged in or not, this enables the functionality
		indexData.IsLoggedIn = false
		err := sessmngt.ValidateSession(c)
		if err == nil {
			indexData.IsLoggedIn = true
		}

		return c.Render(200, "index", indexData)
	})

	// Page to display errors on
	e.GET("/error", func(c echo.Context) error {
		// The navbar changes based on if a user is logged in or not, this enables the functionality
		errorPageData.IsLoggedIn = false
		err := sessmngt.ValidateSession(c)
		if err == nil {
			errorPageData.IsLoggedIn = true
		}

		return c.Render(200, "error-page", errorPageData)
	})

	// Endpoint for the link creation form
	e.POST("/create", func(c echo.Context) error {
		return HandleAddLink(c, db, config, &indexData)
	})

	// Endpoint that handles link deletion from the /user endpoint page
	e.POST("/delete", func(c echo.Context) error {
		return HandleDeleteLink(c)
	}, sessmngt.SessionMiddleware)

	// Endpoint that redirects the user to the stored url if it exists
	e.GET("/:shortcode", func(c echo.Context) error {
		return HandleRedirect(c, db, config)
	})

	loginData := globalstructs.LoginData{} // Data used by login/register pages
	// Endpoint that handles serving the login page
	e.GET("/login", func(c echo.Context) error {
		loginData.HasError = false
		loginData.ErrorText = ""
		loginData.LoginForm.Email = ""
		loginData.HCaptchaSiteKey = config.HCaptcha.SiteKey

		// The navbar changes based on if a user is logged in or not, this enables the functionality
		loginData.IsLoggedIn = false
		err := sessmngt.ValidateSession(c)
		if err == nil {
			loginData.IsLoggedIn = true
		}

		return sessmngt.HandleLoginPage(c, &loginData, config)
	})

	// Endpoint that handles the submission of the login form on /login
	e.POST("/login", func(c echo.Context) error {
		return sessmngt.HandleLoginSession(c, &loginData, config)
	})

	// Endpoint that logs the user out and redirects them to the login page
	e.GET("/logout", func(c echo.Context) error {
		return sessmngt.HandleLogout(c, config)
	})

	// Endpoint that serves the register page
	registerData := globalstructs.RegisterData{}
	e.GET("/register", func(c echo.Context) error {
		registerData.HasError = false
		registerData.ErrorText = ""
		registerData.IsLoggedIn = false
		registerData.Success = false
		registerData.HCaptchaSiteKey = config.HCaptcha.SiteKey

		// The navbar changes based on if a user is logged in or not, this enables the functionality
		registerData.IsLoggedIn = false
		err := sessmngt.ValidateSession(c)
		if err == nil {
			registerData.IsLoggedIn = true
		}

		return sessmngt.HandleRegisterPage(c, &registerData, config)
	})

	// Endpoint that handles the submission of the register form on /register
	e.POST("/register", func(c echo.Context) error {
		return sessmngt.HandleRegisterSession(c, &registerData, config)
	})

	// Endpoint for the user dashboard
	userPageData := globalstructs.UserPageData{}
	e.GET("/user", func(c echo.Context) error {
		userPageData.IsLoggedIn = true // We can assume that this is the case as sessmngt.SessionMiddleware will only allow authenticated users
		return HandleUserPage(c, db, &userPageData, config)
	}, sessmngt.SessionMiddleware)

	e.Logger.Fatal(e.StartTLS(":"+strconv.Itoa(config.Server.Port), config.Auth.TLSCert, config.Auth.TLSKey)) // Run the server
}

package main

import (
	"database/sql"
	"html/template"
	"io"
	"strconv"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vtallen/go-link-shortener/internal/conf"
	"github.com/vtallen/go-link-shortener/internal/pagestructs"
	"github.com/vtallen/go-link-shortener/internal/sessmngt"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
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

	SetupDB(db)

	e := echo.New() // Create the web server

	// Setup middleware
	e.Use(middleware.Logger())
	e.Use(dbMiddleware(db)) // Injects the database variable into the request context
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(config.Auth.CookieSecret))))

	e.Static("/images", "images")
	e.Static("/css", "css")

	e.Renderer = newTemplate() // Load the templates

	// Setup data structs for the different pages
	indexData := pagestructs.IndexData{}
	indexData.Server = &config.Server
	errorPageData := pagestructs.ErrorPageData{ErrorText: "No error"}

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
		indexData.IsLoggedIn = false

		err := sessmngt.ValidateSession(c)
		if err == nil {
			indexData.IsLoggedIn = true
		}
		return HandleAddLink(c, db, config, &indexData)
	})

	// Endpoint that redirects the user to the stored url if it exists
	e.GET("/:shortcode", func(c echo.Context) error {
		return HandleRedirect(c, db, config)
	})

	loginData := pagestructs.LoginData{} // Data used by login/register pages
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
	registerData := pagestructs.RegisterData{}
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
	userPageData := pagestructs.UserPageData{}
	e.GET("/user", func(c echo.Context) error {
		userPageData.IsLoggedIn = true // We can assume that this is the case as sessmngt.SessionMiddleware will only allow authenticated users
		return HandleUserPage(c, db, &userPageData, config)
	}, sessmngt.SessionMiddleware)

	e.Logger.Info(config.HCaptcha.SiteKey)

	e.Logger.Fatal(e.StartTLS(":"+strconv.Itoa(config.Server.Port), config.Auth.TLSCert, config.Auth.TLSKey)) // Run the server
}

package main

import (
	"database/sql"
	"html/template"
	"io"
	"net/http"
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

/*
* Name: IndexData
*
* Description: This struct is used to pass data to the index page.
 */
type IndexData struct {
	ShortcodeForm ShortcodeForm
	Server        *conf.Server
}

type UserPageData struct{}

type ShortcodeForm struct {
	URL      string
	Result   string
	HasError bool
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

	// insert, err := db.Prepare("INSERT INTO links (id, shortcode, url) VALUES (?, ?, ?)")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// var id int = codegen.GenRandID(config.Shortcodes.Universe, config.Shortcodes.ShortcodeLength)
	// var shortcode string = codegen.BaseTenToUniverse(id, config.Shortcodes.Universe)
	// url := "https://www.startpage.com"
	// _, err = insert.Exec(id, shortcode, url)
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }
	// fmt.Println("inserted a link | id: %d | shortcode: %s | url: %s\n", id, shortcode, url)

	// PrintLinksTable(db)
	// PrintUsersTable(db)
	sessmngt.PrintSessionTable(db)

	// Setup middleware
	e.Use(middleware.Logger())
	e.Use(dbMiddleware(db)) // Injects the database variable into the request context
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(config.Auth.CookieSecret))))
	// e.Use(sessmngt.SessionMiddleware)

	// Sends the database into echo.Context so that it can be accessed
	// Setup db middleware

	e.Static("/images", "images")
	e.Static("/css", "css")

	e.Renderer = newTemplate() // Load the templates

	// Setup data structs for the different pages
	indexData := IndexData{}
	indexData.Server = &config.Server
	errorPageData := pagestructs.ErrorPageData{"No error"}

	// Serve the index page
	e.GET("/", func(c echo.Context) error {
		indexData.ShortcodeForm.URL = ""
		indexData.ShortcodeForm.Result = ""
		indexData.ShortcodeForm.HasError = false

		// testdb := c.Get("db").(*sql.DB)
		// var email string
		// err := testdb.QueryRow("SELECT email FROM users WHERE email = ?", "root@root.com").Scan(&email)
		// if err != nil {
		// 	return err
		// } else {
		// 	fmt.Println("\n\n" + email + "\n\n")
		// }
		return c.Render(200, "index", indexData)
	})

	// Page to display errors on
	e.GET("/error", func(c echo.Context) error {
		return c.Render(200, "error-page", errorPageData)
	})

	// Endpoint for the link creation form
	e.POST("/create", func(c echo.Context) error {
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

		return sessmngt.HandleLoginPage(c, &loginData, config)
	})

	// Endpoint that handles the submission of the login form on /login
	e.POST("/login", func(c echo.Context) error {
		return sessmngt.HandleLoginSession(c, db, &loginData, config)
	})

	// Endpoint that logs the user out and redirects them to the login page
	e.GET("/logout", func(c echo.Context) error {
		return sessmngt.HandleLogout(c, config)
	})

	// Endpoint that serves the register page
	registerData := pagestructs.RegisterData{}
	e.GET("/register", func(c echo.Context) error {
		loginData.HasError = false
		loginData.ErrorText = ""
		return sessmngt.HandleRegisterPage(c, &registerData, config)
	})

	// Endpoint that handles the submission of the register form on /register
	e.POST("/register", func(c echo.Context) error {
		return sessmngt.HandleRegisterSession(c, db, &registerData, config)
	})

	// Endpoint for the user dashboard
	userPageData := UserPageData{}
	e.GET("/user", func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			errorPageData.ErrorText = "Error getting session"
			return c.Render(302, "/error", errorPageData)
		}
		// TODO - actually validate sessions
		if sess.Values["userId"] != nil {
			return HandleUserPage(c, db, &userPageData, config)
		} else {
			return c.Redirect(http.StatusMovedPermanently, "/login")
		}
	})

	// testSession := sessmngt.UserSession{SessId: "12", UserId: 1}
	// testSession.StoreExpiryTime(5)
	// for true {
	// 	if testSession.IsValid() {
	// 		fmt.Println("Valid")
	// 	} else {
	// 		fmt.Println("Not valid")
	// 		break
	// 	}
	// }

	e.Logger.Fatal(e.StartTLS(":"+strconv.Itoa(config.Server.Port), config.Auth.TLSCert, config.Auth.TLSKey)) // Run the server
}

package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vtallen/go-link-shortener/pkg/codegen"
	"gopkg.in/yaml.v2"

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
	Server        *Server
}

/*
* Name: LoginData
*
* Description: This struct is used to pass data to the login page.
 */
type LoginData struct {
	LoginForm       LoginForm
	HasError        bool
	ErrorText       string
	AlreadyLoggedIn bool
}

type LoginForm struct {
	Email    string
	Password string
}

type RegisterData struct {
	RegisterForm RegisterForm
	HasError     bool
	ErrorText    string
	Success      bool
}

type RegisterForm struct {
	Email    string
	Password string
}

type UserPageData struct{}

type ErrorPageData struct {
	ErrorText string
}

type ShortcodeForm struct {
	URL      string
	Result   string
	HasError bool
}
type Shortcodes struct {
	ShortcodeLength int    `yaml:"shortcode_length"`
	Universe        string `yaml:"shortcode_universe"`
}

type Auth struct {
	ApiKeyLen        int    `yaml:"api_key_length"`
	RootUsername     string `yaml:"root_username"`
	RootPassword     string `yaml:"root_password"`
	TLSCert          string `yaml:"tls_cert"`
	TLSKey           string `yaml:"tls_key"`
	CookieMaxAgeDays int    `yaml:"cookie_max_age_days"`
	CookieSecret     string `yaml:"cookie_secret"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Config struct {
	Shortcodes Shortcodes
	Auth       Auth
	Server     Server
}

func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	fmt.Println(config)

	return &config, nil
}

func main() {
	// Setup database connection
	db, err := sql.Open("sqlite3", "./shortener.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	SetupDB(db)

	e := echo.New() // Create the web server

	// config := NewConfig()
	config, err := LoadConfig("config.yaml")
	if err != nil {
		e.Logger.Fatal(err.Error())
	}

	// fmt.Println("config.Auth.CookieSecret: ", config.Auth.CookieSecret)

	insert, err := db.Prepare("INSERT INTO links (id, shortcode, url) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	var id int = codegen.GenRandID(config.Shortcodes.Universe, config.Shortcodes.ShortcodeLength)
	var shortcode string = codegen.BaseTenToUniverse(id, config.Shortcodes.Universe)
	url := "https://www.startpage.com"
	_, err = insert.Exec(id, shortcode, url)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("inserted a link | id: %d | shortcode: %s | url: %s\n", id, shortcode, url)

	PrintLinksTable(db)
	PrintUsersTable(db)

	e.Use(middleware.Logger())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(config.Auth.CookieSecret))))

	e.Static("/images", "images")
	e.Static("/css", "css")

	e.Renderer = newTemplate() // Load the templates

	// Setup data structs for the different pages
	indexData := IndexData{}
	indexData.Server = &config.Server
	errorPageData := ErrorPageData{"No error"}

	// Serve the index page
	e.GET("/", func(c echo.Context) error {
		indexData.ShortcodeForm.URL = ""
		indexData.ShortcodeForm.Result = ""
		indexData.ShortcodeForm.HasError = false
		return c.Render(200, "index", indexData)
	})

	e.GET("/error", func(c echo.Context) error {
		return c.Render(200, "error-page", errorPageData)
	})

	// Endpoint for the link creation form
	e.POST("/create", func(c echo.Context) error {
		return HandleAddLink(c, db, config, &indexData)
	})

	loginData := LoginData{} // Data used by login/register pages

	e.GET("/login", func(c echo.Context) error {
		loginData.HasError = false
		loginData.ErrorText = ""
		loginData.LoginForm.Email = ""

		return HandleLoginPage(c, &loginData, config)
	})
	e.POST("/login", func(c echo.Context) error {
		return HandleLoginSession(c, db, &loginData, config)
	})

	e.GET("/logout", func(c echo.Context) error {
		return HandleLogout(c, config)
	})

	registerData := RegisterData{}
	e.GET("/register", func(c echo.Context) error {
		loginData.HasError = false
		loginData.ErrorText = ""
		return HandleRegisterPage(c, &registerData, config)
	})
	e.POST("/register", func(c echo.Context) error {
		return HandleRegisterSession(c, db, &registerData, config)
	})

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

	// Handle use of the redirect function
	e.GET("/:shortcode", func(c echo.Context) error {
		return HandleRedirect(c, db, config)
	})

	e.Logger.Fatal(e.StartTLS(":"+strconv.Itoa(config.Server.Port), config.Auth.TLSCert, config.Auth.TLSKey)) // Run the server
}

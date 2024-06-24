package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/vtallen/go-link-shortener/pkg/codegen"
	"gopkg.in/yaml.v2"

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
	LoginForm LoginForm
	HasError  bool
	ErrorText string
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
	ApiKeyLen    int    `yaml:"api_key_length"`
	RootUsername string `yaml:"root_username"`
	RootPassword string `yaml:"root_password"`
	TLSCert      string `yaml:"tls_cert"`
	TLSKey       string `yaml:"tls_key"`
	CookieSecret string `yaml:"cookie_secret`
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

	// config := NewConfig()
	config, err := LoadConfig("config.yaml")
	if err != nil {
		panic(err)
	}

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

	e := echo.New()
	e.Use(middleware.Logger())
	// e.Use(middleware.CORS())
	// e.Use(middleware.Logger())

	// Sessions setup

	// e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	e.Static("/images", "images")
	e.Static("/css", "css")

	e.Renderer = newTemplate() // Load the templates

	// Setup data structs for the different pages
	indexData := IndexData{}
	indexData.Server = &config.Server

	// Serve the index page
	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", indexData)
	})

	// Handle use of the redirect function
	e.GET("/:shortcode", func(c echo.Context) error {
		return HandleRedirect(c, db, config)
	})

	// Endpoint for the link creation form
	e.POST("/create", func(c echo.Context) error {
		return HandleAddLink(c, db, config, &indexData)
	})

	loginData := LoginData{} // Data used by login/register pages

	e.GET("/login", func(c echo.Context) error {
		return HandleLoginPage(c, &loginData, config)
	})
	e.POST("/login", func(c echo.Context) error {
		return nil
	})

	registerData := RegisterData{}
	e.GET("/register", func(c echo.Context) error {
		return HandleRegisterPage(c, &registerData, config)
	})
	e.POST("/register", func(c echo.Context) error {
		return HandleRegisterSession(c, db, &registerData, config)
	})

	e.Logger.Fatal(e.StartTLS(":"+strconv.Itoa(config.Server.Port), config.Auth.TLSCert, config.Auth.TLSKey)) // Run the server
}

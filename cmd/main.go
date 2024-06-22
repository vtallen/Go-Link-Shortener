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

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
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

type ShortcodeForm struct {
	URL      string
	Result   string
	HasError bool
}

type IndexData struct {
	ShortcodeForm ShortcodeForm
	Server        *Server
	Title         string
	Shortcode     string
}

type Shortcodes struct {
	ShortcodeLength int    `yaml:"shortcode_length"`
	Universe        string `yaml:"shortcode_universe"`
}

type Auth struct {
	ApiKeyLen    int    `yaml:"api_key_length"`
	RootUsername string `yaml:"root_username"`
	RootPassword string `yaml:"root_password"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Config struct {
	MaxChars   int
	ApiKeyLen  int
	Universe   string
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

	e := echo.New()
	e.Use(middleware.Logger())
	data := IndexData{}
	data.Server = &config.Server

	e.Static("/images", "images")
	e.Static("/css", "css")

	e.Renderer = newTemplate() // Load the templates

	// Serve the index page
	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", data)
	})

	// Handle use of the redirect function
	e.GET("/:shortcode", func(c echo.Context) error {
		return HandleRedirect(c, db, config)
	})

	// Endpoint for the link creation form
	e.POST("/create", func(c echo.Context) error {
		// url := c.FormValue("url")
		// if url != "" {
		// 	fmt.Println(url)
		// 	codegen.GenRandID(config.Universe, config.MaxChars)
		// 	fmt.Printf("\n\n\n\n %d", id)
		// 	return c.String(200, "shortcode: abcd")
		// 	// return c.Render(200, "index", CreateShortcodeData{Shortcode: "test", Title: "This is from /create"})
		// }

		// return c.String(200, "url is empty")
		return HandleAddLink(c, db, config, &data)
	})

	e.Logger.Fatal(e.Start(":" + strconv.Itoa(config.Server.Port))) // Run the server
}

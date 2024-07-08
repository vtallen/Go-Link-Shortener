package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/vtallen/go-link-shortener/internal/globalstructs"

	"github.com/labstack/echo/v4"
	"github.com/vtallen/go-link-shortener/internal/sessmngt"
	"github.com/vtallen/go-link-shortener/pkg/codegen"
)

type APIKey struct {
	key  string
	user string
}

func SetupDB(db *sql.DB) {
	// Create the links table if it doesn't exist
	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS links (id INTEGER PRIMARY KEY, shortcode TEXT, url TEXT, userId INTEGER, clicks INTEGER DEFAULT 0)")
	if err != nil {
		log.Fatal(err)
		panic("DB setup failed, table links")
	}
	statement.Exec()

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS apikeys (key TEXT, user TEXT)")
	if err != nil {
		log.Fatal(err)
		panic("DB setup failed, table apikeys")
	}
	statement.Exec()

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT NOT NULL, username TEXT NOT NULL, password TEXT NOT NULL, permissions TEXT NOT NULL)")
	if err != nil {
		log.Fatal(err)
		panic("DB setup failed, table users")
	}
	statement.Exec()

	// sessId given as TEXT as a gorilla sessions session id is a string
	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS sessions (sessId INTEGER PRIMARY KEY, expiryTimeUnix INTEGER NOT NULL, userId INTEGER NOT NULL)")
	if err != nil {
		log.Fatal(err)
		panic("DB setup failed, table sessions")
	}
	statement.Exec()
}

func dbMiddleware(db *sql.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", db)
			return next(c)
		}
	}
}

func GenUniqueID(db *sql.DB, universe string, maxchars int) (int, error) {
	maxiters := 10000
	idx := 0
	for {
		if idx > maxiters {
			return 0, errors.New("genUniqueID timeout reached, no unique id found")
		}
		// Create a random id
		id := codegen.GenRandID(universe, maxchars)
		// Check if that id already exists
		_, err := GetLink(db, id)
		// If no link with this id is found, err will not be nil, meaning this id is safe
		if err != nil {
			return id, nil
		}
	}

	// return 0, errors.New("genUniqueID this should not be possible to reach")
}

func GenShortcode(universe string, maxchars int) (string, int) {
	id := codegen.GenRandID(universe, maxchars)
	return codegen.BaseTenToUniverse(id, universe), id
}

func GetLink(db *sql.DB, id int) (*globalstructs.Link, error) {
	var link globalstructs.Link
	err := db.QueryRow("SELECT id, url, userId, clicks FROM links WHERE id = ?", id).Scan(&link.ID, &link.Url, &link.UserId, &link.Clicks)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func AddLink(db *sql.DB, id int, shortcode string, url string, userId int) error {
	insert, err := db.Prepare("INSERT INTO links (id, shortcode, url, userId) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = insert.Exec(id, shortcode, url, userId)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}

	return nil
}

func DeleteLink(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM links WHERE id = ?", id)
	return err
}

func IncrementLinkClickCount(db *sql.DB, linkId int) error {
	statement, err := db.Prepare("UPDATE links SET clicks = clicks + 1 WHERE id = ? ")
	if err != nil {
		return err
	}

	_, err = statement.Exec(linkId)
	if err != nil {
		return err
	}

	return nil
}

func GetLinkClicks(db *sql.DB, linkId int) (int, error) {
	var clicks int
	err := db.QueryRow("SELECT clicks FROM links WHERE id = ?", linkId).Scan(&clicks)
	if err != nil {
		return 0, err
	}

	return clicks, nil
}

func GetUserLinks(db *sql.DB, userId int) ([]globalstructs.Link, error) {
	var links []globalstructs.Link

	rows, err := db.Query("SELECT id, shortcode, url, clicks FROM links WHERE userId = ?", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var link globalstructs.Link
		err = rows.Scan(&link.ID, &link.Shortcode, &link.Url, &link.Clicks)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	return links, nil
}

func GetAllUsers(db *sql.DB) []sessmngt.UserLogin {
	rows, err := db.Query("SELECT email, username, password, permissions FROM users")
	if err != nil {
		log.Fatal(err.Error())
	}

	var users []sessmngt.UserLogin
	for rows.Next() {
		var user sessmngt.UserLogin
		if err := rows.Scan(&user.Email, &user.Username, &user.Password, &user.Permissions); err != nil {
			log.Fatal(err.Error())
		}

		users = append(users, user)
	}

	return users
}

func GetAllLinks(db *sql.DB) []globalstructs.Link {
	rows, err := db.Query("SELECT id, shortcode, url, userId, clicks FROM links")
	if err != nil {
		log.Fatal(err.Error())
	}

	var links []globalstructs.Link
	for rows.Next() {
		var link globalstructs.Link
		if err := rows.Scan(&link.ID, &link.Shortcode, &link.Url, &link.UserId, &link.Clicks); err != nil {
			log.Fatal(err.Error())
		}

		links = append(links, link)
	}

	return links
}

func PrintLinksTable(db *sql.DB) {
	var links []globalstructs.Link = GetAllLinks(db)

	for idx := 0; idx < len(links); idx++ {
		fmt.Printf("id: %d | shortcode: %s | url: %s | userId: %d | clicks: %d\n", links[idx].ID, links[idx].Shortcode, links[idx].Url, links[idx].UserId, links[idx].Clicks)
	}
}

func PrintUsersTable(db *sql.DB) {
	var users []sessmngt.UserLogin = GetAllUsers(db)

	for idx := 0; idx < len(users); idx++ {
		fmt.Printf("email: %s | username: %s | password: %s | permissions: %s\n", users[idx].Email, users[idx].Username, users[idx].Password, users[idx].Permissions)
	}
}

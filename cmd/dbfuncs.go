package main

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"

	"github.com/vtallen/go-link-shortener/internal/sessmngt"
	"github.com/vtallen/go-link-shortener/pkg/codegen"
)

type Link struct {
	ID        int
	Shortcode string
	Url       string
}

type APIKey struct {
	key  string
	user string
}

func SetupDB(db *sql.DB) {
	fmt.Println("\n\nSetting up database\n\n")
	// Create the links table if it doesn't exist
	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS links (id INTEGER PRIMARY KEY, shortcode TEXT, url TEXT)")
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
}

func GenUniqueID(db *sql.DB, universe string, maxchars int) (int, error) {
	maxiters := 10000
	idx := 0
	for true {
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

	return 0, errors.New("genUniqueID this should not be possible to reach")
}

func GenShortcode(universe string, maxchars int) (string, int) {
	id := codegen.GenRandID(universe, maxchars)
	return codegen.BaseTenToUniverse(id, universe), id
}

func GetLink(db *sql.DB, id int) (*Link, error) {
	var link Link
	err := db.QueryRow("SELECT id, url FROM links WHERE id = ?", id).Scan(&link.ID, &link.Url)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func AddLink(db *sql.DB, id int, shortcode string, url string) error {
	insert, err := db.Prepare("INSERT INTO links (id, shortcode, url) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = insert.Exec(id, shortcode, url)
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

func GetAllLinks(db *sql.DB) []Link {
	rows, err := db.Query("SELECT id, shortcode, url FROM links")
	if err != nil {
		log.Fatal(err.Error())
	}

	var links []Link
	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.ID, &link.Shortcode, &link.Url); err != nil {
			log.Fatal(err.Error())
		}

		links = append(links, link)
	}

	return links
}

func GetAPIKeys(db *sql.DB) []APIKey {
	rows, err := db.Query("SELECT key, user FROM apikeys")
	if err != nil {
		log.Fatal(err.Error())
	}

	var keys []APIKey
	for rows.Next() {
		var key APIKey
		if err := rows.Scan(&key.key, &key.user); err != nil {
			log.Fatal(err.Error())
		}

		keys = append(keys, key)
	}

	return keys
}

func PrintLinksTable(db *sql.DB) {
	var links []Link = GetAllLinks(db)

	for idx := 0; idx < len(links); idx++ {
		fmt.Printf("id: %d | shortcode: %s | url: %s\n", links[idx].ID, links[idx].Shortcode, links[idx].Url)
	}
}

func PrintUsersTable(db *sql.DB) {
	var users []sessmngt.UserLogin = GetAllUsers(db)

	for idx := 0; idx < len(users); idx++ {
		fmt.Printf("email: %s | username: %s | password: %s | permissions: %s\n", users[idx].Email, users[idx].Username, users[idx].Password, users[idx].Permissions)
	}
}

func PrintAPIKeys(db *sql.DB) {
	var apikeys []APIKey = GetAPIKeys(db)

	for idx := 0; idx < len(apikeys); idx++ {
		fmt.Printf("key: %s | user: %s\n", apikeys[idx].key, apikeys[idx].user)
	}
}

// GenerateAPIKey generates a random API key with the specified length
func GenerateAPIKey(length int) (string, error) {
	// Calculate the byte size needed to create the API key
	byteSize := length / 2 // We use hex encoding, so each byte corresponds to 2 characters

	// Create a byte slice to hold the random bytes
	apiKeyBytes := make([]byte, byteSize)

	// Generate random bytes
	_, err := rand.Read(apiKeyBytes)
	if err != nil {
		return "", err
	}

	// Convert the random bytes to a hexadecimal string
	apiKey := hex.EncodeToString(apiKeyBytes)

	return apiKey, nil
}

func AddAPIKey(db *sql.DB, key string, user string, permissions string) error {
	insert, err := db.Prepare("INSERT INTO apikeys (key, user, permissions) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = insert.Exec(key, user, permissions)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}

	return nil
}

func DeleteAPIKey(db *sql.DB, key string) error {
	_, err := db.Exec("DELETE FROM apikeys WHERE key = ?", key)
	return err
}

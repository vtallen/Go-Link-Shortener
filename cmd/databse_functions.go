/*
* File: cmd/database_functions.go
*
* Description: This file contains all the functions used to interact with the database as it pertains to
*              the links and users tables. Anything to do with sessions is handled in the sessmngt package.
 */

package main

import (
	"database/sql"
	"errors"
	"log"

	"github.com/vtallen/go-link-shortener/internal/globalstructs"

	"github.com/labstack/echo/v4"
	"github.com/vtallen/go-link-shortener/internal/sessmngt"
	"github.com/vtallen/go-link-shortener/pkg/codegen"
)

/*
* Funmction: SetupDB
*
* Parameters: db *sql.DB    - A pointer to the database object
*             e  *echo.Echo - A pointer to the echo object for logging
*
* Returns: None (an error that happens here results in the program closing)
*
* Description: This function is used to setup the database by creating the links, users, and sessions tables
*              if they dont already exist
*
 */
func SetupDB(db *sql.DB, e *echo.Echo) {
	// Create the links table if it doesn't exist
	statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS links (id INTEGER PRIMARY KEY, shortcode TEXT, url TEXT, userId INTEGER, clicks INTEGER DEFAULT 0)")
	if err != nil {
		e.Logger.Fatalf("DB setup failed on table links. Error: %s", err.Error())
	}
	_, err = statement.Exec()
	if err != nil {
		e.Logger.Fatalf("DB setup failed on table links. Error: %s", err.Error())
	}

	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT NOT NULL, username TEXT NOT NULL, password TEXT NOT NULL, permissions TEXT NOT NULL)")
	if err != nil {
		e.Logger.Fatalf("DB setup failed on table users. Error: %s", err.Error())
	}
	_, err = statement.Exec()
	if err != nil {
		e.Logger.Fatalf("DB setup failed on table users. Error: %s", err.Error())
	}

	// sessId given as TEXT as a gorilla sessions session id is a string
	statement, err = db.Prepare("CREATE TABLE IF NOT EXISTS sessions (sessId INTEGER PRIMARY KEY, expiryTimeUnix INTEGER NOT NULL, userId INTEGER NOT NULL)")
	if err != nil {
		e.Logger.Fatalf("DB setup failed on table sessions. Error: %s", err.Error())
	}
	_, err = statement.Exec()
	if err != nil {
		e.Logger.Fatalf("DB setup failed on table sessions. Error: %s", err.Error())
	}
}

/*
* Function: dbMiddleware
*
* Parameters: db *sql.DB - A pointer to the database object
*
* Returns: echo.MiddlewareFunc - A middleware function that sets the database object in the echo context
*
* Description: This function is used to create a middleware function that sets the database object in the context
*              of requests. This allows handlers to access the database object without having to pass it in as a parameter
*
 */
func dbMiddleware(db *sql.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", db)
			return next(c)
		}
	}
}

/*
* Function: GenUniqueID
*
* Parameters: db       *sql.DB - A pointer to the database object
*            universe  string  - The universe of characters to use when generating the id
*            maxchars  int     - The maximum number of characters the id can be
*
* Returns: int   - The unique id that was generated
*          error - Any error that occurred during the generation of the id
*
* Description: This function is used to generate a unique id for a link. It generates a random id and checks if
*           it already exists in the database. If it does, it generates another id and checks again. This process
*          is repeated until a unique id is found or a timeout is reached, in which case an error is returned
*
 */
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
}

/*
* Function: GenShortcode
*
* Parameters: universe  string - The universe of characters to use when generating the shortcode
*            maxchars  int    - The maximum number of characters the shortcode can be
*
* Returns: string - The shortcode that was generated
*
* Description: This function is used to generate a shortcode for a link. It generates a random id and converts it to
*              the base b representation of that number where b is the length of universe
*
 */
func GenShortcode(universe string, maxchars int) (string, int) {
	id := codegen.GenRandID(universe, maxchars)
	return codegen.BaseTenToUniverse(id, universe), id
}

/*
* Function: GetLink
*
* Parameters:  db *sql.DB - A pointer to the database object
*              id int     - The id of the link to get
*
* Returns: *globalstructs.Link - A pointer to the link that was retrieved
*          error               - Any error that occurred during the retrieval of the link
*
* Description:
 */
func GetLink(db *sql.DB, id int) (*globalstructs.Link, error) {
	var link globalstructs.Link
	err := db.QueryRow("SELECT id, url, userId, clicks FROM links WHERE id = ?", id).Scan(&link.ID, &link.Url, &link.UserId, &link.Clicks)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

/*
* Function: AddLink
*
* Parameters: db        *sql.DB - A pointer to the database object
*             id        int     - The id of the link to add
*             shortcode string  - The shortcode of the link to add
*             url       string  - The url of the link should redirect to
*             userId    int     - The id of the user that created the link
*
* Returns: error - Any error that occurred during the insertion of the link
*
* Description: This function is used to add a link to the links table in the database
*
 */
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

/*
* Function: DeleteLink
*
* Parameters: db *sql.DB - A pointer to the database object
*             id int    - The id of the link to delete
*
* Returns: error - Any error that occurred during the deletion of the link
*
* Description: This function is used to delete a link from the links table in the database
*              based on the id of the link
*
 */
func DeleteLink(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM links WHERE id = ?", id)
	return err
}

/*
* Function: IncrementLinkClickCount
*
* Parameters: db     *sql.DB - A pointer to the database object
*             linkId int     - The id of the link to increment
*
* Returns: error - Any error that occurred during the increment of the link click count
*
* Description: This function is used to increment the click count of a link in the database
 */
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

/*
* Function: GetLinkClicks
*
* Parameters: db     *sql.DB - A pointer to the database object
*             linkId int     - The id of the link to get the click count for
*
* Returns: int   - The number of clicks the link has
*          error - Any error that occurred during the retrieval of the click count
*
* Description: This function is used to get the number of clicks a link has
*
 */
func GetLinkClicks(db *sql.DB, linkId int) (int, error) {
	var clicks int
	err := db.QueryRow("SELECT clicks FROM links WHERE id = ?", linkId).Scan(&clicks)
	if err != nil {
		return 0, err
	}

	return clicks, nil
}

/*
* Function: GetUserLinks
*
* Parameters: db     *sql.DB - A pointer to the database object
*             userId int     - The id of the user to get the
*
* Returns: []globalstructs.Link - A slice of links that the user has created
*          error                - Any error that occurred during the retrieval
*
* Description: This function is used to get all the links that a user has created
 */
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

/*
* Function: GetAllUsers
*
* Parameters: db *sql.DB - A pointer to the database object
*
* Returns: []sessmngt.UserLogin - A slice of all the users in the database
*
* Description: This function is used to get all the users in the users table in the database
*           and return them as a slice. Used mostly for debugging
 */
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

/*
* Function: GetAllLinks
*
* Parameters: db *sql.DB - A pointer to the database object
*
* Returns: []globalstructs.Link - A slice of all the links in the database
*
* Description: This function is used to get all the links in the links table in the database
 */
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

/*
* Function: PrintLinksTable
*
* Parameters: db *sql.DB    - A pointer to the database object
*             e  *echo.Echo - A pointer to the echo object for logging
*
* Returns:
*
* Description:
 */
func PrintLinksTable(db *sql.DB, e *echo.Echo) {
	var links []globalstructs.Link = GetAllLinks(db)

	for idx := 0; idx < len(links); idx++ {
		e.Logger.Debugf("id: %d | shortcode: %s | url: %s | userId: %d | clicks: %d\n", links[idx].ID, links[idx].Shortcode, links[idx].Url, links[idx].UserId, links[idx].Clicks)
	}
}

/*
* Function: PrintUsersTable
*
* Parameters: db *sql.DB - A pointer to the database object
*             e  *echo.Echo - A pointer to the echo object for logging
*
* Returns: None
*
* Description: This function is used to print all the users in the users table to the log
*
 */
func PrintUsersTable(db *sql.DB, e *echo.Echo) {
	var users []sessmngt.UserLogin = GetAllUsers(db)

	for idx := 0; idx < len(users); idx++ {
		e.Logger.Debugf("email: %s | username: %s | password: %s | permissions: %s\n", users[idx].Email, users[idx].Username, users[idx].Password, users[idx].Permissions)
	}
}

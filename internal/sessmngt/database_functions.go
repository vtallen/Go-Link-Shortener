/*
* File: internal/sessmngt/database_functions.go
*
* Description: This file contains functions related to session management and authentication that need to use the database
 */
package sessmngt

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

/*
=======================================================
* Session struct functions
=======================================================
*/

/*
* Struct: UserSession
*
* Description: This struct represents a user session in the database, and is used throughout the program to represent a session.
 */
type UserSession struct {
	SessId         int64 // Unique id of the session
	ExpiryTimeUnix int64 // The unix time at which the session is no longer valid
	UserId         int   // Unique id of the user
}

/*
* Function: StoreExpiryTime
*
* Parameters: maxAgeSeconds int64 - How long, in seconds, a session should be allowed to last
*
* Description: Acts on a user struct, the passed in value will be used to calculate the Unix time
*              at which the sessions should expire. The value is then stored into the struct's
*              ExpiryTimeUnix field
*
 */
func (usr *UserSession) StoreExpiryTime(maxAgeSeconds int64) {
	// Get the current unix time
	unixTime := time.Now().Unix()

	// Add the maxAgeSeconds to the current time
	expiryTime := unixTime + maxAgeSeconds

	// Store it in the expiry field
	usr.ExpiryTimeUnix = expiryTime
}

/*
* Function: IsValid
*
* Parameters: None
*
* Returns: bool - true if the current unix time is less than the stored unix time
*
* Description: Used to see if the UserSession struct contains a valid expiry time
*
 */
func (usr *UserSession) IsValid() bool {
	unixTime := time.Now().Unix()

	return unixTime < usr.ExpiryTimeUnix
}

/*
* Function: StoreDB
*
* Parameters: db *sql.DB - The database to store the struct into
*
* Returns: error
*
* Description: Takes the values stored in this struct and inserts them into the database. It will error if
*              it already exists in the database
*
 */
func (usr *UserSession) StoreDB(db *sql.DB) error {
	statement, err := db.Prepare("INSERT INTO sessions (sessId, expiryTimeUnix, userId) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}

	_, err = statement.Exec(usr.SessId, usr.ExpiryTimeUnix, usr.UserId)
	if err != nil {
		return err
	}

	return nil
}

/*
* Function: DeleteDB
*
* Parameters: db *sql.DB - The database to store the struct into
*
* Returns: error
*
* Description: Deletes this UserSession based on the sessID field
*
 */

func (usr *UserSession) DeleteDB(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM sessions where sessId = ?", usr.SessId)
	return err
}

/*
=======================================================
* Session functions
=======================================================
*/

/*
* Function: GetSessionStruct
*
* Parameters: db *sql.DB - The database to store the struct into
*             sessID int64 - The id of the session to retrive from the database
*
* Returns: *UserSession, error - The session with sessID
*
* Description: Retrives the session with sessID from the database and returns it as a UserSession struct
*
 */

func GetSessionStruct(db *sql.DB, sessId int64) (*UserSession, error) {
	var session *UserSession = &UserSession{}
	err := db.QueryRow("SELECT sessID, expiryTimeUnix, userId FROM sessions WHERE sessID = ?", sessId).Scan(&session.SessId, &session.ExpiryTimeUnix, &session.UserId)
	if err != nil {
		return nil, err
	}

	return session, nil
}

/*
* Function: GetAllSessions
*
* Parameters: db *sql.DB - The database to store the struct into
*
* Returns: []UserSession - A slice of all sessions currently in the database
*
* Description: Returns a slice containing all user sessions in the database
*
 */
func GetAllSessions(db *sql.DB) []UserSession {
	rows, err := db.Query("SELECT sessID, expiryTimeUnix, userId FROM sessions")
	if err != nil {
		log.Fatal(err.Error())
	}

	var allSessions []UserSession
	for rows.Next() {
		var usrSess UserSession
		if err := rows.Scan(&usrSess.SessId, &usrSess.ExpiryTimeUnix, &usrSess.UserId); err != nil {
			log.Fatal(err.Error())
		}

		allSessions = append(allSessions, usrSess)
	}

	return allSessions
}

/*
* Function: PrintSessionTable
*
* Parameters: db *sql.DB - The database to store the struct into
*
* Returns: None
*
* Description: Prints a formatted table of all sessions in the database to the standard out
*
 */
func PrintSessionTable(db *sql.DB) {
	allSessions := GetAllSessions(db)
	fmt.Println(len(allSessions))
	for idx := 0; idx < len(allSessions); idx++ {
		fmt.Printf("sessID: %d | expiryTimeUnix: %d | usrId: %d\n", allSessions[idx].SessId, allSessions[idx].ExpiryTimeUnix, allSessions[idx].UserId)
	}
}

/*=======================================================
* User struct functions
=======================================================*/

/*
* Struct: User
*
* Fields: Id int - The id of the user
*         Email string - The email of the user
*         Username string - The username of the user (not unique, gets populated by email)
*        Password string - The hashed password of the user
*        Permissions string - The access level of the user
*
* Description: This struct represents a user in the database, and is used throughout
*              the program to represent a user.
 */
type UserLogin struct {
	Id          int
	Email       string
	Username    string
	Password    string
	Permissions string
}

/*
* Function: AddUser
*
* Parameters: db *sql.DB - The database to add the user to
*             email string - The email of the user to add
*             username string - The username of the user to add
*             password string - The password of the user to add
*             permissions string - The access level of the user to add
*
* Description: This function adds a user to the database with the specified email, username,
 */
func AddUser(db *sql.DB, email string, username string, password string, permissions string) error {
	insert, err := db.Prepare("INSERT INTO users (email, username, password, permissions) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = insert.Exec(email, username, password, permissions)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}

	return nil
}

/*
* Name: RemoveUser
*
* Parameters: db *sql.DB - The database to remove the user from
*             email string - The email of the user to remove
*
* Description: This function removes a user from the database with the specified email.
*
 */
func RemoveUser(db *sql.DB, email string) error {
	_, err := db.Exec("DELETE FROM users WHERE email = ?", email)
	return err
}

/*
* Name: GetUserByEmail
*
* Parameters: db *sql.DB - The database to get the user from
*             email string - The email of the user to get
*
* Description: This function retrieves a user from the database with the specified email.
*
* Returns: *User - If the user is found, the user is returned. If the user is not found, nil is returned.
*          error - If there is an error retrieving the user, the error is returned.
 */
func GetUserByEmail(db *sql.DB, email string) (*UserLogin, error) {
	var user UserLogin = UserLogin{}

	err := db.QueryRow("SELECT id, email, username, password, permissions FROM users WHERE email = ?", email).Scan(&user.Id, &user.Email, &user.Username, &user.Password, &user.Permissions)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

/*
* Name: GetUserById
*
* Parameters: db *sql.DB - The application database
*             id int - The id of the user to get
*
* Description: This function retrieves a user from the database with the specified id.
*
* Returns: *UserLogin - If the user is found, the user is returned. If the user is not found, nil is returned.
*          error - If there is an error retrieving the user, the error is returned.
*
 */
func GetUserById(db *sql.DB, id int) (*UserLogin, error) {
	var user UserLogin

	err := db.QueryRow("SELECT id, email, username, password, permissions FROM users WHERE id = ?", id).Scan(&user.Id, &user.Email, &user.Username, &user.Password, &user.Permissions)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

/*
* Name: GetUserId
*
* Parameters: db *sql.DB - The application database
*             email string - The email of the user to get the id of
*
* Description: This function retrieves the id of a user from the database with the specified email.
 */
func GetUserId(db *sql.DB, email string) (int, error) {
	var id int
	err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

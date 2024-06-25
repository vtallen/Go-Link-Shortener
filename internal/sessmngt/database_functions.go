package sessmngt

import (
	"database/sql"
	"log"
)

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

	// user.Id = id
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

package main

import (
	"database/sql"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type UserLogin struct {
	Id          int
	Email       string
	Username    string
	Password    string
	Permissions string
}

/*
* Function: HashPassword
*
* Parameters: password string - The password to hash
*
* Description: This function takes a password string and hashes it using bcrypt.
* The hashed password is then stored in the User struct.
*
* Returns: error - If there is an error hashing the password, the error is returned.
 */
func (u *UserLogin) HashPassword(password string) error {
	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	// if err != nil {
	// 	return err
	// }

	// u.Password = string(hashedPassword)
	// return nil
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	u.Password = hashedPassword
	return nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

/*
* Function: CheckPassword
*
* Parameters: hash string - The hashed password to compare to
*             password string - The password to compare
*
* Description: This function takes a hashed password and a password and compares them.
*
* Returns: error - If the passwords do not match, the error is returned.
 */
func CheckPassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
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

func GetUser(db *sql.DB, id int) (*UserLogin, error) {
	var user UserLogin

	err := db.QueryRow("SELECT email, username, password, permissions FROM users WHERE id = ?", id).Scan(&user.Email, &user.Username, &user.Password, &user.Permissions)
	if err != nil {
		return nil, err
	}

	user.Id = id
	return &user, nil
}

func GetUserId(db *sql.DB, email string) (int, error) {
	var id int
	err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

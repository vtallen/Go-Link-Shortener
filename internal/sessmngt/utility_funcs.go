package sessmngt

import "golang.org/x/crypto/bcrypt"

/*
* Function: HashPassword
*
* Parameters: password string - The password to hash
*
* Description: This function takes a password and hashes it.
*
* Returns: string - The hashed password
*          error - If there is an error hashing the password, the error is returned.
 */

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

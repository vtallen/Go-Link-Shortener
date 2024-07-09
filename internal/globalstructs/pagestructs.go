/*
* File: internal/globalstructs/pagestructs.go
*
* Description: This file contains all the structs used to pass data to the html templates as well as
*            the structs used to hold data for the web server.
 */

package globalstructs

import (
	"github.com/vtallen/go-link-shortener/internal/conf"
)

/*
* Struct: IndexData
*
* Description: This struct is used to pass data to the index page.
 */
type IndexData struct {
	ShortcodeForm   ShortcodeForm // Contains information for the re-filling of the form upon unseccessful completion
	Server          *conf.Server  // Contains config information about the hostname of the server for the generated shortcodes
	HCaptchaSiteKey string        // Used to enable the use of hCaptcha

	IsLoggedIn bool // Used by the navbar to change what appears based on if a user is logged in
}

/*
* Struct: UserPageData
*
* Description: This struct is used to pass data to the user page.
*
 */
type UserPageData struct {
	LinksData  []Link // All of the user's links that they have created
	IsLoggedIn bool   // Used by the navbar to change what appears based on if a user is logged in.
	// This should always be true for this route as the session middleware is called by route /user
	LinksDataEmpty bool // Used to determine if the user has any links to display
}

/*
* Struct: ShortcodeForm
*
* Description: This struct is used to pass data to the shortcode form on the index page.
*              This struct is used to re-fill the form with the previous data if the form was submitted with errors.
*
 */
type ShortcodeForm struct {
	URL       string // The url that the user wants to shorten
	Result    string // The result of the shortcode generation
	HasError  bool   // If the form was submitted with errors
	ErrorText string // The error text to display if the form was submitted with errors
}

/*
* Struct: ErrorPageData
*
* Description: This struct is used to pass data to the error page.
*
 */
type ErrorPageData struct {
	ErrorText  string // The error text to display
	IsLoggedIn bool   // Used by the navbar to change what appears based on if a user is logged in
}

/*
* Struct: LoginData
*
* Description: This struct is used to pass data to the login page.
*
 */
type LoginData struct {
	LoginForm       LoginForm // Contains the data from the login form
	HasError        bool      // If the form was submitted with errors
	HCaptchaSiteKey string    // Used to enable the use of hCaptcha
	ErrorText       string    // The error text to display if the form was submitted with errors
	IsLoggedIn      bool      // Used by the navbar to change what appears based on if a user is logged in
}

/*
* Struct: LoginForm
*
* Description: This struct is used to pass data to the login form on the login page
*              This struct is used to re-fill the form with the previous data if the form was submitted with errors.
 */
type LoginForm struct {
	Email    string
	Password string
}

/*
* Struct: RegisterData
*
* Description: This struct is used to pass data to the register page.
*
 */
type RegisterData struct {
	RegisterForm    RegisterForm // Contains the data from the register form
	HasError        bool         // If the form was submitted with errors
	ErrorText       string       // The error text to display if the form was submitted with errors
	HCaptchaSiteKey string       // Used to enable the use of hCaptcha
	Success         bool         // true if the registration was successful
	IsLoggedIn      bool         // Used by the navbar to change what appears based on if a user is logged in
}

/*
* Struct: RegisterForm
*
* Description: This struct is used to pass data to the register form on the register page
*              This struct is used to re-fill the form with the previous data if the form was submitted with errors.
*
 */
type RegisterForm struct {
	Email    string // The email that the user wants to register
	Password string // The password that the user wants to use
}

/*
* Struct: Link
*
* Description: Used to represent a link in the database
 */
type Link struct {
	ID        int    // The id of the link in the database
	Shortcode string // The shortcode used to access this link. Is a base b representation of ID
	Url       string // The url that the shortcode redirects to
	UserId    int    // The id of the user that created this link. -1 if the link was created by an unauthenticated user
	Clicks    int    // The number of times the link has been clicked
}

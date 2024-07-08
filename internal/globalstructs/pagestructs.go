package globalstructs

import (
	"github.com/vtallen/go-link-shortener/internal/conf"
)

/*
* Name: IndexData
*
* Description: This struct is used to pass data to the index page.
 */
type IndexData struct {
	ShortcodeForm   ShortcodeForm // Contains information for the re-filling of the form upon unseccessful completion
	Server          *conf.Server  // Contains config information about the hostname of the server for the generated shortcodes
	HCaptchaSiteKey string        // Used to enable the use of hCaptcha

	IsLoggedIn bool // Used by the navbar to change what appears based on if a user is logged in
}

type UserPageData struct {
	LinksData      []Link
	IsLoggedIn     bool
	LinksDataEmpty bool
}

type ShortcodeForm struct {
	URL       string
	Result    string
	HasError  bool
	ErrorText string
}
type ErrorPageData struct {
	ErrorText  string
	IsLoggedIn bool // Used by the navbar to change what appears based on if a user is logged in
}

/*
* Name: LoginData
*
* Description: This struct is used to pass data to the login page.
 */
type LoginData struct {
	LoginForm       LoginForm
	HasError        bool
	HCaptchaSiteKey string
	ErrorText       string
	IsLoggedIn      bool
}

type LoginForm struct {
	Email    string
	Password string
}

type RegisterData struct {
	RegisterForm    RegisterForm
	HasError        bool
	ErrorText       string
	HCaptchaSiteKey string
	Success         bool
	IsLoggedIn      bool
}

type RegisterForm struct {
	Email    string
	Password string
}

type Link struct {
	ID        int
	Shortcode string
	Url       string
	UserId    int
	Clicks    int
}

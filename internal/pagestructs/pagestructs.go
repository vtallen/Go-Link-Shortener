package pagestructs

import "github.com/vtallen/go-link-shortener/internal/conf"

/*
* Name: IndexData
*
* Description: This struct is used to pass data to the index page.
 */
type IndexData struct {
	ShortcodeForm   ShortcodeForm
	Server          *conf.Server
	HCaptchaSiteKey string
}

type UserPageData struct{}

type ShortcodeForm struct {
	URL      string
	Result   string
	HasError bool
}
type ErrorPageData struct {
	ErrorText string
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

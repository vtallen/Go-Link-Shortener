package pagestructs

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
	ErrorText       string
	AlreadyLoggedIn bool
}

type LoginForm struct {
	Email    string
	Password string
}

type RegisterData struct {
	RegisterForm RegisterForm
	HasError     bool
	ErrorText    string
	Success      bool
}

type RegisterForm struct {
	Email    string
	Password string
}

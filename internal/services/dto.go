package services

type RegisterInput struct {
	Username  string
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type AuthTokens struct {
	Access  string `json:"token"`
	Refresh string `json:"-"`
}

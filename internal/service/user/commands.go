package user

type RegisterCommand struct {
	Token    string
	Login    string
	Password string
}

type AuthCommand struct {
	Login    string
	Password string
}

type LogoutCommand struct {
	Token string
}

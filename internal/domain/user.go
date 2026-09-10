package domain

import "uuid"

type User struct {
	id       uuid.UUID
	login    Login
	password Password
	token    Token
}

func NewUser(login Login, password Password) *User {
	return &User{
		id:       uuid.New(),
		login:    login,
		password: password,
		token:    Token{},
	}
}

func (u *User) SetToken(token Token) {
	u.token = token
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Login() Login {
	return u.login
}

func (u *User) Password() Password {
	return u.password
}

func (u *User) Token() Token {
	return u.token
}

func RestoreUser(id uuid.UUID, login string, password string, token string) *User {
	return &User{
		id:       id,
		login:    Login{value: login},
		password: Password{encoded: password},
		token:    Token{value: token},
	}
}

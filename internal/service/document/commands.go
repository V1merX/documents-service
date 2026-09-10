package document

import (
	"github.com/V1merX/documents-service/internal/domain"
)

type ListCommand struct {
	Requester domain.Login
	Target    *string
	Key       *string
	Value     *string
	Limit     int
}

type CreateCommand struct {
	Login   domain.Login
	Name    string
	Mime    string
	File    bool
	Public  bool
	JSON    []byte
	Content []byte
	Grant   []string
}

type GetCommand struct {
	Login domain.Login
	ID    string
}

type DeleteCommand struct {
	Login domain.Login
	ID    string
}

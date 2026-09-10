package convertor

import (
	"encoding/json"

	"github.com/V1merX/documents-service/internal/domain"
	"github.com/V1merX/documents-service/internal/service/document"
	"github.com/V1merX/documents-service/internal/service/user"
	"github.com/V1merX/documents-service/internal/transport/http/request"
	"github.com/V1merX/documents-service/internal/transport/http/response"
)

func ToDocumentResponse(doc domain.Document) *response.Doc {
	return &response.Doc{
		ID:      doc.ID().String(),
		Name:    doc.Name(),
		Mime:    doc.Mime(),
		File:    doc.File(),
		Public:  doc.Public(),
		Created: doc.Created(),
		Grant:   doc.Grant(),
	}
}

func ToDocumentsResponse(docs *[]domain.Document) *response.ListDocsResponse {
	resp := make([]response.Doc, len(*docs))
	for i, d := range *docs {
		resp[i] = *ToDocumentResponse(d)
	}
	return &response.ListDocsResponse{Docs: resp}
}

func ToAuthResponse(token string) *response.AuthResponse {
	return &response.AuthResponse{
		Token: token,
	}
}

func ToRegisterResponse(login string) *response.RegisterResponse {
	return &response.RegisterResponse{
		Login: login,
	}
}

func ToLogoutResponse(id string) map[string]bool {
	return map[string]bool{id: true}
}

func ToDocDeleteResponse(deletedID string) map[string]bool {
	return map[string]bool{deletedID: true}
}

func ToAuthCommand(req request.AuthRequest) user.AuthCommand {
	return user.AuthCommand{
		Login:    req.Login,
		Password: req.Password,
	}
}

func ToLogoutCommand(token string) user.LogoutCommand {
	return user.LogoutCommand{
		Token: token,
	}
}

func ToRegisterCommand(req request.RegisterRequest) user.RegisterCommand {
	return user.RegisterCommand{
		Token:    req.Token,
		Login:    req.Login,
		Password: req.Password,
	}
}

func ToCreateDocsResponse(fileName string, jsonData []byte) *response.CreateDocsResponse {
	resp := &response.CreateDocsResponse{File: fileName}
	if len(jsonData) > 0 {
		resp.JSON = json.RawMessage(jsonData)
	}

	return resp
}

func ToDocCreateCommand(meta request.DocumentMeta, login domain.Login, jsonData, content []byte) document.CreateCommand {
	return document.CreateCommand{
		Login:   login,
		Name:    meta.Name,
		Mime:    meta.Mime,
		File:    meta.File,
		Public:  meta.Public,
		JSON:    jsonData,
		Content: content,
		Grant:   meta.Grant,
	}
}

func ToDocDeleteCommand(login domain.Login, id string) document.DeleteCommand {
	return document.DeleteCommand{
		Login: login,
		ID:    id,
	}
}

func ToDocGetCommand(login domain.Login, id string) document.GetCommand {
	return document.GetCommand{
		Login: login,
		ID:    id,
	}
}

func ToDocListCommand(requester domain.Login, limit int, key, value, login string) document.ListCommand {
	cmd := document.ListCommand{
		Requester: requester,
		Limit:     limit,
	}

	if login != "" {
		cmd.Target = &login
	}

	if key != "" {
		cmd.Key = &key
	}

	if value != "" {
		cmd.Value = &value
	}

	return cmd
}

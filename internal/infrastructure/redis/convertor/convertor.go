package convertor

import (
	"github.com/V1merX/documents-service/internal/domain"
	"github.com/V1merX/documents-service/internal/infrastructure/redis/model"
)

func DocToDomain(doc model.Document) *domain.Document {
	return domain.RestoreDocument(
		doc.ID,
		doc.Owner,
		doc.Name,
		doc.Mime,
		doc.File,
		doc.Public,
		doc.JSON,
		doc.Content,
		doc.Created,
		doc.Grant,
	)
}

func DocsToDomain(docs []model.Document) []domain.Document {
	result := make([]domain.Document, 0, len(docs))
	for _, d := range docs {
		result = append(result, *DocToDomain(d))
	}

	return result
}

func DocsToModel(docs []domain.Document) []model.Document {
	result := make([]model.Document, 0, len(docs))
	for i := range docs {
		result = append(result, DocToModel(&docs[i]))
	}

	return result
}

func DocToModel(doc *domain.Document) model.Document {
	grant := doc.Grant()
	logins := make([]string, len(grant))
	for i, g := range grant {
		logins[i] = g.Value()
	}

	return model.Document{
		ID:      doc.ID(),
		Owner:   doc.Owner().Value(),
		Name:    doc.Name(),
		Mime:    doc.Mime(),
		File:    doc.File(),
		Public:  doc.Public(),
		JSON:    doc.JSON(),
		Content: doc.Content(),
		Created: doc.Created(),
		Grant:   logins,
	}
}

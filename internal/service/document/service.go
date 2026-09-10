package document

import (
	"context"
	"uuid"

	"github.com/V1merX/documents-service/internal/domain"
	"go.uber.org/zap"
)

type repo interface {
	Create(ctx context.Context, doc *domain.Document) error
	Get(ctx context.Context, id uuid.UUID) (*domain.Document, error)
	List(ctx context.Context, requester, target domain.Login, key, value *string, limit int) (*[]domain.Document, error)
	Delete(ctx context.Context, owner domain.Login, docID uuid.UUID) error
}

type docCache interface {
	Get(ctx context.Context, id uuid.UUID) (*domain.Document, bool, error)
	Put(ctx context.Context, doc *domain.Document) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type listCache interface {
	Get(ctx context.Context, key ListKey) (*[]domain.Document, bool, error)
	Put(ctx context.Context, key ListKey, docs *[]domain.Document) error
	Invalidate(ctx context.Context, logins ...domain.Login) error
}

type ListKey struct {
	Requester domain.Login
	Target    domain.Login
	Key       *string
	Value     *string
	Limit     int
}

type Service struct {
	log       *zap.Logger
	repo      repo
	docCache  docCache
	listCache listCache
}

func NewService(log *zap.Logger, repo repo, docCache docCache, listCache listCache) *Service {
	return &Service{log: log, repo: repo, docCache: docCache, listCache: listCache}
}

func (s *Service) Create(ctx context.Context, cmd CreateCommand) (*domain.Document, error) {
	grant := make([]domain.Login, 0, len(cmd.Grant)+1)
	for _, g := range cmd.Grant {
		login, err := domain.NewLogin(g)
		if err != nil {
			return nil, err
		}
		grant = append(grant, login)
	}
	grant = append(grant, cmd.Login)

	doc, err := domain.NewDocument(cmd.Login, cmd.Name, cmd.Mime, cmd.File, cmd.Public, cmd.JSON, cmd.Content, grant)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, err
	}

	if err := s.listCache.Invalidate(ctx, grant...); err != nil {
		s.log.Error("Failed to invalidate lists", zap.Error(err))
	}

	return doc, nil
}

func (s *Service) List(ctx context.Context, cmd ListCommand) (*[]domain.Document, error) {
	var (
		target domain.Login
		err    error
	)
	if cmd.Target != nil {
		target, err = domain.NewLogin(*cmd.Target)
		if err != nil {
			return nil, err
		}
	}

	key := ListKey{cmd.Requester, target, cmd.Key, cmd.Value, cmd.Limit}

	docs, found, err := s.listCache.Get(ctx, key)
	if err != nil {
		s.log.Error("Failed to get list from cache", zap.Error(err))
	}
	if err == nil && found {
		return docs, nil
	}

	docs, err = s.repo.List(ctx, cmd.Requester, target, cmd.Key, cmd.Value, cmd.Limit)
	if err != nil {
		return nil, err
	}

	if err := s.listCache.Put(ctx, key, docs); err != nil {
		s.log.Error("Failed to put list in cache", zap.Error(err))
	}

	return docs, nil
}

func (s *Service) Get(ctx context.Context, cmd GetCommand) (*domain.Document, error) {
	docID, err := uuid.Parse(cmd.ID)
	if err != nil {
		return nil, domain.ErrInvalidDocumentID
	}

	if docID == uuid.Nil() {
		return nil, domain.ErrInvalidDocumentID
	}

	var (
		doc   *domain.Document
		found bool
	)

	doc, found, err = s.docCache.Get(ctx, docID)
	if err != nil {
		s.log.Error("Failed to get document from cache", zap.Error(err), zap.String("id", docID.String()))
	}
	if err != nil || !found {
		doc, err = s.repo.Get(ctx, docID)
		if err != nil {
			return nil, err
		}

		if err := s.docCache.Put(ctx, doc); err != nil {
			s.log.Error("Failed to put document in cache", zap.Error(err))
		}
	}

	if !doc.HasAccess(cmd.Login) {
		return nil, domain.ErrForbidden
	}

	return doc, nil
}

func (s *Service) Delete(ctx context.Context, cmd DeleteCommand) (string, error) {
	docID, err := uuid.Parse(cmd.ID)
	if err != nil {
		return "", domain.ErrInvalidDocumentID
	}

	if docID == uuid.Nil() {
		return "", domain.ErrInvalidDocumentID
	}

	doc, err := s.repo.Get(ctx, docID)
	if err != nil {
		return "", err
	}

	if doc.Owner() != cmd.Login {
		return "", domain.ErrForbidden
	}

	if err := s.repo.Delete(ctx, cmd.Login, docID); err != nil {
		return "", err
	}

	if err := s.docCache.Delete(ctx, docID); err != nil {
		s.log.Error("Failed to delete document from cache", zap.Error(err), zap.String("id", docID.String()))
	}

	if err := s.listCache.Invalidate(ctx, doc.Grant()...); err != nil {
		s.log.Error("Failed to invalidate lists", zap.Error(err))
	}

	return docID.String(), nil
}

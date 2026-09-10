package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"uuid"

	"github.com/V1merX/documents-service/internal/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var listFilterColumns = map[string]string{
	"id":      "id::text",
	"name":    "name",
	"mime":    "mime",
	"file":    "file::text",
	"public":  "public::text",
	"created": "created_at::text",
}

type fileStorage interface {
	Save(id uuid.UUID, content []byte) (string, error)
	Read(rel string) ([]byte, error)
	Remove(rel string) error
}

type DocumentRepo struct {
	pool  *pgxpool.Pool
	files fileStorage
	log   *zap.Logger
}

func NewDocumentRepo(pool *pgxpool.Pool, files fileStorage, log *zap.Logger) *DocumentRepo {
	return &DocumentRepo{pool: pool, files: files, log: log}
}

func (r *DocumentRepo) Create(ctx context.Context, doc *domain.Document) error {
	var contentPath *string
	if doc.File() {
		rel, err := r.files.Save(doc.ID(), doc.Content())
		if err != nil {
			return fmt.Errorf("store document file: %w", err)
		}
		contentPath = &rel
	}

	var jsonData *string
	if len(doc.JSON()) > 0 {
		jsonData = new(string(doc.JSON()))
	}

	grant := doc.Grant()
	sharedWith := make([]string, len(grant))
	for i, g := range grant {
		sharedWith[i] = g.Value()
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO documents
		     (id, owner, name, mime, file, public, json_data, content_path, created_at, shared_with)
		 VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10)`,
		doc.ID(), doc.Owner().Value(), doc.Name(), doc.Mime(), doc.File(), doc.Public(),
		jsonData, contentPath, doc.Created(), sharedWith)
	if err != nil {
		r.removeOrphan(contentPath, doc.ID())

		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return domain.ErrBadRequest
			case pgerrcode.ForeignKeyViolation:
				return domain.ErrUserNotFound
			}
		}

		return fmt.Errorf("create document: %w", err)
	}

	return nil
}

func (r *DocumentRepo) Get(ctx context.Context, id uuid.UUID) (*domain.Document, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT owner, name, mime, file, public, json_data::text, content_path, created_at, shared_with
		 FROM documents
		 WHERE id = $1`, id)

	var (
		owner, name, mime     string
		file, public          bool
		jsonData, contentPath *string
		created               time.Time
		sharedWith            []string
	)

	if err := row.Scan(&owner, &name, &mime, &file, &public,
		&jsonData, &contentPath, &created, &sharedWith); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDocumentNotFound
		}

		return nil, fmt.Errorf("get document: %w", err)
	}

	var content []byte
	if contentPath != nil {
		var err error
		if content, err = r.files.Read(*contentPath); err != nil {
			return nil, fmt.Errorf("read document file: %w", err)
		}
	}

	return domain.RestoreDocument(id, owner, name, mime, file, public,
		rawJSON(jsonData), content, created, sharedWith), nil
}

func (r *DocumentRepo) List(
	ctx context.Context,
	requester, target domain.Login,
	key, value *string,
	limit int,
) (*[]domain.Document, error) {
	var query strings.Builder

	query.WriteString(
		`SELECT id, owner, name, mime, file, public, json_data::text, created_at, shared_with
		 FROM documents
		 WHERE (owner = $1 OR public = true OR $1 = ANY (shared_with))`)

	args := []any{requester.Value()}

	if target.Value() == "" {
		query.WriteString(" AND owner = $1")
	} else {
		args = append(args, target.Value())
		fmt.Fprintf(&query, " AND owner = $%d", len(args))
	}

	if key != nil && value != nil {
		column, ok := listFilterColumns[*key]
		if !ok {
			return nil, domain.ErrInvalidFilterKey
		}

		args = append(args, *value)
		fmt.Fprintf(&query, " AND %s = $%d", column, len(args))
	}

	args = append(args, limit)
	fmt.Fprintf(&query, " ORDER BY name, created_at LIMIT $%d", len(args))

	rows, err := r.pool.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()

	docs := make([]domain.Document, 0, limit)
	for rows.Next() {
		var (
			id           uuid.UUID
			owner, name  string
			mime         string
			file, public bool
			jsonData     *string
			created      time.Time
			sharedWith   []string
		)

		if err := rows.Scan(&id, &owner, &name, &mime, &file, &public,
			&jsonData, &created, &sharedWith); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}

		docs = append(docs, *domain.RestoreDocument(id, owner, name, mime,
			file, public, rawJSON(jsonData), nil, created, sharedWith))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}

	return &docs, nil
}

func (r *DocumentRepo) Delete(ctx context.Context, owner domain.Login, docID uuid.UUID) error {
	var contentPath *string

	err := r.pool.QueryRow(ctx,
		`DELETE FROM documents
		 WHERE id = $1 AND owner = $2
		 RETURNING content_path`, docID, owner.Value()).Scan(&contentPath)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrDocumentNotFound
		}

		return fmt.Errorf("delete document: %w", err)
	}

	r.removeOrphan(contentPath, docID)

	return nil
}

func (r *DocumentRepo) removeOrphan(contentPath *string, docID uuid.UUID) {
	if contentPath == nil {
		return
	}

	if err := r.files.Remove(*contentPath); err != nil {
		r.log.Error("Failed to remove document file",
			zap.Error(err),
			zap.String("id", docID.String()),
			zap.String("path", *contentPath),
		)
	}
}

func rawJSON(raw *string) []byte {
	if raw == nil {
		return nil
	}

	return []byte(*raw)
}

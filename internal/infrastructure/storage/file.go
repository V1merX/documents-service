package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"uuid"
)

type FileStorage struct {
	root string
}

func New(root string) (*FileStorage, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve storage root: %w", err)
	}

	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, fmt.Errorf("create storage root: %w", err)
	}

	return &FileStorage{root: abs}, nil
}

func (s *FileStorage) Save(id uuid.UUID, content []byte) (string, error) {
	name := id.String()
	rel := filepath.Join(name[:2], name)
	abs := s.resolve(rel)

	if err := os.MkdirAll(filepath.Dir(abs), 0o750); err != nil {
		return "", fmt.Errorf("create storage dir: %w", err)
	}

	tmp := abs + ".tmp"
	if err := os.WriteFile(tmp, content, 0o640); err != nil {
		return "", fmt.Errorf("write temp file: %w", err)
	}

	if err := os.Rename(tmp, abs); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("commit file: %w", err)
	}

	return rel, nil
}

func (s *FileStorage) Read(rel string) ([]byte, error) {
	content, err := os.ReadFile(s.resolve(rel))
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return content, nil
}

func (s *FileStorage) Remove(rel string) error {
	if err := os.Remove(s.resolve(rel)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove file: %w", err)
	}

	return nil
}

func (s *FileStorage) resolve(rel string) string {
	return filepath.Join(s.root, filepath.Join("/", rel))
}

package storage

import (
	"os"
	"path/filepath"

	"github.com/tanmaydeobhankar/nebulafs/internal/files"
)

// DiskStore implements Store using the local filesystem
type DiskStore struct {
	BaseDir string
}

// NewDiskStore creates a new DiskStore
func NewDiskStore(baseDir string) (*DiskStore, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	return &DiskStore{BaseDir: baseDir}, nil
}

func (s *DiskStore) WriteChunk(chunk files.Chunk) error {
	path := filepath.Join(s.BaseDir, chunk.Hash)
	return os.WriteFile(path, chunk.Content, 0644)
}

func (s *DiskStore) ReadChunk(hash string) (files.Chunk, error) {
	path := filepath.Join(s.BaseDir, hash)
	content, err := os.ReadFile(path)
	if err != nil {
		return files.Chunk{}, err
	}

	return files.Chunk{
		Hash:    hash,
		Content: content,
		Size:    len(content),
	}, nil
}

func (s *DiskStore) HasChunk(hash string) bool {
	path := filepath.Join(s.BaseDir, hash)
	_, err := os.Stat(path)
	return err == nil
}

package storage

import (
	"github.com/tanmaydeobhankar/nebulafs/internal/files"
)

// an interface to store chunks locally
type Store interface {
	WriteChunk(chunk files.Chunk) error

	ReadChunk(hash string) (files.Chunk, error)

	HasChunk(hash string) bool
}

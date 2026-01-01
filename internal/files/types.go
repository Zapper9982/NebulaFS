package files

import (
	"crypto/sha1"
	"encoding/hex"
)

const ChunkSize = 1024 * 1024 // 1MB

type Chunk struct {
	Index   int    `json:"index"`
	Size    int    `json:"size"`
	Hash    string `json:"hash"` // sha1 hash
	Content []byte `json:"content"`
}

// metadata represents the structure of a file in the system
type FileMetadata struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Size      int64   `json:"size"`
	Type      string  `json:"type"`
	Chunks    []Chunk `json:"chunks"`
	Encrypted bool    `json:"encrypted"`
}

// hash calc using SHA1
func CalculateHash(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

package files

import (
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/tanmaydeobhankar/nebulafs/internal/crypto"
)

// ProcessFile splits a file into encrypted chunks and returns metadata
func ProcessFile(path string) (FileMetadata, []Chunk, []byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return FileMetadata{}, nil, nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return FileMetadata{}, nil, nil, err
	}

	// Generate encryption key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return FileMetadata{}, nil, nil, err
	}

	var chunks []Chunk
	buffer := make([]byte, ChunkSize)
	index := 0

	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return FileMetadata{}, nil, nil, err
		}
		if n == 0 {
			break
		}

		data := buffer[:n]

		// Encrypt
		encryptedData, err := crypto.EncryptAES256(data, key)
		if err != nil {
			return FileMetadata{}, nil, nil, err
		}

		// Calculate Hash of Encrypted Data (this is the key for storage)
		hash := crypto.HashSHA1(encryptedData)

		chunk := Chunk{
			Index:   index,
			Size:    len(encryptedData),
			Hash:    hash,
			Content: encryptedData,
		}
		chunks = append(chunks, chunk)
		index++
	}

	// Calculate ID for the file itself (hash of the name + size + time?)
	// Or just a random ID. Let's use a random ID for now or hash of the first chunk hash?
	// Let's use SHA1 of the file name + size for determinism in this simple example,
	// typically you'd want a random UUID.
	fileID := crypto.HashSHA1([]byte(filepath.Base(path) + time.Now().String()))

	metadata := FileMetadata{
		ID:        fileID,
		Name:      filepath.Base(path),
		Size:      stat.Size(),
		Type:      filepath.Ext(path),
		Chunks:    chunks, // Note: In a real system, we might only store Hash references here to save RAM
		Encrypted: true,
	}

	return metadata, chunks, key, nil
}

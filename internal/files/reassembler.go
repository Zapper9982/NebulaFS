package files

import (
	"errors"
	"sort"

	"github.com/tanmaydeobhankar/nebulafs/internal/crypto"
)

// ReassembleFile reconstructs a file from chunks
func ReassembleFile(chunks []Chunk, key []byte) ([]byte, error) {
	// Sort chunks by index
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Index < chunks[j].Index
	})

	var assembledData []byte

	for _, chunk := range chunks {
		// Verify Hash?
		if crypto.HashSHA1(chunk.Content) != chunk.Hash {
			return nil, errors.New("chunk hash mismatch - data corruption")
		}

		// Decrypt
		decrypted, err := crypto.DecryptAES256(chunk.Content, key)
		if err != nil {
			return nil, err
		}

		assembledData = append(assembledData, decrypted...)
	}

	return assembledData, nil
}

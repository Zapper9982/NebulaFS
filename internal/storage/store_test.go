package storage

import (
	"os"
	"testing"

	"github.com/tanmaydeobhankar/nebulafs/internal/files"
)

func TestDiskStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nebulafs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewDiskStore(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	chunk := files.Chunk{
		Hash:    "test-hash",
		Content: []byte("test-content"),
		Size:    12,
	}

	// Test Write
	if err := store.WriteChunk(chunk); err != nil {
		t.Fatalf("Failed to write chunk: %v", err)
	}

	// Test HasChunk
	if !store.HasChunk("test-hash") {
		t.Errorf("Expected chunk to exist")
	}

	// Test Read
	readChunk, err := store.ReadChunk("test-hash")
	if err != nil {
		t.Fatalf("Failed to read chunk: %v", err)
	}

	if string(readChunk.Content) != "test-content" {
		t.Errorf("Content mismatch. Got %s", readChunk.Content)
	}
}

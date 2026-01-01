package files

import (
	"bytes"
	"os"
	"testing"
)

func TestProcessAndReassemble(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "nebulafs_test_file")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	originalContent := []byte("This is a test file for NebulaFS chunking and encryption.")
	if _, err := tmpFile.Write(originalContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Process
	meta, chunks, key, err := ProcessFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ProcessFile failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("Expected chunks, got 0")
	}

	if !meta.Encrypted {
		t.Error("Metadata should indicate encryption")
	}

	// Reassemble
	reassembled, err := ReassembleFile(chunks, key)
	if err != nil {
		t.Fatalf("ReassembleFile failed: %v", err)
	}

	if !bytes.Equal(originalContent, reassembled) {
		t.Errorf("Content mismatch. Expected '%s', got '%s'", originalContent, reassembled)
	}
}

package node

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNodeFileUploadDownload(t *testing.T) {
	// Setup Temp Dirs
	tmpDir, _ := os.MkdirTemp("", "nebulafs_node_test")
	defer os.RemoveAll(tmpDir)

	storageDir := filepath.Join(tmpDir, "storage")
	os.Mkdir(storageDir, 0755)

	// Create Node
	conf := NodeConfig{
		Port:       5000,
		StorageDir: storageDir,
	}
	n, err := NewNode(conf)
	if err != nil {
		t.Fatal(err)
	}

	// Create Input File
	inputFile := filepath.Join(tmpDir, "input.txt")
	content := []byte("Hello Distributed World!")
	os.WriteFile(inputFile, content, 0644)

	// Test Upload
	meta, keyHex, err := n.UploadFile(inputFile)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// Check if chunks exist in storage
	for _, c := range meta.Chunks {
		if !n.Store.HasChunk(c.Hash) {
			t.Errorf("Chunk %s missing from storage", c.Hash)
		}
	}

	// Test Download
	outputFile := filepath.Join(tmpDir, "output.txt")
	err = n.DownloadFile(meta, keyHex, outputFile)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Verify Content
	readContent, _ := os.ReadFile(outputFile)
	if string(readContent) != string(content) {
		t.Errorf("Content mismatch")
	}
}

func TestDistributedTransfer(t *testing.T) {
	// Setup Temp Root
	tmpDir, _ := os.MkdirTemp("", "nebulafs_dist_test")
	defer os.RemoveAll(tmpDir)

	// Node 1 (Bootstrap / Storage Node)
	store1 := filepath.Join(tmpDir, "store1")
	node1, err := NewNode(NodeConfig{Port: 6001, StorageDir: store1})
	if err != nil {
		t.Fatal(err)
	}
	go node1.Start()
	time.Sleep(500 * time.Millisecond) // Let it start

	// Node 2 (Uploader)
	store2 := filepath.Join(tmpDir, "store2")
	node2, err := NewNode(NodeConfig{Port: 6002, StorageDir: store2, BootstrapPeers: []string{"127.0.0.1:6001"}})
	if err != nil {
		t.Fatal(err)
	}
	go node2.Start()
	time.Sleep(500 * time.Millisecond) // Bootstrap

	// Create Data on Node 2
	inputFile := filepath.Join(tmpDir, "secret.txt")
	content := []byte("Top Secret Data distributed across the galaxy")
	os.WriteFile(inputFile, content, 0644)

	// Upload from Node 2 (Should replicate to Node 1 via DHT closest logic)
	// Since hashes are random, it might NOT always pick Node 1 if there were many nodes.
	// But with 2 nodes, Node 1 is definitely in the "closest 3".
	meta, keyHex, err := node2.UploadFile(inputFile)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	time.Sleep(1 * time.Second) // Allow async replication to finish

	// Verify Node 1 has the chunk
	chunkHash := meta.Chunks[0].Hash
	if !node1.Store.HasChunk(chunkHash) {
		// Debug: check if node1 got anything
		t.Errorf("Node 1 did not receive the replicated chunk %s", chunkHash)
	}

	// Node 3 (Downloader - Empty Store)
	store3 := filepath.Join(tmpDir, "store3")
	node3, err := NewNode(NodeConfig{Port: 6003, StorageDir: store3, BootstrapPeers: []string{"127.0.0.1:6001"}})
	if err != nil {
		t.Fatal(err)
	}
	go node3.Start()
	time.Sleep(500 * time.Millisecond)

	// Download on Node 3 (Should fetch from Node 1 or Node 2)
	outputFile := filepath.Join(tmpDir, "retrieved.txt")
	err = node3.DownloadFile(meta, keyHex, outputFile)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Calculate checksum
	retrievedContent, _ := os.ReadFile(outputFile)
	if string(retrievedContent) != string(content) {
		t.Errorf("Content mismatch on Node 3")
	}
}

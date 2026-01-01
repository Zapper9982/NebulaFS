package node

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/tanmaydeobhankar/nebulafs/internal/dht"
	"github.com/tanmaydeobhankar/nebulafs/internal/files"
	"github.com/tanmaydeobhankar/nebulafs/internal/p2p"
)

// UploadFile processes a file and stores its chunks
func (n *Node) UploadFile(path string) (files.FileMetadata, string, error) {
	fmt.Printf("Processing file: %s\n", path)

	// 1. Chunk and Encrypt
	metadata, chunks, key, err := files.ProcessFile(path)
	if err != nil {
		return files.FileMetadata{}, "", err
	}

	fmt.Printf("File split into %d chunks. ID: %s\n", len(chunks), metadata.ID)

	// 2. Distribute Chunks
	for _, chunk := range chunks {
		// A. Store Locally (Always)
		if err := n.Store.WriteChunk(chunk); err != nil {
			fmt.Printf("Error writing chunk locally: %v\n", err)
			continue
		}

		// B. Publish to Network (DHT)
		// Find closest nodes to the chunk hash
		chunkID := dht.NewID(chunk.Hash)
		contacts := n.DHT.RoutingTable.FindClosestContacts(chunkID, 3)

		// Create Payload
		payload, _ := json.Marshal(chunk)
		msg := p2p.Message{
			Type:    p2p.MsgStoreChunk,
			Sender:  n.DHT.ID.Hex(),
			Payload: payload,
		}

		fmt.Printf("Replicating chunk %s to %d peers...\n", chunk.Hash[:8], len(contacts))
		for _, contact := range contacts {
			if contact.Address == n.Transport.(*p2p.WebSocketTransport).Address {
				continue // Don't send to self
			}
			go n.Transport.SendMessage(contact.Address, msg)
		}
	}

	return metadata, fmt.Sprintf("%x", key), nil
}

// DownloadFile retrieves chunks and reconstructs the file
func (n *Node) DownloadFile(metadata files.FileMetadata, keyHex string, outputPath string) error {
	fmt.Printf("Downloading file: %s (ID: %s)\n", metadata.Name, metadata.ID)

	key, err := hexDecode(keyHex)
	if err != nil {
		return fmt.Errorf("invalid key: %v", err)
	}

	var gatheredChunks []files.Chunk

	for _, chunkMeta := range metadata.Chunks {
		// 1. Check Local
		chunk, err := n.Store.ReadChunk(chunkMeta.Hash)
		if err == nil {
			chunk.Index = chunkMeta.Index
			gatheredChunks = append(gatheredChunks, chunk)
			continue
		}

		// 2. If not local, Ask Network
		fmt.Printf("Chunk %s missing locally. Requesting from network...\n", chunkMeta.Hash[:8])

		// Find probable providers (closest peers to chunk Hash)
		chunkID := dht.NewID(chunkMeta.Hash)
		contacts := n.DHT.RoutingTable.FindClosestContacts(chunkID, 5)

		reqPayload, _ := json.Marshal(p2p.ChunkRequestPayload{Hash: chunkMeta.Hash})
		reqMsg := p2p.Message{
			Type:    p2p.MsgRequestChunk,
			Sender:  n.DHT.ID.Hex(),
			Payload: reqPayload,
		}

		// Send Request to all
		for _, contact := range contacts {
			go n.Transport.SendMessage(contact.Address, reqMsg)
		}

		// Wait for chunk to appear in storage (Handling incoming MsgStoreChunk saves to disk)
		// Polling for now is simple
		found := false
		for i := 0; i < 10; i++ {
			time.Sleep(500 * time.Millisecond)
			if n.Store.HasChunk(chunkMeta.Hash) {
				chunk, _ := n.Store.ReadChunk(chunkMeta.Hash)
				chunk.Index = chunkMeta.Index
				gatheredChunks = append(gatheredChunks, chunk)
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("Failed to retrieve chunk %s\n", chunkMeta.Hash[:8])
			return fmt.Errorf("chunk %s missing", chunkMeta.Hash)
		}
	}

	// 3. Reassemble
	data, err := files.ReassembleFile(gatheredChunks, key)
	if err != nil {
		return err
	}

	// 4. Save
	return os.WriteFile(outputPath, data, 0644)
}

func hexDecode(s string) ([]byte, error) {
	var data []byte
	_, err := fmt.Sscanf(s, "%x", &data)
	return data, err
}

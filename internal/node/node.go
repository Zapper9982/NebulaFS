package node

import (
	"encoding/json"
	"fmt"

	"github.com/tanmaydeobhankar/nebulafs/internal/dht"
	"github.com/tanmaydeobhankar/nebulafs/internal/files"
	"github.com/tanmaydeobhankar/nebulafs/internal/p2p"
	"github.com/tanmaydeobhankar/nebulafs/internal/storage"
)

// Node represents the full NebulaFS node
type Node struct {
	DHT       *dht.DHT
	Store     storage.Store
	Transport p2p.Transport
	Config    NodeConfig
}

type NodeConfig struct {
	Port           int
	BootstrapPeers []string
	StorageDir     string
}

func NewNode(config NodeConfig) (*Node, error) {
	store, err := storage.NewDiskStore(config.StorageDir)
	if err != nil {
		return nil, err
	}

	address := fmt.Sprintf("127.0.0.1:%d", config.Port) // Using IP for consistent dial
	transport := p2p.NewWebSocketTransport(address)

	id := dht.NewID(address)
	dhtNode := dht.NewDHT(id, address)

	n := &Node{
		DHT:       dhtNode,
		Store:     store,
		Transport: transport,
		Config:    config,
	}

	n.registerHandlers(transport)
	return n, nil
}

func (n *Node) Start() error {
	go func() {
		fmt.Printf("Node %s listening on %d\n", n.DHT.ID.Hex()[:8], n.Config.Port) // Transport address might be just :port
		// We need to pass the actual listen address
		if err := n.Transport.Listen(fmt.Sprintf(":%d", n.Config.Port)); err != nil {
			fmt.Println("Transport error:", err)
		}
	}()

	// Connect to bootstrap peers
	if len(n.Config.BootstrapPeers) > 0 {
		fmt.Printf("Bootstrapping to %v...\n", n.Config.BootstrapPeers)
		for _, peerAddr := range n.Config.BootstrapPeers {
			// In real Kademlia:
			// 1. Ping bootstrap
			// 2. FindNode(Self) to find neighbors

			// Just Send Ping for now
			msg := p2p.Message{
				Type:   p2p.MsgDHTPing,
				Sender: n.DHT.ID.Hex(),
			}
			if err := n.Transport.SendMessage(peerAddr, msg); err != nil {
				fmt.Printf("Failed to bootstrap to %s: %v\n", peerAddr, err)
			}
		}
	}

	select {}
}

func (n *Node) registerHandlers(t *p2p.WebSocketTransport) {
	// ... handlers ...
	// Since Transport interface doesn't expose RegisterHandler (it was on concrete type),
	// we passed *p2p.WebSocketTransport here.

	// DHT PING
	t.RegisterHandler(p2p.MsgDHTPing, func(p *p2p.Peer, msg p2p.Message) {
		// Update table
		senderID := dht.NewID(msg.Sender)
		contact := dht.Contact{ID: senderID, Address: p.Address}
		n.DHT.AddNode(contact)
		fmt.Printf("[%d] Received PING from %s\n", n.Config.Port, msg.Sender[:8])

		// Reply with PONG
		pong := p2p.Message{
			Type:   p2p.MsgDHTPong,
			Sender: n.DHT.ID.Hex(),
		}
		n.Transport.SendMessage(p.Address, pong)
	})

	// DHT PONG
	t.RegisterHandler(p2p.MsgDHTPong, func(p *p2p.Peer, msg p2p.Message) {
		senderID := dht.NewID(msg.Sender)
		contact := dht.Contact{ID: senderID, Address: p.Address}
		n.DHT.AddNode(contact)
		fmt.Printf("[%d] Received PONG from %s\n", n.Config.Port, msg.Sender[:8])
	})

	// STORE CHUNK (Replica)
	t.RegisterHandler(p2p.MsgStoreChunk, func(p *p2p.Peer, msg p2p.Message) {
		var chunk files.Chunk
		if err := json.Unmarshal(msg.Payload, &chunk); err != nil {
			return
		}
		fmt.Printf("[%d] Received Chunk %s for replication\n", n.Config.Port, chunk.Hash[:8])
		n.Store.WriteChunk(chunk)
	})

	// REQUEST CHUNK
	t.RegisterHandler(p2p.MsgRequestChunk, func(p *p2p.Peer, msg p2p.Message) {
		var req p2p.ChunkRequestPayload
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			return
		}

		fmt.Printf("[%d] Received Request for Chunk %s\n", n.Config.Port, req.Hash[:8])

		// Check local storage
		chunk, err := n.Store.ReadChunk(req.Hash)
		if err != nil {
			// Don't have it
			return
		}

		// Send back
		payload, _ := json.Marshal(chunk)
		response := p2p.Message{
			Type:    p2p.MsgStoreChunk, // Replying with the chunk
			Sender:  n.DHT.ID.Hex(),
			Payload: payload,
		}

		n.Transport.SendMessage(p.Address, response) // Reply to sender
	})
}

package p2p

import (
	"net"
	"time"
)

// Peer represents a remote node in the network
type Peer struct {
	ID       string    `json:"id"`
	Address  string    `json:"address"` // IP:Port
	LastSeen time.Time `json:"last_seen"`
	Conn     net.Conn  `json:"-"`
	Outbound bool      `json:"outbound"`
}

// Node represents the local node instance
type Node struct {
	ID        string
	Address   string
	Peers     map[string]*Peer
	Transport Transport
}

// Transport handles the low-level network communication
type Transport interface {
	Listen(address string) error
	Dial(address string) error
	SendMessage(address string, msg Message) error
	Close() error
}

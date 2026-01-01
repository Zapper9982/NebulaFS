package dht

import (
	"sync"
)

// DHT represents the Distributed Hash Table
type DHT struct {
	ID           ID
	RoutingTable *RoutingTable
	Storage      map[string][]byte // Simple in-memory storage for K-V pairs (e.g. providers)
	Mutex        sync.RWMutex
}

// NewDHT creates a new DHT node
func NewDHT(id ID, address string) *DHT {
	self := Contact{ID: id, Address: address}
	return &DHT{
		ID:           id,
		RoutingTable: NewRoutingTable(self),
		Storage:      make(map[string][]byte),
	}
}

// --- RPC Handlers (Local Logic) ---

// HandlePing responds to a ping
func (dht *DHT) HandlePing(sender Contact) Contact {
	dht.RoutingTable.AddContact(sender)
	return dht.RoutingTable.Self
}

// HandleStore stores a key-value pair
func (dht *DHT) HandleStore(sender Contact, key string, value []byte) {
	dht.RoutingTable.AddContact(sender)
	dht.Mutex.Lock()
	defer dht.Mutex.Unlock()
	dht.Storage[key] = value
}

// HandleFindNode finds the K closest nodes to a target
func (dht *DHT) HandleFindNode(sender Contact, target ID) []Contact {
	dht.RoutingTable.AddContact(sender)
	return dht.RoutingTable.FindClosestContacts(target, K)
}

// HandleFindValue tries to find a value, otherwise returns closest nodes
func (dht *DHT) HandleFindValue(sender Contact, key string) ([]byte, []Contact) {
	dht.RoutingTable.AddContact(sender)
	dht.Mutex.RLock()
	val, ok := dht.Storage[key]
	dht.Mutex.RUnlock()

	if ok {
		return val, nil
	}

	// If not found, return closest nodes to the key's hash (treating key hash as ID)
	keyID := NewID(key)
	return nil, dht.RoutingTable.FindClosestContacts(keyID, K)
}

// --- High Level Operations ---

// AddNode adds a known node to the routing table
func (dht *DHT) AddNode(c Contact) {
	dht.RoutingTable.AddContact(c)
}

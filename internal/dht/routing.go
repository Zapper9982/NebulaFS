package dht

import (
	"sort"
	"sync"
)

// Bucket represents a K-bucket
type Bucket struct {
	Contacts []Contact
}

// RoutingTable manages peers in buckets
type RoutingTable struct {
	Self    Contact
	Buckets [IDLength * 8]*Bucket
	Mutex   sync.RWMutex
}

// NewRoutingTable creates a new routing table
func NewRoutingTable(self Contact) *RoutingTable {
	rt := &RoutingTable{
		Self: self,
	}
	for i := range rt.Buckets {
		rt.Buckets[i] = &Bucket{}
	}
	return rt
}

// AddContact adds a contact to the routing table
func (rt *RoutingTable) AddContact(c Contact) {
	rt.Mutex.Lock()
	defer rt.Mutex.Unlock()

	bucketIndex := rt.bucketIndex(c.ID)
	bucket := rt.Buckets[bucketIndex]

	// Check if already exists
	for i, existing := range bucket.Contacts {
		if existing.ID == c.ID {
			// Move to end (most recently seen)
			bucket.Contacts = append(bucket.Contacts[:i], bucket.Contacts[i+1:]...)
			bucket.Contacts = append(bucket.Contacts, c)
			return
		}
	}

	// Add if bucket not full
	if len(bucket.Contacts) < K {
		bucket.Contacts = append(bucket.Contacts, c)
	} else {
		// If bucket full, check if we can split or replace...
		// Simplified: drop if full (standard Kademlia ping-checks old ones first)
	}
}

// FindClosestContacts finds the K closest contacts to a target ID
func (rt *RoutingTable) FindClosestContacts(target ID, count int) []Contact {
	rt.Mutex.RLock()
	defer rt.Mutex.RUnlock()

	var allContacts []Contact
	for _, b := range rt.Buckets {
		allContacts = append(allContacts, b.Contacts...)
	}

	// Sort by distance
	sort.Slice(allContacts, func(i, j int) bool {
		distI := allContacts[i].ID.XOR(target).Int()
		distJ := allContacts[j].ID.XOR(target).Int()
		return distI.Cmp(distJ) < 0
	})

	if len(allContacts) > count {
		return allContacts[:count]
	}
	return allContacts
}

// bucketIndex calculates the index of the bucket for a given ID
func (rt *RoutingTable) bucketIndex(id ID) int {
	dist := rt.Self.ID.XOR(id)
	for i := 0; i < IDLength; i++ {
		for j := 0; j < 8; j++ {
			if (dist[i]>>uint8(7-j))&0x1 != 0 {
				return i*8 + j
			}
		}
	}
	return IDLength*8 - 1
}

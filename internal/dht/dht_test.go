package dht

import (
	"fmt"
	"testing"
)

func TestRoutingTable(t *testing.T) {
	selfID := NewID("node-self")
	self := Contact{ID: selfID, Address: "127.0.0.1:3000"}
	rt := NewRoutingTable(self)

	// Add some contacts
	for i := 0; i < 50; i++ {
		id := NewID(fmt.Sprintf("node-%d", i))
		rt.AddContact(Contact{ID: id, Address: fmt.Sprintf("127.0.0.1:%d", 4000+i)})
	}

	// Find closest
	target := NewID("target-key")
	closest := rt.FindClosestContacts(target, K)

	if len(closest) > K {
		t.Errorf("Expected at most %d contacts, got %d", K, len(closest))
	}

	// Verify order (closest first)
	lastDist := closest[0].ID.XOR(target).Int()
	for i := 1; i < len(closest); i++ {
		dist := closest[i].ID.XOR(target).Int()
		if dist.Cmp(lastDist) < 0 {
			t.Errorf("Contacts not sorted by distance")
		}
		lastDist = dist
	}
}

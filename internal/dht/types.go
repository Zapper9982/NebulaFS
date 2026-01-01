package dht

import (
	"crypto/sha1"
	"encoding/hex"
	"math/big"
)

const (
	// K is the bucket size (k-bucket)
	K = 20
	// Alpha is the concurrency parameter for lookups
	Alpha = 3
	// IDLength is the length of the NodeID in bytes (SHA-1 = 20)
	IDLength = 20
)

// ID represents a 160-bit SHA-1 hash
type ID [IDLength]byte

// Contact represents a node in the DHT
type Contact struct {
	ID      ID     `json:"id"`
	Address string `json:"address"`
}

// NewID creates a new ID from a string
func NewID(data string) ID {
	hash := sha1.Sum([]byte(data))
	return ID(hash)
}

// Hex returns the hex string representation of the ID
func (id ID) Hex() string {
	return hex.EncodeToString(id[:])
}

// XOR calculates the XOR distance between two IDs
func (id ID) XOR(other ID) ID {
	var result ID
	for i := 0; i < IDLength; i++ {
		result[i] = id[i] ^ other[i]
	}
	return result
}

// Int returns the big.Int representation of the ID (for distance comparison)
func (id ID) Int() *big.Int {
	return new(big.Int).SetBytes(id[:])
}

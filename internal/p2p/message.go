package p2p

import "encoding/json"

// MessageType defines the type of message
type MessageType string

const (
	MsgHandshake    MessageType = "HANDSHAKE"
	MsgDHTPing      MessageType = "DHT_PING"
	MsgDHTPong      MessageType = "DHT_PONG"
	MsgDHTStore     MessageType = "DHT_STORE"
	MsgDHTFindNode  MessageType = "DHT_FIND_NODE"
	MsgDHTFindValue MessageType = "DHT_FIND_VALUE"
	MsgStoreChunk   MessageType = "STORE_CHUNK"
	MsgRequestChunk MessageType = "REQUEST_CHUNK"
	MsgFileTransfer MessageType = "FILE_TRANSFER"
)

// Message represents a general P2P message
type Message struct {
	Type    MessageType     `json:"type"`
	Sender  string          `json:"sender"` // Sender ID
	Payload json.RawMessage `json:"payload"`
}

// DHTPayload represents general DHT data
type DHTPayload struct {
	TargetID string `json:"target_id,omitempty"`
	Key      string `json:"key,omitempty"`
	Value    []byte `json:"value,omitempty"`
	Contacts []byte `json:"contacts,omitempty"` // Marshaled list of contacts
}

// ChunkRequestPayload represents a request for a file chunk
type ChunkRequestPayload struct {
	Hash string `json:"hash"`
}

package p2p

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketTransport implements Transport using WebSockets
type WebSocketTransport struct {
	Address  string
	Upgrader websocket.Upgrader
	Handlers map[string]func(*Peer, Message)
	Peers    map[string]*Peer
	Mutex    sync.RWMutex
}

func NewWebSocketTransport(address string) *WebSocketTransport {
	return &WebSocketTransport{
		Address: address,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		Handlers: make(map[string]func(*Peer, Message)),
		Peers:    make(map[string]*Peer),
	}
}

func (t *WebSocketTransport) RegisterHandler(msgType MessageType, handler func(*Peer, Message)) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	t.Handlers[string(msgType)] = handler
}

func (t *WebSocketTransport) Listen(address string) error {
	if address != "" {
		t.Address = address
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", t.handleWS)
	// ListenAndServe uses DefaultServeMux if handler is nil.
	// We must pass our mux.
	server := &http.Server{
		Addr:    t.Address,
		Handler: mux,
	}
	return server.ListenAndServe()
}

func (t *WebSocketTransport) Dial(address string) error {
	t.Mutex.Lock()
	if _, exists := t.Peers[address]; exists {
		t.Mutex.Unlock()
		return nil
	}
	t.Mutex.Unlock()

	url := fmt.Sprintf("ws://%s/ws", address)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}

	t.handleNewConnection(conn, address, true)
	return nil
}

func (t *WebSocketTransport) SendMessage(address string, msg Message) error {
	// Ensure connected
	if err := t.Dial(address); err != nil {
		return err
	}

	t.Mutex.RLock()
	peer, exists := t.Peers[address]
	t.Mutex.RUnlock()

	if !exists {
		return fmt.Errorf("peer %s not connected", address)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = peer.Conn.Write(data)
	return err
}

func (t *WebSocketTransport) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := t.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Upgrade failed: %v\n", err)
		return
	}
	// Initial ID is unknown, will be set on first message with Sender field
	t.handleNewConnection(conn, conn.RemoteAddr().String(), false)
}

func (t *WebSocketTransport) handleNewConnection(conn *websocket.Conn, address string, outbound bool) {
	adapter := NewWSConnAdapter(conn)
	peer := &Peer{
		Conn:     adapter,
		Address:  address,
		Outbound: outbound,
	}

	t.Mutex.Lock()
	t.Peers[address] = peer
	t.Mutex.Unlock()

	go t.readLoop(peer, conn)
}

func (t *WebSocketTransport) readLoop(peer *Peer, conn *websocket.Conn) {
	defer func() {
		conn.Close()
		t.Mutex.Lock()
		delete(t.Peers, peer.Address)
		t.Mutex.Unlock()
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			fmt.Printf("Error unmarshaling message: %v\n", err)
			continue
		}

		if msg.Sender != "" {
			peer.ID = msg.Sender
		}

		t.Mutex.RLock()
		handler, exists := t.Handlers[string(msg.Type)]
		t.Mutex.RUnlock()

		if exists {
			handler(peer, msg)
		}
	}
}

func (t *WebSocketTransport) Close() error {
	return nil
}

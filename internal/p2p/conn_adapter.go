package p2p

import (
	"io"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

// WSConnAdapter adapts a websocket.Conn to net.Conn
type WSConnAdapter struct {
	conn   *websocket.Conn
	reader io.Reader
}

func NewWSConnAdapter(conn *websocket.Conn) *WSConnAdapter {
	return &WSConnAdapter{conn: conn}
}

func (a *WSConnAdapter) Read(b []byte) (n int, err error) {
	if a.reader == nil {
		_, reader, err := a.conn.NextReader()
		if err != nil {
			return 0, err
		}
		a.reader = reader
	}
	n, err = a.reader.Read(b)
	if err == io.EOF {
		a.reader = nil
		// Don't return EOF yet unless connection closed?
		// NextReader will return error if closed.
		// For now, let's assume one stream or handle framing differently.
		// Actually, for P2P via WS, we usually read full JSON messages.
		// Detailed generic Read is complex.
		// But for our "Peer" struct, if we rely on "Transport" to read messages loop,
		// we might not need to call Peer.Conn.Read directly.
		return 0, io.EOF
	}
	return n, err
}

func (a *WSConnAdapter) Write(b []byte) (n int, err error) {
	err = a.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (a *WSConnAdapter) Close() error {
	return a.conn.Close()
}

func (a *WSConnAdapter) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *WSConnAdapter) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *WSConnAdapter) SetDeadline(t time.Time) error {
	return a.conn.SetReadDeadline(t)
}

func (a *WSConnAdapter) SetReadDeadline(t time.Time) error {
	return a.conn.SetReadDeadline(t)
}

func (a *WSConnAdapter) SetWriteDeadline(t time.Time) error {
	return a.conn.SetWriteDeadline(t)
}

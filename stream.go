package wsio

import (
	"github.com/gorilla/websocket"
)

// Stream implements io.Writer and io.Reader on top of a websocket connection
type Stream struct {
	Conn *websocket.Conn
}

// NewStream returns a new Stream instance for the given websocket connection
func NewStream(c *websocket.Conn) *Stream {
	return &Stream{
		Conn: c,
	}
}

func (s *Stream) Write(p []byte) (n int, err error) {
	err = s.Conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return -1, err
	}
	return len(p), err
}

func (s *Stream) Read(p []byte) (n int, err error) {
	_, b, err := s.Conn.ReadMessage()
	return len(b), err
}

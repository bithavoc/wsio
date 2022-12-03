package wsio

import (
	"errors"
	"io"
	"sync"

	"github.com/gorilla/websocket"
)

// Stream implements io.Writer and io.Reader on top of a websocket connection
type Stream struct {
	Conn      *websocket.Conn
	readBuf   []byte
	readMutex sync.Mutex
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
	return len(p), nil
}

func (s *Stream) Read(p []byte) (n int, err error) {
	s.readMutex.Lock()
	defer s.readMutex.Unlock()
	var b []byte
	var messageType int
	if len(s.readBuf) > 0 {
		b = s.readBuf
		s.readBuf = nil
	} else {
	readNextMessage:
		messageType, b, err = s.Conn.ReadMessage()
		// log.Printf("received message type, %v, %T", messageType, err)
		var closeError *websocket.CloseError
		if errors.As(err, &closeError) {
			return endOfFile()
		} else if err != nil {
			return 0, err
		}
		if messageType == websocket.CloseMessage {
			return endOfFile()
		}
		if messageType != websocket.BinaryMessage && messageType != websocket.TextMessage {
			// log.Printf("received text message, reading more")
			goto readNextMessage
		}
	}
	n = copy(p, b)
	s.readBuf = b[n:len(b)]
	return n, err
}

func endOfFile() (int, error) {
	return 0, io.EOF
}

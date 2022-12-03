package wsio_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bithavoc/wsio"

	"crypto/rand"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestReadShort(t *testing.T) {
	content := []string{
		`msg 1`,
		`msg 2`,
		`msg 3`,
		`msg 4`,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("failed to upgrade, %s", err.Error())
			return
		}
		defer conn.Close()
		for _, msg := range content {
			conn.WriteJSON(msg)
			// wrt, err := conn.NextWriter(websocket.BinaryMessage)
			// if err != nil {
			// 	return
			// }
			// wrt.Write([]byte(msg))
		}
	})
	clientsvr := httptest.NewServer(mux)
	defer clientsvr.Close()
	baseURL := strings.ReplaceAll(clientsvr.URL, "http://", "ws://")
	connA, _, err := websocket.DefaultDialer.Dial(baseURL+"/ws", nil)
	require.Nil(t, err)
	defer connA.Close()
	wst := wsio.NewStream(connA)
	dec := json.NewDecoder(wst)

	result := []string{}

	for {
		t.Logf("receiving")
		ctrl := ""
		if err := dec.Decode(&ctrl); err != nil {
			if err != io.EOF {
				t.Logf("failed to deserialize control: %s", err.Error())
			}
			break
		}
		result = append(result, ctrl)
		t.Logf("control command received: %#v", ctrl)
	}
	require.Equal(t, content, result)
}

func TestReadClose(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("failed to upgrade, %s", err.Error())
			return
		}
		defer conn.Close()
	})
	clientsvr := httptest.NewServer(mux)
	defer clientsvr.Close()
	baseURL := strings.ReplaceAll(clientsvr.URL, "http://", "ws://")
	connA, _, err := websocket.DefaultDialer.Dial(baseURL+"/ws", nil)
	require.Nil(t, err)
	defer connA.Close()
	wst := wsio.NewStream(connA)
	dec := json.NewDecoder(wst)

	msg := ""
	err = dec.Decode(&msg)
	require.Equal(t, io.EOF, err)
}

func TestReadLarge(t *testing.T) {
	buff := make([]byte, 1000*1000)
	rand.Read(buff)
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("failed to upgrade, %s", err.Error())
			return
		}
		defer conn.Close()
		conn.WriteMessage(websocket.BinaryMessage, buff)
	})
	clientsvr := httptest.NewServer(mux)
	defer clientsvr.Close()
	baseURL := strings.ReplaceAll(clientsvr.URL, "http://", "ws://")
	connA, _, err := websocket.DefaultDialer.Dial(baseURL+"/ws", nil)
	require.Nil(t, err)
	defer connA.Close()
	wst := wsio.NewStream(connA)

	readAll, err := io.ReadAll(wst)
	require.Nil(t, err)
	require.Equal(t, buff, readAll)
}

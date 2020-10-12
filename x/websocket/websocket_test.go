package websocket

import (
	"github.com/stretchr/testify/require"
	"mercury/x/log"
	"net/http"
	"os"
	"testing"
	"time"
)

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := Upgrade(w, r)
	if err != nil {
		log.Error("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		message, err := c.ReadMessage()
		if err != nil {
			log.Error("read:", err)
			break
		}
		log.Info("recv:", string(message))
		if string(message) == "err" {
			c.Close()
			return
		}
		err = c.WriteMessage(TextMessage, message)
		if err != nil {
			log.Error("write:", err)
			break
		}
	}
}

func TestConnection(t *testing.T) {
	// Initialize log
	lvl, _ := log.LvlFromString("info")
	log.Root().SetHandler(log.LvlFilterHandler(lvl, log.StreamHandler(os.Stdout, log.LogfmtFormat())))

	go func() {
		http.HandleFunc("/", echo)
		err := http.ListenAndServe(":8081", nil)
		t.Log("listen err", err)
	}()

	time.Sleep(time.Second)
	connection, err := Dial("ws://127.0.0.1:8081")
	require.Nil(t, err)
	data := []byte("message 1")
	err = connection.WriteTextMessage(data)
	require.Nil(t, err)
	readData, err := connection.ReadMessage()
	require.Nil(t, err)
	require.Equal(t, data, readData)
	err = connection.WriteTextMessage([]byte("err"))
	require.Nil(t, err)
	time.Sleep(time.Second)
	err = connection.WriteTextMessage(data)
	require.Nil(t, err)
	readData, err = connection.ReadMessage()
	require.Error(t, err)
}

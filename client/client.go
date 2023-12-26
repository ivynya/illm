package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// authorize to an illm relay as a provider
	u := url.URL{Scheme: "wss", Host: "io.ivy.direct", Path: "/aura/provider"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{
		"Authorization": []string{"Basic aXZ5LWF1cmEtYWRtaW46R21XNlhkOHZoVWhLM1hrQVJoNFo="},
	})
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	log.Printf("connected to %s", u.String())

	// websocket client read loop
	done := make(chan struct{})
	go read(c, done)

	// program maintainance loop
	ticker := time.NewTicker(time.Second * 45)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("{\"action\": \"ping\"}"))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func read(c *websocket.Conn, done chan struct{}) {
	defer close(done)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		req, err := decodeRequest(message)
		if err != nil {
			log.Println("decode:", err)
			return
		}
		log.Printf("recv: %s", req.Action)

		switch req.Action {
		case "generate":
			log.Println("generate")
			completion, err := generate(c, req)
			if err != nil {
				log.Println("generate:", err)
				return
			}
			_ = completion
		}
	}
}

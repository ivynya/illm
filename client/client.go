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

// global environment variables
var (
	auth       = os.Getenv("AUTH")
	identifier = os.Getenv("IDENTIFIER")
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// authorize to an illm relay as a provider
	u := url.URL{Scheme: "wss", Host: "io.ivy.direct", Path: "/aura/provider"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{
		"Authorization": []string{"Basic " + auth},
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
		log.Printf("recv: %s (tag %s)", req.Action, req.Tag)

		switch req.Action {
		case "generate":
			completion, err := generate(c, req)
			if err != nil {
				log.Println("generate:", err)
				return
			}
			_ = completion
		case "identify":
			res, err := encodeRequest(req.Tag, "identify", identifier)
			if err != nil {
				log.Println("encode:", err)
				return
			}
			err = c.WriteMessage(websocket.TextMessage, res)
		case "summarize-youtube":
			complete, err := summarize(c, req)
			if err != nil {
				log.Println("summarize:", err)
				return
			}
			_ = complete
		}
	}
}

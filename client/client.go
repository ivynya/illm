package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// Request struct
type Request struct {
	Action   string `json:"action"`
	Data     string `json:"data"`
	Generate struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
	} `json:"generate"`
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: "io.ivy.direct", Path: "/aura"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{
		"Authorization": []string{"Basic aXZ5LWF1cmEtYWRtaW46R21XNlhkOHZoVWhLM1hrQVJoNFo="},
	})
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			req := &Request{}
			err = json.Unmarshal(message, &req)
			if err != nil {
				log.Println("json:", err)
				return
			}
			log.Printf("recv: %s", req.Action)

			switch req.Action {
			case "generate":
				log.Println("generate")
				llm, err := ollama.New(ollama.WithModel(req.Generate.Model))
				if err != nil {
					log.Fatal(err)
				}
				ctx := context.Background()
				completion, err := llm.Call(ctx, req.Generate.Prompt,
					llms.WithTemperature(0.8),
					llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
						responseMessage := string(chunk)
						response := &Request{
							Action: "response",
							Data:   responseMessage,
						}
						responseJson, err := json.Marshal(response)
						if err != nil {
							log.Fatal(err)
						}
						c.WriteMessage(websocket.TextMessage, responseJson)
						return nil
					}),
				)
				if err != nil {
					log.Fatal(err)
				}

				_ = completion
				c.WriteMessage(websocket.TextMessage, []byte("{\"action\": \"response-end\"}"))
			case "stop":
				log.Println("stop")
			}
		}
	}()

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

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
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

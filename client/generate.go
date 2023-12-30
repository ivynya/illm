package main

import (
	"context"
	"log"

	"github.com/gorilla/websocket"
	"github.com/ivynya/illm/internal"
	"github.com/ivynya/illm/ollama"
	"github.com/tmc/langchaingo/llms"
)

func generate(c *websocket.Conn, req *internal.Request) ([]*llms.Generation, error) {
	llm, err := ollama.New(ollama.WithModel(req.Generate.Model), ollama.WithServerURL(ollama_url))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llm.Generate(ctx,
		[]string{req.Generate.Prompt},
		req.Generate.Context,
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			resp, err := encodeRequest(req.Tag, "response", string(chunk))
			if err != nil {
				log.Fatal(err)
			}
			c.WriteMessage(websocket.TextMessage, resp)
			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	return completion, nil
}

package main

import (
	"context"
	"log"

	"github.com/gorilla/websocket"
	"github.com/ivynya/illm/ollama"
	"github.com/tmc/langchaingo/llms"
)

func generate(c *websocket.Conn, req *Request) ([]*llms.Generation, error) {
	llm, err := ollama.New(ollama.WithModel(req.Generate.Model), ollama.WithServerURL("http://host.docker.internal:11434"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llm.Generate(ctx,
		[]string{req.Generate.Prompt},
		req.Generate.Context,
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			resp, err := encodeRequest("response", string(chunk))
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

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/websocket/v2"
	"github.com/ivynya/illm/internal"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func main() {
	clients := make(map[string]*websocket.Conn)
	providers := make(map[string]*websocket.Conn)

	app := fiber.New()
	app.Use(basicauth.New(basicauth.Config{
		Users: map[string]string{
			"ivy-aura-admin": "GmW6Xd8vhUhK3XkARh4Z",
		},
	}))

	// Provider websocket endpoint
	app.Get("/aura/provider", websocket.New(func(c *websocket.Conn) {
		// Register new provider and give it a random tag
		tag, err := gonanoid.New()
		if err != nil {
			log.Fatal(err)
		}
		providers[tag] = c

		// Log join message
		fmt.Println("Provider joined from " + c.RemoteAddr().String())
		fmt.Println("Total providers:", len(providers))
		updateConnCount("providers", &providers)

		for {
			// Read message from provider
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("Websocket read error:", err)
				break
			}

			// Decode message into request struct
			req := &internal.Request{}
			err = json.Unmarshal(msg, &req)
			if err != nil {
				log.Println("JSON decode error:", err)
				break
			}

			// Relay message to client with matching tag
			err = broadcastToClient(clients, req)
			if err != nil {
				log.Println("Websocket write error:", err)
				// Delete client if it is no longer connected
				delete(clients, req.Tag)
			}
		}

		// Unregister provider
		delete(providers, tag)
		updateConnCount("providers", &providers)
	}))

	// WebSocket endpoint
	app.Get("/aura/client", websocket.New(func(c *websocket.Conn) {
		// Register new client and give it a random tag
		tag, err := gonanoid.New()
		if err != nil {
			log.Fatal(err)
		}
		clients[tag] = c

		// Log join message
		fmt.Println("Client joined from " + c.RemoteAddr().String())
		fmt.Println("Total clients:", len(clients))
		updateConnCount("clients", &clients)

		for {
			// Read message from client
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("Websocket read error:", err)
				break
			}

			// Decode message into request struct
			req := &internal.Request{}
			err = json.Unmarshal(msg, &req)
			if err != nil {
				log.Println("JSON decode error:", err)
				break
			}

			// Tag request with client tag
			req.Tag = tag

			// Send request to provider
			err = broadcastToProvider(providers, req)
			if err != nil {
				log.Println("Websocket write error:", err)
				// Delete provider if it is no longer connected
				delete(providers, req.Tag)
				// Send error message to client
				c.WriteMessage(websocket.TextMessage, []byte(`{"action":"error","data":"Provider disconnected"}`))
			}
		}

		// Unregister client
		delete(clients, tag)
		updateConnCount("clients", &clients)
	}))

	// Start the server
	log.Fatal(app.Listen(":3000"))
}

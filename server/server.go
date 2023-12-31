package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/websocket/v2"
	"github.com/ivynya/illm/internal"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
	username = os.Getenv("USERNAME")
	password = os.Getenv("PASSWORD")
)

func main() {
	clients := make(map[string]*websocket.Conn)
	providers := make(map[string]*websocket.Conn)

	app := fiber.New()
	app.Use(basicauth.New(basicauth.Config{
		Users: map[string]string{
			username: password,
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
		broadcastConnectionStats(clients, providers)

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

			// No tag means won't be sent to any client
			if req.Tag == "" {
				continue
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
		broadcastConnectionStats(clients, providers)
	}))

	// WebSocket endpoint
	app.Get("/aura/client", websocket.New(func(c *websocket.Conn) {
		// Register new client and give it a random tag
		tag, err := gonanoid.New()
		if err != nil {
			log.Fatal(err)
		}
		clients[tag] = c

		// Log join message and broadcast counts
		fmt.Println("Client joined from " + c.RemoteAddr().String())
		fmt.Println("Total clients:", len(clients))
		broadcastConnectionStats(clients, providers)

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

			// If action is identify, broadcast to all providers
			if req.Action == "identify" {
				broadcastAll(providers, req)
				continue
			}

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
		broadcastConnectionStats(clients, providers)
	}))

	// Start the server
	log.Fatal(app.Listen(":3000"))
}

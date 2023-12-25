package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/websocket/v2"
)

func main() {
	clients := make(map[*websocket.Conn]bool)

	app := fiber.New()
	app.Use(basicauth.New(basicauth.Config{
		Users: map[string]string{
			"ivy": "GmW6Xd8vhUhK3XkARh4Z",
		},
	}))

	// WebSocket endpoint
	app.Get("/aura", websocket.New(func(c *websocket.Conn) {
		// Register new client
		clients[c] = true

		// Log join message
		fmt.Println("Client joined from " + c.RemoteAddr().String())

		for {
			// Read message from client
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("Websocket read error:", err)
				break
			}
			// Disconnect client when it sends something that isnt json
			if !json.Valid(msg) {
				log.Println("Invalid JSON")
				break
			}

			// Iterate through all clients
			for client := range clients {
				err := client.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Println("Websocket write error:", err)
					delete(clients, client)
				}
			}
		}

		// Unregister client
		delete(clients, c)
	}))

	// Start the server
	log.Fatal(app.Listen(":3000"))
}

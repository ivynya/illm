package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func main() {
	app := fiber.New()

	// WebSocket endpoint
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		for {
			// Read message from client
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("Websocket read error:", err)
				break
			}

			// Print received message
			fmt.Println("Received message:", string(msg))

			// Write message back to client
			err = c.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("Websocket write error:", err)
				break
			}
		}
	}))

	// Streaming POST endpoint
	app.Post("/stream", func(c *fiber.Ctx) error {
		// Set response headers
		c.Set(fiber.HeaderContentType, "text/plain")
		c.Set(fiber.HeaderContentDisposition, "attachment; filename=\"stream.txt\"")

		write := func(w io.Writer) bool {
			// Write data to the response writer
			_, err := w.Write([]byte("Streaming data...\n"))
			if err != nil {
				log.Println("Streaming write error:", err)
				return false
			}

			// Simulate delay between writes
			time.Sleep(1 * time.Second)

			// Continue streaming
			return true
		}

		// Start streaming response
		c.Status(http.StatusOK)
		write(c)
		write(c)

		return nil
	})

	// Start the server
	log.Fatal(app.Listen(":3000"))
}

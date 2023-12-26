package main

import (
	"encoding/json"
	"math/rand"
	"strconv"

	"github.com/gofiber/websocket/v2"
	"github.com/ivynya/illm/internal"
)

func tagRequest(tag string, req *internal.Request) *internal.Request {
	req.Tag = tag
	return req
}

// pick a random provider and send the request to it
func broadcastToProvider(c map[string]*websocket.Conn, req *internal.Request) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	pick := rand.Intn(len(c))
	for _, provider := range c {
		if pick == 0 {
			err := provider.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				return err
			}
			return nil
		}
		pick--
	}
	return nil
}

func broadcastToClient(c map[string]*websocket.Conn, req *internal.Request) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	err = c[req.Tag].WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return err
	}
	return nil
}

// broadcast to all connections and return false if >= 1 failure
func broadcastAll(c map[string]*websocket.Conn, req *internal.Request) bool {
	data, _ := json.Marshal(req)

	ok := true
	for tag, conn := range c {
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			delete(c, tag)
			ok = false
		}
	}
	return ok
}

// broadcast number of clients and providers to all clients
func broadcastConnectionStats(clients map[string]*websocket.Conn, providers map[string]*websocket.Conn) {
	retry_remaining := 3
	ok := true
	for retry_remaining > 0 {
		ok = ok && broadcastAll(clients, &internal.Request{
			Action: "clients",
			Data:   strconv.Itoa(len(clients)),
		})
		ok = ok && broadcastAll(clients, &internal.Request{
			Action: "providers",
			Data:   strconv.Itoa(len(providers)),
		})
		if ok {
			break
		}
		retry_remaining--
	}
}

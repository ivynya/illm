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

func broadcastAll(c map[string]*websocket.Conn, req *internal.Request) error {
	data, _ := json.Marshal(req)

	for tag, conn := range c {
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			delete(c, tag)
			return err
		}
	}
	return nil
}

func updateConnCount(clientType string, c *map[string]*websocket.Conn) {
	err := broadcastAll(*c, &internal.Request{
		Action: clientType,
		Data:   strconv.Itoa(len(*c)),
	})
	if err != nil {
		updateConnCount(clientType, c)
	}
}

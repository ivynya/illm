package main

import (
	"encoding/json"

	"github.com/ivynya/illm/internal"
)

func decodeRequest(message []byte) (*internal.Request, error) {
	req := &internal.Request{}
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func encodeRequest(tag string, action string, data string) ([]byte, error) {
	resp := &internal.Request{
		Tag:    tag,
		Action: action,
		Data:   data,
	}
	respJson, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return respJson, nil
}

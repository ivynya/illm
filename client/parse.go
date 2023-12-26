package main

import (
	"encoding/json"
)

func decodeRequest(message []byte) (*Request, error) {
	req := &Request{}
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func encodeRequest(action string, data string) ([]byte, error) {
	resp := &Request{
		Action: action,
		Data:   data,
	}
	respJson, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return respJson, nil
}

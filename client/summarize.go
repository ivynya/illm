package main

import (
	"encoding/json"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/ivynya/illm/internal"
	"github.com/ivynya/illm/ollama"
	"github.com/kkdai/youtube/v2"
)

func summarize(c *websocket.Conn, req *internal.Request) (bool, error) {
	videoID := req.Data
	client := youtube.Client{}

	video, err := client.GetVideo(videoID)
	if err != nil {
		return false, err
	}

	transcript, err := client.GetTranscript(video)
	if err != nil {
		return false, err
	}

	info := &ollama.GenerateResponse{
		Model:    req.Generate.Model,
		Response: "Video: `" + video.Title + "`\nTranscript length: `" + strconv.Itoa(len(transcript.String())) + "`\n\n",
		Done:     false,
	}
	infoJson, err := json.Marshal(info)
	if err != nil {
		return false, err
	}
	infoResp, err := encodeRequest(req.Tag, "response", string(infoJson))
	if err != nil {
		return false, err
	}
	c.WriteMessage(websocket.TextMessage, infoResp)

	req.Generate.Prompt = "Summarize the following video. Only include information from the video in your response. Video: " + video.Title + "\n\n" + transcript.String() + "\n\nSummary:"
	req.Generate.Context = []int{}

	complete, err := generate(c, req)
	if err != nil {
		return false, err
	}
	_ = complete

	return true, nil
}

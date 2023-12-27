package main

import (
	"github.com/gorilla/websocket"
	"github.com/ivynya/illm/internal"
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

	req.Generate.Prompt = "Summarize the following video: " + video.Title + "\n\n" + transcript.String() + "\n\nSummary:"
	req.Generate.Context = []int{}

	complete, err := generate(c, req)
	if err != nil {
		return false, err
	}
	_ = complete

	return true, nil
}

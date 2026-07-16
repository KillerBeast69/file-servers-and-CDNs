package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {
	var out bytes.Buffer

	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)

	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to obtain aspect ratio: %v", err)
	}

	type stream struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}

	type ProbeOutput struct {
		Streams []stream `json:"streams"`
	}

	var data ProbeOutput
	if err := json.Unmarshal(out.Bytes(), &data); err != nil {
		return "", err
	}

	if len(data.Streams) == 0 {
		return "other", nil
	}

	width := data.Streams[0].Width
	height := data.Streams[0].Height

	if width == 0 || height == 0 {
		return "other", nil
	}

	if width*9 == height*16 {
		return "landscape", nil
	}

	if width*16 == height*9 {
		return "portrait", nil
	}
	return "other", nil
}

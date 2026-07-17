package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	videoURL := *video.VideoURL

	parts := strings.Split(videoURL, ",")
	if len(parts) != 2 {
		return video, fmt.Errorf("invalid url format")
	}

	bucket := parts[0]
	key := parts[1]

	signedURL, err := generatePresignedURL(cfg.s3Client, bucket, key, 15*time.Minute)
	if err != nil {
		return video, err
	}

	video.VideoURL = &signedURL

	return video, nil
}

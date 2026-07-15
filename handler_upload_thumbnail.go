package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here

	const maxMemory = 10 << 20
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to parse form", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to parse form file", err)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")

	parsedMediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid content type", err)
		return
	}

	if parsedMediaType != "image/jpeg" && parsedMediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "please upload jpeg or png file", nil)
		return
	}

	mediaParts := strings.Split(parsedMediaType, "/")
	if len(mediaParts) != 2 {
		respondWithError(w, http.StatusBadRequest, "invalid content type", nil)
		return
	}
	extension := mediaParts[1]

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't get video", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "not authorized to update this video", nil)
		return
	}

	randBytes := make([]byte, 32)
	_, err = rand.Read(randBytes)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to generate random bytes", err)
		return
	}

	randString := base64.RawURLEncoding.EncodeToString(randBytes)

	fileName := fmt.Sprintf("%s.%s", randString, extension)
	filePath := filepath.Join(cfg.assetsRoot, fileName)

	output, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create file", err)
		return
	}

	defer output.Close()

	_, err = io.Copy(output, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to copy file", err)
		return
	}

	thumbnailURL := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, fileName)
	video.ThumbnailURL = &thumbnailURL

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't update video record", err)
		return
	}

	respondWithJSON(w, http.StatusOK, struct{}{})
}

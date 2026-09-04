# Tubely — Video Hosting with S3 & CloudFront

A backend service (with a minimal vanilla-JS frontend) for uploading, storing, and streaming videos — built while learning how to serve files at scale using AWS S3 and CloudFront.

Built from the [Learn File Servers and CDNs with S3 and CloudFront](https://www.boot.dev/courses/learn-file-servers-s3-cloudfront-golang) course on [boot.dev](https://www.boot.dev).

## Features

- **Auth** — JWT-based access tokens with refresh/revoke support, Argon2id password hashing
- **Video metadata** — create, retrieve, list, and delete video records (SQLite-backed)
- **Thumbnail uploads** — stored locally under `/assets`, served with cache-busting headers
- **Video uploads**
  - Validates MP4 content type
  - Detects aspect ratio (landscape/portrait/other) via `ffprobe`
  - "Fast start" processing via `ffmpeg` (moves the `moov` atom to the front for faster playback)
  - Uploads processed video to an S3 bucket, keyed by aspect ratio
  - Serves the final video through a CloudFront distribution
- **Static frontend** — simple login/signup + video management UI in `app/`

## Requirements

- [Go](https://go.dev/doc/install) 1.24+
- [ffmpeg & ffprobe](https://ffmpeg.org/download.html) (must be in your `PATH`)
- [SQLite 3](https://www.sqlite.org/download.html) (optional, for inspecting the DB directly)
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) configured with credentials (`aws configure`)
- An S3 bucket and CloudFront distribution

## Getting Started

### 1. Install dependencies

```bash
go mod download
```

### 2. Download sample assets (optional)

```bash
./samplesdownload.sh
```

Downloads sample images/videos into `samples/` for local testing.

### 3. Configure environment variables

```bash
cp .env.example .env
```

Then fill in the values:

```
DB_PATH="./tubely.db"
JWT_SECRET="<your-secret>"
PLATFORM="dev"
FILEPATH_ROOT="./app"
ASSETS_ROOT="./assets"
S3_BUCKET="<your-bucket>"
S3_REGION="<your-region>"
S3_CF_DISTRO="<your-cloudfront-domain>"
PORT="8091"
```

AWS credentials are picked up automatically from `~/.aws/credentials`.

### 4. Run the server

```bash
go run .
```

Visit `http://localhost:8091/app/` in your browser. On first run, a `tubely.db` SQLite file and an `assets/` directory will be created automatically.

## API Overview

| Method | Path                              | Description                     |
|--------|------------------------------------|----------------------------------|
| POST   | `/api/users`                      | Create a user                   |
| POST   | `/api/login`                      | Log in, get access + refresh tokens |
| POST   | `/api/refresh`                    | Exchange a refresh token for a new access token |
| POST   | `/api/revoke`                     | Revoke a refresh token          |
| POST   | `/api/videos`                     | Create video metadata (draft)   |
| GET    | `/api/videos`                     | List the current user's videos  |
| GET    | `/api/videos/{videoID}`           | Get a single video               |
| DELETE | `/api/videos/{videoID}`           | Delete a video                   |
| POST   | `/api/thumbnail_upload/{videoID}` | Upload a JPEG/PNG thumbnail      |
| POST   | `/api/video_upload/{videoID}`     | Upload an MP4 video (→ S3/CloudFront) |
| POST   | `/admin/reset`                    | Reset the database (dev only)   |

## Project Structure

```
main.go                     Server setup, config, routes
handler_*.go                HTTP handlers (auth, users, videos, uploads)
assets.go                   Local assets directory setup
cache.go                    No-cache middleware for served assets
fast_video.go                ffmpeg fast-start processing
get_aspect_ratio.go          ffprobe aspect ratio detection
json.go                      JSON response helpers
reset.go                     Dev-only DB reset endpoint
internal/
  auth/                      JWT, password hashing, bearer token parsing
  database/                  SQLite client, users/videos/refresh_tokens
app/                         Static frontend (HTML/CSS/JS)
```

## How It Works

1. A user logs in and creates a video draft via `POST /api/videos`.
2. A thumbnail is uploaded and saved locally, served from `/assets`.
3. The video file is uploaded, validated as MP4, processed with `ffmpeg` for fast-start playback, and its aspect ratio is detected with `ffprobe`.
4. The processed file is uploaded to S3 under a key like `landscape/<random-hex>.mp4`.
5. The video's `video_url` is set to the CloudFront distribution URL and saved to the database, ready for streaming in the browser.

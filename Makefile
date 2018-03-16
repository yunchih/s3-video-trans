
build:
	go build github.com/yunchih/s3-video-trans/cmd/minio-video-transcoder

static:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-s -w -extldflags "-static"' github.com/yunchih/s3-video-trans/cmd/minio-video-transcoder

# S3 video transcoder

This is a tool that downloads video objects from S3-compatible storage (like Minio)
, transcodes them into a desire format and uploads them back (the new filename is prefixed
with a marker to signal it's been transcoded).

Currently, the Minio implementation is provided.
You can find the transcoding command and the prefix in the `cmd/minio-video-transcoder/transcode.go` file.

## Usage

```bash
# Build it
go build github.com/yunchih/s3-video-trans/cmd/minio-video-transcoder

# Run the following command to see usage
./minio-video-transcoder
```

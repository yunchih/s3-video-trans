package main

import (
	"fmt"
	"log"
	"os"

	"github.com/yunchih/s3-video-trans/pkg/env"
	"github.com/yunchih/s3-video-trans/pkg/minio"
	"github.com/yunchih/s3-video-trans/pkg/s3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide filename to be uploaded")
		os.Exit(1)
	}

	minioEndpoint := env.Get("MINIO_ENDPOINT")
	minioAccessID := env.Get("MINIO_ID")
	minioAccessKey := env.Get("MINIO_KEY")
	minioUseSSL := env.GetBool("MINIO_USESSL")
	minioBucket := env.Get("MINIO_BUCKET")

	if minioEndpoint == "" || minioAccessID == "" || minioAccessKey == "" || minioBucket == "" {
		os.Exit(1)
	}

	cfg := s3.Config{minioEndpoint, minioAccessID, minioAccessKey, minioUseSSL}
	mc, err := minio.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	target := os.Args[1]

	if err := mc.Remove(minioBucket, target); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Successfully remove object %s\n", target)
}

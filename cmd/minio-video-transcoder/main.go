package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "sort"
    "strings"
    "time"

    "github.com/yunchih/s3-video-trans/pkg/env"
    "github.com/yunchih/s3-video-trans/pkg/minio"
    "github.com/yunchih/s3-video-trans/pkg/s3"
    "github.com/yunchih/s3-video-trans/pkg/worker"
)

var helpMsg = `
  Usage of minio-video-transcoder:
      ./minio-video-transcoder -days=[N]

      Required commandline arguments:
          -days=N: Transcode video objects whose last modification time is
                   N days until now (the value can be any positive real number)

      Optional commandline arguments:
          -workers=N: The number of transcoding worker threads (default to 1)

      Required environment variables:
          MINIO_ENDPOINT: The URL of minio server (with port)
          MINIO_ID: Your Minio user ID
          MINIO_KEY: Your Minio user password
          MINIO_USESSL: Whether or not to connect with SSL enabled ("yes" or "no")
          MINIO_BUCKET: Name of target bucket
`

func process (mc *minio.Minio, minioBucket string, objKey string, dumpTranscoder bool) {
    tmpSrc := getTempFile()
    defer os.Remove(tmpSrc.Name())

    log.Printf("Downloading object '%s'\n", objKey)
    if err := mc.Download(minioBucket, objKey, tmpSrc); err != nil {
        log.Fatal(err)
    }

    tmpDst := getTempFile()
    defer os.Remove(tmpDst.Name())
    log.Printf("Transcoding object '%s'\n", objKey)
    transcoderOutput := runTranscoder(tmpSrc.Name(), tmpDst.Name())
    runTranscoder(tmpSrc.Name(), tmpDst.Name())

    transcoded := renamePrefix + objKey
    log.Printf("Uploading transcoded object to '%s'\n", transcoded)
    if err := mc.Upload(minioBucket, transcoded, tmpDst); err != nil {
        log.Fatal(err)
    }

    tmpSrc.Close()
    tmpDst.Close()

    if dumpTranscoder {
        log.Print(transcoderOutput)
    }
}

func help() {
    fmt.Print("\n")
    fmt.Print(helpMsg)
    fmt.Print("\n")
    os.Exit(1)
}

func main () {
    backInDays := flag.Float64("days", -1.0, "The range in days that we look back to search for possible targets")
    workers    := flag.Int64("workers", 1, "The number of transcoding worker threads (default to 1)")
    dump       := flag.String("dump", "no", "Whether or not to dump transcoder output")

    flag.Parse()

    if *backInDays < 0.0 {
        fmt.Print("-days flag required")
        help()
    }

    dumpTranscoder := false
    if *dump != "no" {
        dumpTranscoder = true
    }

    minioEndpoint  := env.Get("MINIO_ENDPOINT")
    minioAccessID  := env.Get("MINIO_ID")
    minioAccessKey := env.Get("MINIO_KEY")
    minioUseSSL    := env.GetBool("MINIO_USESSL")
    minioBucket    := env.Get("MINIO_BUCKET")

    if minioEndpoint == "" || minioAccessID == "" || minioAccessKey == "" || minioBucket == "" {
        help()
        os.Exit(1)
    }

    cfg := s3.Config{minioEndpoint, minioAccessID, minioAccessKey, minioUseSSL}
    mc, err := minio.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // List all objects in Minio
    objs, err := mc.List(minioBucket)
    if err != nil {
        log.Fatal(err)
    }

    // Filter out objects that contain the renamePrefix
    // so we don't transcode videos that have been transcoded
    var transcodedObjs []string
    var transCandids s3.Objects
    for _, obj := range objs {
        k := obj.Key
        if i := strings.Index(k, renamePrefix); i != -1 {
            strippedKey := k[:i] + k[i+len(renamePrefix):]
            transcodedObjs = append(transcodedObjs, strippedKey)
        } else {
            transCandids = append(transCandids, obj)
        }
    }

    worker := worker.NewWorker(*workers)

    // Don't transcode those older than what we specified via '-days' flag
    unixTimeLowerBound := time.Now().Unix() - int64(*backInDays * 24.0 * 60.0 * 60.0)
    sort.Strings(transcodedObjs)
    for _, obj := range transCandids {
        k := obj.Key
        if obj.LastModifiedUnix > unixTimeLowerBound {
            if !keyExists(transcodedObjs, k) {
                ret := worker.Spawn(func () {
                    log.Printf("Processing video '%s' in bucket '%s'\n", k, minioBucket)
                    process(&mc, minioBucket, k, dumpTranscoder)
                })

                if ret < 0 {
                    break
                }
            }
        }
    }

    worker.WaitAll()
}

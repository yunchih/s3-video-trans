package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "sort"
    "strconv"
    "strings"
    "time"

    "github.com/yunchih/s3-video-trans/pkg/minio"
    "github.com/yunchih/s3-video-trans/pkg/s3"
)

var helpMsg = `
  Usage of minio-video-transcoder:
      ./minio-video-transcoder -days=[N]

      Required commandline arguments:
          -days=N: Transcode video objects whose last modification time is
                   N days until now

      Required environment variables:
          MINIO_ENDPOINT: The URL of minio server (with port)
          MINIO_ID: Your Minio user ID
          MINIO_KEY: Your Minio user password
          MINIO_USESSL: Whether or not to connect with SSL enabled ("yes" or "no")
          MINIO_BUCKET: Name of target bucket
`

func process (mc *minio.Minio, minioBucket string, objKey string) {
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

    transcoded := renamePrefix + objKey
    log.Printf("Loading transcoded object to '%s'\n", transcoded)
    if err := mc.Upload(minioBucket, transcoded, tmpDst); err != nil {
        log.Fatal(err)
    }

    tmpSrc.Close()
    tmpDst.Close()

    log.Print(transcoderOutput)
}

func help() {
    fmt.Print("\n")
    fmt.Print(helpMsg)
    fmt.Print("\n")
    os.Exit(1)
}

func getEnv(key string) string {
    value := os.Getenv(key)
    if len(value) == 0 {
        fmt.Printf("Environment variable '%s' not set", key)
        help()
    }
    return value
}

func getEnvBool(key string) bool {
    if getEnv(key) == "yes" {
        return true
    }
    return false
}

func getEnvInt(key string) int {
    val, err := strconv.Atoi(getEnv(key))
    if err != nil {
        fmt.Printf("Error while setting env '%s': ", key)
        fmt.Println(err)
        help()
    }
    return val
}

func main () {
    backInDays := flag.Float64("days", -1.0, "The range in days that we look back to search for possible targets")
    flag.Parse()

    if *backInDays < 0.0 {
        fmt.Print("-days flag required")
        help()
    }

    minioEndpoint  := getEnv("MINIO_ENDPOINT")
    minioAccessID  := getEnv("MINIO_ID")
    minioAccessKey := getEnv("MINIO_KEY")
    minioUseSSL    := getEnvBool("MINIO_USESSL")
    minioBucket    := getEnv("MINIO_BUCKET")

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
        log.Println(k, strings.Index(k, renamePrefix))
        if i := strings.Index(k, renamePrefix); i != -1 {
            strippedKey := k[:i] + k[i+len(renamePrefix):]
            transcodedObjs = append(transcodedObjs, strippedKey)
        } else {
            transCandids = append(transCandids, obj)
        }
    }

    // Don't transcode those older than what we specified via '-days' flag
    unixTimeLowerBound := time.Now().Unix() - int64(*backInDays * 24.0 * 60.0 * 60.0)
    sort.Strings(transcodedObjs)
    for _, obj := range transCandids {
        k := obj.Key
        if obj.LastModifiedUnix > unixTimeLowerBound {
            if !keyExists(transcodedObjs, k) {
                log.Printf("Processing video '%s' in bucket '%s'\n", k, minioBucket)
                process(&mc, minioBucket, k)
            }
        }
    }

}

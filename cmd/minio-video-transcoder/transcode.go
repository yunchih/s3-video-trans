package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "os/exec"
    "sort"
    "strings"
)

const (
    renamePrefix   = "trans_/"
    tempFilePrefix = "transcoder-"
    transcoderCommand = "ffmpeg"
    transcoderArgs = "-y -i %s -codec:v libx264 " +
                     "-preset:v slow -tune:v stillimage " +
                     "-pix_fmt yuv420p -crf 23 -g 60 " +
                     "-keyint_min 30 -bf 3 -b-pyramid strict " +
                     "-x264opts open_gop=0:ref=3:fast_pskip=1:no_dct_decimate:no-cabac " +
                     "-vsync cfr -strict -2 -codec:a aac -b:a 128k -ac 2 -f mp4 %s"
)

func getTempFile () *os.File {
    tmp, err := ioutil.TempFile("", tempFilePrefix)
    if err != nil {
        log.Fatal(err)
    }

    return tmp
}

// Find a string in the list, assuming it's sorted
func keyExists (list []string, x string) bool {
    i := sort.SearchStrings(list, x)
    return i != len(list) && list[i] == x
}

func runTranscoder (sourceFile string, targetFile string) string {
    args := fmt.Sprintf(transcoderArgs, sourceFile, targetFile)
    _args := strings.Fields(args)
    out, err := exec.Command(transcoderCommand, _args...).CombinedOutput()
    if err != nil {
        log.Fatal(transcoderCommand, err, string(out))
    }

    return string(out)
}

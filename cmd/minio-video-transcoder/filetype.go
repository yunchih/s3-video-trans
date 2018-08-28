package main

import (
    "log"
	"os"

    "gopkg.in/h2non/filetype.v1"
)

func isVideo(filename string) bool {
    file, _ := os.Open(filename)
    defer file.Close()

    // First 261 bytes suffice
    header := make([]byte, 261)
    if _, err := file.Read(header); err != nil {
		log.Fatal(err)
    }

    return filetype.IsVideo(header)
}

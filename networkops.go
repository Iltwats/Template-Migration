package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Method to download and save patch file
func savePatchFile(urls []string, tag string) bool {
	fmt.Println("Downloading Patch files...")
	var fileLen = 0
	for i, url := range urls {
		i = len(urls) - 1 - i
		name := fmt.Sprintf("%s-%d", tag, i)
		out, _ := os.Create(name + ".patch")
		// timeout if it takes more than 10 secs
		client := http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(url + ".patch")
		if err != nil {
			log.Fatalln("Timeout", err.Error())
		}
		_, _ = io.Copy(out, resp.Body)
		fileLen++
		resp.Body.Close()
		out.Close()
	}
	fmt.Println("Download complete for all patch files.")
	return fileLen == len(urls)
}

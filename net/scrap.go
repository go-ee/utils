package net

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-ee/utils/eio"
)

func DownloadFromURL(url string, to string) (name string, err error) {
	tokens := strings.Split(url, "/")
	name = tokens[len(tokens)-1]
	fileName := fmt.Sprintf("%v/%v", to, name)

	if !eio.FileExists(fileName) {
		log.Println("Downloading", url, "to", fileName)
		var output *os.File
		if output, err = os.Create(fileName); err != nil {
			log.Println("Error while creating", fileName, "-", err)
			return
		}
		defer output.Close()

		var response *http.Response
		if response, err = http.Get(url); err != nil {
			log.Println("Error while downloading", url, "-", err)
			return
		}
		defer response.Body.Close()

		var bytes int64
		if bytes, err = io.Copy(output, response.Body); err != nil {
			log.Println("Error while downloading", url, "-", err)
			return
		}
		log.Println(name, bytes, "bytes downloaded.")
	}
	return
}

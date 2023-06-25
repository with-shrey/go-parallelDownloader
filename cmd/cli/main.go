package main

import (
	"flag"
	"github.com/with-shrey/go-parallelDownloader"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	config := parallelDownloader.Config{
		MaxNumParallelDownloads: 0,
		ChunkSize:               0,
		Retries:                 0,
	}
	var urlPath string
	var timeoutInMinutes int
	var printUsage bool
	flag.StringVar(&urlPath, "url", "", "url path to download from")
	flag.Int64Var(&config.MaxNumParallelDownloads, "maxParallelDownloads", 0, "maximum number of parts to be downloaded in parallel")
	flag.IntVar(&config.Retries, "retry", 5, "number of retries to perform for failed request")
	flag.IntVar(&timeoutInMinutes, "timeoutInMinutes", 10, "request timeout in minutes")
	flag.BoolVar(&printUsage, "help", false, "print help")

	flag.Parse()
	if printUsage {
		flag.Usage()
	}
	client := http.Client{
		Timeout: time.Duration(timeoutInMinutes) * time.Minute,
	}
	filePath, err :=
		parallelDownloader.New(
			client,
			config).Download(
			urlPath,
		)
	if err != nil {
		log.Printf("Error downloading file %s", err)
		os.Exit(1)
	}
	log.Printf("File downloaded at %s", filePath)
	os.Exit(0)
}

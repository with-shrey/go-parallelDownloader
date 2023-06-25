package parallelDownloader

import (
	"context"
	"fmt"
	"github.com/with-shrey/go-parallelDownloader/internal/fs"
	"github.com/with-shrey/go-parallelDownloader/internal/http"
	"math"
	netHttp "net/http"
	"net/url"
	"sync"
)

type Config struct {
	MaxNumParallelDownloads int64
	ChunkSize               int64
	Retries                 int
}

var (
	ErrMaxNumParallelDownloadsPositive = fmt.Errorf("max parallel downloads should be positive")
	ErrChunkSizePositive               = fmt.Errorf("chuck size should be positive")
	ErrRetriesPositive                 = fmt.Errorf("max parallel downloads should be positive")
)

const (
	DefaultRetriesCount = 5
	DefaultChunkSize    = int64(10 * 1024 * 1024) // 1 MB
)

func validateConfig(conf *Config) error {
	if conf.MaxNumParallelDownloads < 0 {
		return ErrMaxNumParallelDownloadsPositive
	}
	if conf.ChunkSize < 0 {
		return ErrChunkSizePositive
	}
	if conf.Retries < 0 {
		return ErrRetriesPositive
	}

	if conf.ChunkSize == 0 {
		conf.ChunkSize = DefaultChunkSize
	}
	if conf.Retries == 0 {
		conf.Retries = DefaultRetriesCount
	}
	return nil
}

func (d downloader) downloadPart(ctx context.Context, urlPath string, stat http.FileStat, start int64, end int64) error {
	reader, err := d.httpDownloader.GetFilePart(ctx, urlPath, start, end, d.Config.Retries)
	if err != nil {
		return err
	}
	defer reader.Close()
	err = fs.WriteToFile(stat.FileName, start, reader)
	if err != nil {
		return err
	}
	return nil
}

func (d downloader) downloadParts(urlPath string, stat http.FileStat) error {
	var maxProcesses = d.Config.MaxNumParallelDownloads
	if d.Config.MaxNumParallelDownloads == 0 {
		maxProcesses = int64(math.Ceil(float64(stat.Length) / float64(d.Config.ChunkSize)))
	}

	guard := make(chan bool, maxProcesses)
	errChan := make(chan error, 1)

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	var start int64 = 0
	for ; start < stat.Length; start += d.Config.ChunkSize {
		end := start + d.Config.ChunkSize
		if start+d.Config.ChunkSize > stat.Length {
			end = stat.Length
		}
		start := start
		wg.Add(1)
		go func() {
			defer wg.Done()
			guard <- true
			defer func() {
				<-guard
			}()
			if ctx.Err() == context.Canceled {
				return
			}
			err := d.downloadPart(ctx, urlPath, stat, start, end)
			if err != nil {
				errChan <- err
				return
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	err := <-errChan
	cancel()
	return err
}

type downloader struct {
	Client         netHttp.Client
	httpDownloader http.DownloaderHTTP
	Config         Config
}

func New(client netHttp.Client, config Config) *downloader {
	return &downloader{
		Client:         client,
		Config:         config,
		httpDownloader: http.DownloaderHTTP{Client: client},
	}
}

// Download file in parts using config provided
func (d downloader) Download(urlPath string) (string, error) {
	err := validateConfig(&d.Config)
	if _, err := url.ParseRequestURI(urlPath); err != nil {
		return "", fmt.Errorf("url should be valid")
	}

	if err != nil {
		return "", err
	}
	fileStat, err := d.httpDownloader.GetFileStats(urlPath)
	if err != nil {
		return "", err
	}
	err = fs.CreateEmptyFile(fileStat.FileName, fileStat.Length)
	if err != nil {
		return "", err
	}
	err = d.downloadParts(urlPath, fileStat)
	if err != nil {
		fs.DeleteFile(fileStat.FileName)
		return "", err
	}
	return fileStat.FileName, nil
}

package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
)

type FileStat struct {
	Length   int64
	Hash     string
	FileName string
}

type DownloaderHTTP struct {
	Client http.Client
}

func (d DownloaderHTTP) GetFileStats(url string) (FileStat, error) {
	request, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return FileStat{}, err
	}
	response, err := d.Client.Do(request)
	if err != nil {
		return FileStat{}, err
	}
	stat := FileStat{
		Length:   response.ContentLength,
		Hash:     response.Header.Get("ETag"),
		FileName: path.Base(request.URL.Path),
	}
	return stat, nil
}

func (d DownloaderHTTP) GetFilePart(ctx context.Context, url string, start int64, end int64, retries int) (io.ReadCloser, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	response, err := d.Client.Do(request)
	if (err != nil || response.StatusCode >= 300) && retries > 0 {
		return d.GetFilePart(ctx, url, start, end, retries-1)
	}
	if err != nil {
		return nil, err
	}

	return response.Body, nil
}

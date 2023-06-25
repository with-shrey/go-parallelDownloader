package http

import (
	"context"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"io"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGetFileStats(t *testing.T) {
	length := int64(math.Abs(gofakeit.Float64()))
	etag := gofakeit.Word()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.FormatInt(length, 10))
		w.Header().Set("ETag", etag)
		fmt.Fprintln(w, `Some fake test`)
	}))
	defer ts.Close()
	downloaderHTTP := DownloaderHTTP{}
	stat, err := downloaderHTTP.GetFileStats(ts.URL + "/testfile.txt")
	require.NoError(t, err)
	require.Equal(t, FileStat{
		FileName: "testfile.txt",
		Length:   length,
		Hash:     etag,
	}, stat)
}

func TestGetFilePart(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")
		rangeHeader = strings.Replace(rangeHeader, "bytes=", "", -1)
		ranges := strings.Split(rangeHeader, "-")
		var nums []int
		for _, s := range ranges {
			num, err := strconv.Atoi(s)
			if err != nil {
				continue
			}
			nums = append(nums, num)
		}
		text := `Some faketest Some fake test Some fake test`
		fmt.Fprint(w, text[nums[0]:nums[1]])
	}))
	defer ts.Close()
	downloaderHTTP := DownloaderHTTP{Client: http.Client{
		Timeout: 0,
	}}
	reader, err := downloaderHTTP.GetFilePart(context.Background(), ts.URL+"/testfile.txt", 6, 10, 0)
	require.NoError(t, err)
	bytes, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, "aket", string(bytes))
}

func TestGetFilePart_SuccessfulAfterRetry(t *testing.T) {
	retryCounter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if retryCounter == 1 {
			w.WriteHeader(200)
			fmt.Fprint(w, "success")
		} else {
			w.WriteHeader(404)
			fmt.Fprint(w, "fail")
		}
		retryCounter += 1
	}))
	defer ts.Close()
	downloaderHTTP := DownloaderHTTP{Client: http.Client{
		Timeout: 0,
	}}
	reader, err := downloaderHTTP.GetFilePart(context.Background(), ts.URL+"/testfile.txt", 6, 10, 5)
	require.NoError(t, err)
	bytes, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, retryCounter, 2)
	require.Equal(t, "success", string(bytes))
}

func TestGetFilePart_FailAfterRetry(t *testing.T) {
	retryCounter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprint(w, "fail")
		retryCounter += 1
	}))
	defer ts.Close()
	downloaderHTTP := DownloaderHTTP{Client: http.Client{
		Timeout: 0,
	}}
	reader, err := downloaderHTTP.GetFilePart(context.Background(), ts.URL+"/testfile.txt", 6, 10, 2)
	require.NoError(t, err)
	bytes, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, retryCounter, 3)
	require.Equal(t, "fail", string(bytes))
}

func TestGetFilePart_ConnectFailure(t *testing.T) {
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			// Simulate network error by returning a custom error
			return nil, fmt.Errorf("simulated network error")
		},
	}

	client := http.Client{
		Transport: transport,
		Timeout:   time.Second * 10,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	}))
	defer ts.Close()
	downloaderHTTP := DownloaderHTTP{Client: client}
	reader, err := downloaderHTTP.GetFilePart(context.Background(), ts.URL+"/testfile.txt", 6, 10, 2)
	require.ErrorContains(t, err, "simulated network error")
	require.Equal(t, nil, reader)
}

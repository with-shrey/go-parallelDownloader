package parallelDownloader

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestValidateConfig_ShouldReturnNoError(t *testing.T) {
	err := validateConfig(&Config{
		MaxNumParallelDownloads: 10,
		Retries:                 10,
		ChunkSize:               100,
	})
	require.NoError(t, err)
}

func TestValidateConfig_ShouldReturnErrorIfMaxNumParallelDownloadsIsNegative(t *testing.T) {
	err := validateConfig(&Config{
		MaxNumParallelDownloads: -10,
		Retries:                 10,
		ChunkSize:               100,
	})
	require.ErrorIs(t, ErrMaxNumParallelDownloadsPositive, err)
}

func TestValidateConfig_ShouldReturnErrorIfeRetriesIsNegative(t *testing.T) {
	err := validateConfig(&Config{
		MaxNumParallelDownloads: 10,
		Retries:                 -10,
		ChunkSize:               100,
	})
	require.ErrorIs(t, ErrRetriesPositive, err)
}

func TestValidateConfig_ShouldReturnErrorIfChunkSizeIsNegative(t *testing.T) {
	err := validateConfig(&Config{
		MaxNumParallelDownloads: 10,
		Retries:                 10,
		ChunkSize:               -100,
	})
	require.ErrorIs(t, ErrChunkSizePositive, err)
}

func TestValidateConfig_ShouldSetDefaults(t *testing.T) {
	config := Config{}
	err := validateConfig(&config)
	require.NoError(t, err)
	require.Equal(t, DefaultChunkSize, config.ChunkSize)
	require.Equal(t, DefaultRetriesCount, config.Retries)
	require.Equal(t, int64(0), config.MaxNumParallelDownloads)
}

func TestDownloader_Download(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		text := `Some faketest Some fake test Some fake test`
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
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
			resp := text[nums[0]:nums[1]]
			fmt.Fprint(w, resp)
		} else {
			fmt.Fprint(w, text)
		}
	}))
	defer ts.Close()

	downloader := New(http.Client{}, Config{
		MaxNumParallelDownloads: 0,
		Retries:                 1,
		ChunkSize:               10,
	})
	_, err := downloader.Download(ts.URL + "/testfile.txt")
	require.NoError(t, err)

	bytesRead, err := os.ReadFile("testfile.txt")
	require.NoError(t, err)
	text := string(bytesRead)
	require.Equal(t, `Some faketest Some fake test Some fake test`, text)
}

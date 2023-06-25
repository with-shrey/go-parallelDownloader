package fs

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"io"
	"math"
	"os"
	"path"
	"strings"
	"testing"
)

func TestCreateEmptyFile(t *testing.T) {
	size := gofakeit.Int64()
	size = int64(math.Abs(float64(size)))
	filePath := path.Join(t.TempDir(), "file.txt")
	err := CreateEmptyFile(filePath, size)
	require.NoError(t, err)
	stat, err := os.Stat(filePath)
	require.NoError(t, err)
	require.Equal(t, "file.txt", stat.Name())
	require.Equal(t, int64(0), stat.Size())
}

func TestGetFileReader(t *testing.T) {
	filePath := path.Join(t.TempDir(), "file.txt")
	file, err := os.Create(filePath)
	require.NoError(t, err)
	testString := "Hello i am temp file"
	b := []byte(testString)
	_, err = file.Write(b)
	require.NoError(t, err)

	reader, err := GetFileReader(filePath)
	require.NoError(t, err)
	bytesRead, err := io.ReadAll(reader)
	require.NoError(t, err)

	require.Equal(t, testString, string(bytesRead))
}

func TestWriteToFile(t *testing.T) {
	filePath := path.Join(t.TempDir(), "file.txt")
	testString := "Hello i am temp file"
	err := CreateEmptyFile(filePath, 50)
	require.NoError(t, err)

	err = WriteToFile(filePath, 10, strings.NewReader(testString))
	require.NoError(t, err)

	reader, err := GetFileReader(filePath)
	require.NoError(t, err)
	bytesRead, err := io.ReadAll(reader)
	require.NoError(t, err)

	require.Equal(t, 10, strings.Index(string(bytesRead), testString))
}

func TestDeleteFile(t *testing.T) {
	filePath := path.Join(t.TempDir(), "file.txt")
	_, err := os.Create(filePath)
	require.NoError(t, err)

	err = DeleteFile(filePath)
	require.NoError(t, err)

	_, err = os.Stat(filePath)
	require.ErrorIs(t, err, os.ErrNotExist)
}

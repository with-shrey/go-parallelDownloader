package fs

import (
	"io"
	"os"
)

// CreateEmptyFile creates an empty file in given size
func CreateEmptyFile(path string, size int64) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Seek(size-1, io.SeekStart)
	file.Write([]byte{0})
	return nil
}

// WriteToFile write data to file at given location
func WriteToFile(path string, start int64, dataReader io.Reader) error {
	file, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, dataReader)
	if err != nil {
		return err
	}
	return nil
}

// DeleteFile delete file
func DeleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

// GetFileReader delete file
func GetFileReader(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

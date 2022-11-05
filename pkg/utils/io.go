package utils

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

// SendHTTPError simplifies the process of writing a HTTP status code an a message
// to the response writer
func SendHTTPError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(message))
}

// WriteText is an utility function that writes the given string to the given
// text file (fname)
func WriteText(text, fname string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(text)
	if err != nil {
		return err
	}
	return nil
}

// TempDir returns the path to the temporal directory
func TempDir() string {
	if IsRunningInLambdaEnv() {
		return "/tmp"
	}
	return os.TempDir()
}

// Exists checks if the given file name exists
func Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// DownloadFile downloads a file served in an HTTP file server
func DownloadFile(filepath, url string) (err error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

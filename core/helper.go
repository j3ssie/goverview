package core

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
)

// ReadingFile Reading file and return content as []string
func ReadingFile(filename string) []string {
	var result []string
	if strings.HasPrefix(filename, "~") {
		filename, _ = homedir.Expand(filename)
	}
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return result
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		val := scanner.Text()
		result = append(result, val)
	}

	if err := scanner.Err(); err != nil {
		return result
	}
	return result
}

// ReadingFileUnique Reading file and return content as []string
func ReadingFileUnique(filename string) []string {
	var result []string
	if strings.Contains(filename, "~") {
		filename, _ = homedir.Expand(filename)
	}
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return result
	}

	unique := true
	seen := make(map[string]bool)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		val := scanner.Text()
		// unique stuff
		if val == "" {
			continue
		}
		val = strings.TrimSpace(val)
		if seen[val] && unique {
			continue
		}

		if unique {
			seen[val] = true
			result = append(result, val)
		}
	}

	if err := scanner.Err(); err != nil {
		return result
	}
	return result
}

// WriteToFile write string to a file
func WriteToFile(filename string, data string) (string, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.WriteString(file, data+"\n")
	if err != nil {
		return "", err
	}
	return filename, file.Sync()
}

// Unique unique content of a file and remove blank line
func Unique(filename string) {
	data := ReadingFileUnique(filename)
	WriteToFile(filename, strings.Join(data, "\n"))
}

// AppendTo append string to a file
func AppendTo(filename string, data string) (string, error) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	if _, err := f.Write([]byte(data + "\n")); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return filename, nil
}

// FileExists check if file is exist or not
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// FolderExists check if file is exist or not
func FolderExists(foldername string) bool {
	if _, err := os.Stat(foldername); os.IsNotExist(err) {
		return false
	}
	return true
}

// GetFileNames get all file name with extension
func GetFileNames(dir string, ext string) []string {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	var files []string
	filepath.Walk(dir, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			if strings.HasSuffix(f.Name(), ext) {
				filename, _ := filepath.Abs(path)
				files = append(files, filename)
			}
		}
		return nil
	})
	return files
}

// GetTS get current timestamp and return a string
func GetTS() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

// GenHash gen SHA1 hash from string
func GenHash(text string) string {
	h := sha1.New()
	h.Write([]byte(text))
	hashed := h.Sum(nil)
	return fmt.Sprintf("%x", hashed)
}

// StripPath just strip some invalid string path
func StripPath(raw string) string {
	raw = strings.Replace(raw, "/", "_", -1)
	raw = strings.Replace(raw, " ", "_", -1)
	return raw
}

// Base64Encode just Base64 Encode
func Base64Encode(raw string) string {
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

// Base64Decode just Base64 Encode
func Base64Decode(raw string) string {
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return raw
	}
	return string(data)
}

// URLDecode decode url
func URLDecode(raw string) string {
	decodedValue, err := url.QueryUnescape(raw)
	if err != nil {
		return raw
	}
	return decodedValue
}

// URLEncode Encode query
func URLEncode(raw string) string {
	decodedValue := url.QueryEscape(raw)
	return decodedValue
}

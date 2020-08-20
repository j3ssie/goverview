package utils

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
)

// CalcTimeout calculate timeout
func CalcTimeout(raw string) int {
	raw = strings.ToLower(strings.TrimSpace(raw))
	seconds := raw
	multiply := 1

	matched, _ := regexp.MatchString(`.*[a-z]`, raw)
	if matched {
		unitTime := fmt.Sprintf("%c", raw[len(raw)-1])
		seconds = raw[:len(raw)-1]
		switch unitTime {
		case "s":
			multiply = 1
			break
		case "m":
			multiply = 60
			break
		case "h":
			multiply = 3600
			break
		}
	}

	timeout, err := strconv.Atoi(seconds)
	if err != nil {
		return 0
	}
	return timeout * multiply
}

// GetDomain get domain from the URL
func GetDomain(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err == nil {
		return u.Hostname(), nil
	}
	return raw, err
}

// EmptyDir check if directory is empty or not
func EmptyDir(dir string) bool {
	if !FolderExists(dir) {
		return true
	}
	f, err := os.Open(NormalizePath(dir))
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true
	}
	return false
}

// EmptyFile check if file is empty or not
func EmptyFile(filename string, num int) bool {
	filename = NormalizePath(filename)
	if !FileExists(filename) {
		return true
	}
	data := ReadingLines(filename)
	if len(data) > num {
		return false
	}
	return true
}

// StrToInt string to int
func StrToInt(data string) int {
	i, err := strconv.Atoi(data)
	if err != nil {
		return 0
	}
	return i
}

// GetOSEnv get environment variable
func GetOSEnv(name string, alt string) string {
	variable, ok := os.LookupEnv(name)
	if !ok {
		if alt != "" {
			return alt
		}
		return name
	}
	return variable
}

// MakeDir just make a folder
func MakeDir(folder string) {
	os.MkdirAll(folder, 0750)
}

// GetCurrentDay get current day
func GetCurrentDay() string {
	currentTime := time.Now()
	return fmt.Sprintf("%v", currentTime.Format("2006-01-02_3:4:5"))
}

// NormalizePath the path
func NormalizePath(path string) string {
	if strings.HasPrefix(path, "~") {
		path, _ = homedir.Expand(path)
	}
	return path
}

// GetFileContent Reading file and return content of it
func GetFileContent(filename string) string {
	var result string
	if strings.Contains(filename, "~") {
		filename, _ = homedir.Expand(filename)
	}
	file, err := os.Open(filename)
	if err != nil {
		return result
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return result
	}
	return string(b)
}

// ReadingLines Reading file and return content as []string
func ReadingLines(filename string) []string {
	var result []string
	if strings.HasPrefix(filename, "~") {
		filename, _ = homedir.Expand(filename)
	}
	file, err := os.Open(filename)
	if err != nil {
		return result
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		val := strings.TrimSpace(scanner.Text())
		if val == "" {
			continue
		}
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
	if err != nil {
		return result
	}
	defer file.Close()

	unique := true
	seen := make(map[string]bool)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		val := scanner.Text()
		// unique stuff
		if val == "" {
			continue
		}
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

// AppendToContent append string to a file
func AppendToContent(filename string, data string) (string, error) {
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
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// FolderExists check if file is exist or not
func FolderExists(foldername string) bool {
	if _, err := os.Stat(foldername); os.IsNotExist(err) {
		return false
	}
	return true
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

// GetFileSize get file size of a file in GB
func GetFileSize(src string) float64 {
	var sizeGB float64
	fi, err := os.Stat(NormalizePath(src))
	if err != nil {
		return sizeGB
	}
	// get the size
	size := fi.Size()
	sizeGB = float64(size) / (1024 * 1024 * 1024)
	return sizeGB
}

// RandomString return a random string with length
func RandomString(n int) string {
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var letter = []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letter[seededRand.Intn(len(letter))]
	}
	return string(b)
}

// RunCmdWithOutput just run os command
func RunCmdWithOutput(cmd string) string {
	var output string
	command := []string{
		"sh",
		"-c",
		cmd,
	}
	realCmd := exec.Command(command[0], command[1:]...)
	// output command output to std too
	cmdReader, _ := realCmd.StdoutPipe()
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			output += scanner.Text() + "\n"
		}
	}()
	if err := realCmd.Start(); err != nil {
		return output
	}
	if err := realCmd.Wait(); err != nil {
		return output
	}
	return output
}

// RunOSCommand just run os command
func RunOSCommand(cmd string) {
	command := []string{
		"sh",
		"-c",
		cmd,
	}
	realCmd := exec.Command(command[0], command[1:]...)
	// output command output to std too
	cmdReader, _ := realCmd.StdoutPipe()
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()
	if err := realCmd.Start(); err != nil {
		return
	}
	if err := realCmd.Wait(); err != nil {
		return
	}
}

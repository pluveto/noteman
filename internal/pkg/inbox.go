package pkg

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

func RemoveComment(s string) string {
	if len(s) <= 1 {
		return s
	}
	var sb strings.Builder
	in_quote := false
	in_comment := false
	n := len(s)
	for i := 0; i < n; i++ {
		if in_comment {
			if s[i] == '\n' {
				in_comment = false
				sb.WriteString(string(s[i]))
			}
			continue
		}
		if s[i] == '/' && i+1 < n && s[i+1] == '/' && !in_quote {
			in_comment = true
			continue
		}
		if s[i] == '"' && i-1 >= 0 && s[i-1] != '\\' {
			in_quote = !in_quote
		}
		sb.WriteString(string(s[i]))
	}

	return sb.String()
}

type PathSearcher struct {
	SearchPaths []string
	SearchNames []string
}

func NewPathSearcher() *PathSearcher {
	return &PathSearcher{
		SearchPaths: []string{},
		SearchNames: []string{},
	}
}

// AddSearchPaths adds search paths to the conf loader.
func (c *PathSearcher) AddSearchPaths(paths ...string) *PathSearcher {
	c.SearchPaths = append(c.SearchPaths, paths...)
	return c
}

// AddSearchNames adds search names to the conf loader.
func (c *PathSearcher) AddSearchNames(names ...string) *PathSearcher {
	c.SearchNames = append(c.SearchNames, names...)
	return c
}

// Execute searches for the file in the search paths.
func (c *PathSearcher) Execute() (string, error) {
	for _, path := range c.SearchPaths {
		for _, name := range c.SearchNames {
			full := path + "/" + name
			logrus.Debugln("searching for", full)
			if _, err := os.Stat(full); err == nil {
				return full, nil
			}
		}
	}
	return "", errors.New("file not found")
}

func MustJsonEncode(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func RemoveDupSortedSorted(s *[]string) {
	var ss []string
	for i := 0; i < len(*s); i++ {
		if i-1 >= 0 && (*s)[i] == (*s)[i-1] {
			continue
		}
		ss = append(ss, (*s)[i])
	}
	s = &ss
}

// simpleGlob glob and append file to given slice
func SimpleGlob(dir string, files *[]string) error {
	if files == nil {
		return fmt.Errorf("must give a receive slice for files")
	}
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		logrus.Errorln(err)
		return err
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			SimpleGlob(fileInfo.Name(), files)
		} else {
			*files = append(*files, path.Join(dir, fileInfo.Name()))
		}
	}
	return nil
}

func SimplifyPath(path string) string {
	stack := []string{}
	for _, name := range strings.Split(path, "/") {
		if name == ".." {
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		} else if name != "" && name != "." {
			stack = append(stack, name)
		}
	}
	return strings.Join(stack, "/")
}

func ReplaceDotPath(path string, current string) string {
	if !strings.HasPrefix(path, "/") {
		sim := SimplifyPath(path)
		return current + "/" + sim
	}

	return path
}

func WaitForEnter() {
	fmt.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func QuotePath(path string) string {
	return "\"" + path + "\""
}

func IsPureASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}

func FileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	hash := md5.New()
	_, _ = io.Copy(hash, file)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func PostFile(filefield string, filename string, target_url string, headers map[string]string) (*http.Response, error) {
	body_buf := bytes.NewBufferString("")
	body_writer := multipart.NewWriter(body_buf)

	// use the body_writer to write the Part headers to the buffer
	_, err := body_writer.CreateFormFile(filefield, filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return nil, err
	}

	// the file data will be the second part of the body
	fh, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file")
		return nil, err
	}
	// need to know the boundary to properly close the part myself.
	boundary := body_writer.Boundary()
	//close_string := fmt.Sprintf("\r\n--%s--\r\n", boundary)
	close_buf := bytes.NewBufferString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	// use multi-reader to defer the reading of the file data until
	// writing to the socket buffer.
	request_reader := io.MultiReader(body_buf, fh, close_buf)
	fi, err := fh.Stat()
	if err != nil {
		fmt.Printf("Error Stating file: %s", filename)
		return nil, err
	}
	req, err := http.NewRequest("POST", target_url, request_reader)
	if err != nil {
		return nil, err
	}

	// Set headers for multipart, and Content Length
	req.Header.Add("Content-Type", "multipart/form-data; boundary="+boundary)
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	req.ContentLength = fi.Size() + int64(body_buf.Len()) + int64(close_buf.Len())

	return http.DefaultClient.Do(req)
}

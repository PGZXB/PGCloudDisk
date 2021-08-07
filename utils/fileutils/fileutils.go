package fileutils

import (
	"net/http"
	"os"
	"path"
)

var httpContentTypeOfFileSuffix map[string]string = map[string]string{
	".html": "text/html",
	".txt":  "text/plain;charset=UTF-8",
	".xml":  "text/xml",
	".gif":  "image/gif",
	".jpg":  "image/jpeg",
	".png":  "image/png",
	".mp4":  "video/mpeg4",
	".ico":  "application/octet-stream",
}

func IsDir(fileAddr string) bool {
	s, err := os.Stat(fileAddr)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func GetHttpContentTypeOfFile(file *os.File) (string, error) {

	buffer := make([]byte, 512)

	_, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

func GetHttpContentTypeByFilename(filename string) string {
	suffix := path.Ext(filename)
	if res, ok := httpContentTypeOfFileSuffix[suffix]; ok {
		return res
	}
	return "application/octet-stream"
}

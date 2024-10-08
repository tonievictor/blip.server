package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/tonie-ng/blip/request"
)

const (
	Host                = "Host"
	ContentType         = "Content-Type"
	ContentLength       = "Content-Length"
	UserAgent           = "User-Agent"
	Accept              = "Accept"
	NotFound            = "404 Not Found"
	BadRequest          = "400 Bad Request"
	NotAllowed          = "405 Method Not Allowed"
	InternalServerError = "500 Internal Server Error"
	Ok                  = "200 OK"
	Get                 = "GET"
	Post                = "POST"
	Head                = "HEAD"
)

func main() {
	ln, err := net.Listen("tcp", ":6703")
	if err != nil {
		slog.Error("listen", "error", err)
		return
	}

	slog.Info("Listening on port 6703")
	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("accept", "error", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleGetReq(conn net.Conn, req request.Request) {
	fileInfo, filePath, err := findFile(req.Path)
	if err != nil {
		slog.Info("An error occured opening the file", "error", err)
		WriteHeader(conn, "", NotFound, 0)
		return
	}

	data, _ := os.ReadFile(filePath)
	WriteHeader(conn, fileInfo.Name(), Ok, fileInfo.Size())
	if req.Method != Head {
		conn.Write(data)
	}
	return
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	req, err := request.ParseRequest(conn)
	if err != nil {
		return
	}

	switch req.Method {
	case Get, Head:
		handleGetReq(conn, req)
	default:
		WriteHeader(conn, "", NotAllowed, 0)
	}
}

func WriteHeader(conn net.Conn, filename, status string, contentLength int64) error {
	var res strings.Builder
	res.WriteString("HTTP/1.1 ")
	res.WriteString(status)
	res.WriteString("\r\n")
	res.WriteString("Content-Type: ")
	res.WriteString(GetContentType(filename))
	res.WriteString("\r\n")
	res.WriteString("Content-Length: ")
	res.WriteString(fmt.Sprintf("%d", contentLength))
	res.WriteString("\r\n")
	res.WriteString("Date: ")
	res.WriteString(time.Now().Format(time.RFC1123))
	res.WriteString("\r\n\r\n")

	conn.Write([]byte(res.String()))
	return nil
}

func GetContentType(filename string) string {
	ext := strings.Split(filename, ".")
	switch ext[len(ext)-1] {
	case "js":
		return "application/javascript"
	case "jpg":
		return "image/jpg"
	case "png":
		return "image/png"
	case "html":
		return "text/html"
	default:
		return "text/plain"
	}
}

func findFile(filename string) (os.FileInfo, string, error) {
	if filename == "/" {
		filename = "index.html"
	}
	if filename[0] == '/' {
		filename = filename[1:]
	}

	fileInfo, err := os.Stat(filename)
	if err != nil {
		return nil, "", err
	}

	if fileInfo.IsDir() {
		if filename[len(filename)-1] == '/' {
			filename = fmt.Sprintf("%sindex.html", filename)
		} else {
			filename = fmt.Sprintf("%s/index.html", filename)
		}
		fileInfo, err = os.Stat(filename)
		if err != nil {
			return nil, "", err
		}
	}
	return fileInfo, filename, nil
}

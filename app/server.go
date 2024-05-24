package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type Request struct {
	Path    string
	Headers map[string]string
	Method  string
	Body    string
}

type Server struct {
	dir string
}

func ParseRequest(request string) *Request {
	path := strings.Join(GetPath(request), "/")
	headers := GetHeaders(request)
	method := GetMethod(request)
	body := GetBody(request)

	return &Request{
		Path:    path,
		Headers: headers,
		Method:  method,
		Body:    body,
	}
}

func NewServer(dir string) *Server {
	return &Server{
		dir: dir,
	}
}
func GetMethod(request string) string {
	method := strings.Split(request, " ")[0]
	return method
}
func GetPath(request string) []string {
	path := strings.Split(request, " ")
	path = strings.Split(path[1], "/")
	return path[1:]
}

func GetBody(request string) string {
	parts := strings.Split(request, "\r\n\r\n")
	body := parts[len(parts)-1]
	return body
}

func GetHeaders(request string) map[string]string {
	parts := strings.Split(request, "\r\n\r\n")
	headers := strings.Split(parts[0], "\r\n")[1:]
	headersMap := make(map[string]string)
	for _, header := range headers {
		key := strings.Split(header, ":")[0]
		value := strings.Split(header, ":")[1]
		headersMap[key] = strings.TrimSpace(value)
	}
	return headersMap
}

func handleFiles(dir string, filename string) ([]byte, error) {
	slog.Info(fmt.Sprintf("Handle Files: %v%v", dir, filename))
	files, err := os.ReadDir(dir)
	if err != nil {
		return []byte{}, errors.New("dir doesnt exists")
	}
	fmt.Println(files)
	for _, file := range files {
		if file.Name() == filename {
			bytes, err := os.ReadFile(dir + "/" + filename)
			if err != nil {
				return []byte{}, errors.New("could not open file")
			}
			return bytes, nil
		}
	}
	return []byte{}, errors.New("file not found")
}

func (s *Server) handleConnection(conn net.Conn) {
	buffer := make([]byte, 1024)

	conn.Read(buffer)

	req := ParseRequest(string(buffer))
	slog.Info(fmt.Sprintf("Request: %v /%v", req.Method, req.Path))
	path := strings.Split(req.Path, "/")
	headers := GetHeaders(string(buffer))
	switch path[0] {
	case "":
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	case "user-agent":
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(headers["User-Agent"]), headers["User-Agent"])))
	case "echo":
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(path[1]), path[1])))
	case "files":
		file, err := handleFiles(s.dir, path[1])
		if err != nil {
			if err.Error() == "file not found" {
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return
			}
			conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\nCould not read file"))
			return
		}
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %v\r\n\r\n%v\r\n\r\n", len(file), string(file))))
		return
	default:
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func main() {
	dir := flag.String("directory", "files", "directory")
	flag.Parse()

	s := NewServer(*dir)

	l, err := net.Listen("tcp", "localhost:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go s.handleConnection(conn)
	}

}

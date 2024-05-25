package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Request struct {
	Path    string
	Headers map[string]string
	Method  string
	Body    string
}

type Response struct {
	Status  string
	Headers map[string]string
	Body    string
}

func (r *Response) Deserialize() []byte {
	var resp string
	resp += "HTTP/1.1 "
	resp += r.Status + "\r\n"
	for key, value := range r.Headers {
		// fmt.Println(key)
		// fmt.Println(value)
		resp += fmt.Sprintf("%v: %v\r\n", key, value)
	}
	resp += "\r\n" + r.Body
	return []byte(resp)
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

func writeFile(dir, filename string, body string) error {
	fmt.Println(body)
	// fmt.Println([]byte(body))
	slog.Info(fmt.Sprintf("writing file: %v%v", dir, filename))
	file, err := os.Create(dir + "/" + filename)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Println(len("grape raspberry orange apple apple grape banana banana"))
	fmt.Println(len(strings.TrimRight(body, "\x00")))
	_, err = file.WriteString(strings.TrimRight(body, "\x00"))
	if err != nil {
		return err
	}

	return nil
}

func readFile(dir string, filename string) ([]byte, error) {
	slog.Info(fmt.Sprintf("reading file: %v%v", dir, filename))
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
	defer conn.Close()

	req := ParseRequest(string(buffer))
	slog.Info(fmt.Sprintf("Request: %v /%v", req.Method, req.Path))
	path := strings.Split(req.Path, "/")
	// headers := GetHeaders(string(buffer))

	switch path[0] {
	case "":
		resp := Response{Status: "200 OK", Body: "", Headers: map[string]string{}}
		conn.Write(resp.Deserialize())
		return
		// conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	case "user-agent":
		resp := Response{Status: "200 OK", Body: req.Headers["User-Agent"], Headers: map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(req.Headers["User-Agent"])),
		}}
		conn.Write(resp.Deserialize())
		return
		// conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(headers["User-Agent"]), headers["User-Agent"])))
	case "echo":
		responseHeaders := map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(path[1])),
		}

		if v, ok := req.Headers["Accept-Encoding"]; ok {
			if v == "gzip" {
				responseHeaders["Content-Encoding"] = "gzip"
			}
		}
		resp := Response{Status: "200 OK", Body: path[1], Headers: responseHeaders}
		conn.Write(resp.Deserialize())
		return
		// conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(path[1]), path[1])))
	case "files":
		if req.Method == "GET" {
			file, err := readFile(s.dir, path[1])
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
		} else if req.Method == "POST" {
			err := writeFile(s.dir, path[1], req.Body)
			if err != nil {
				fmt.Println("could not write file", err)
				conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\nCould not write file"))
				return
			}
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 201 Created\r\n\r\n")))
			return
		}

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

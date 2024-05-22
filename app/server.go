package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func GetPath(request string) []string {
	path := strings.Split(request, " ")
	path = strings.Split(path[1], "/")
	return path[1:]
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
	// fmt.Println(headersMap)
	return headersMap
}

func main() {
	l, err := net.Listen("tcp", "localhost:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	buffer := make([]byte, 1024)

	conn.Read(buffer)

	path := GetPath(string(buffer))
	headers := GetHeaders(string(buffer))
	// fmt.Println(path)
	// fmt.Println(len(path))
	switch path[0] {
	case "":
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	case "user-agent":
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(headers["User-Agent"]), headers["User-Agent"])))
	case "echo":
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(path[1]), path[1])))
	default:
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

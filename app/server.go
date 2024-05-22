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
	return path
}

func main() {
	fmt.Println("Logs from your program will appear here!")

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
	fmt.Println(path)
	fmt.Println(len(path))
	if len(path) == 2 {
		fmt.Println("111")
		if path[1] != "" {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		} else {
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		}
	} else if path[1] == "echo" {
		fmt.Println("222")
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(path[2]), path[2])))
	} else {
		// conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		fmt.Println("404")
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
	// if string(buffer).split()
}

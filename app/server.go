package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func GetPath(request string) string {
	path := strings.Split(request, " ")
	paths := strings.Split(path[1], "/")
	fmt.Println(paths)
	return "asdf"
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

	fmt.Println(string(buffer))
	_ = GetPath(string(buffer))
	if strings.Split(string(buffer), " ")[1] == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
	// if string(buffer).split()
}

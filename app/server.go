package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}

}
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read the request
	reader := bufio.NewReader(conn)
	req, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Failed to read the request:", err)
		return
	}

	fmt.Printf("Request: %s\n", req)

	// Parse the request
	requestLine := strings.Fields(req)
	if len(requestLine) < 2 {
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}

	method := requestLine[0]
	path := requestLine[1]

	// Read headers
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// Handle different paths and methods
	switch method {
	case "GET":
		handleGetRequest(conn, path, headers)
	default:
		conn.Write([]byte("HTTP/1.1 405 Method Not Allowed\r\n\r\n"))
	}
}

func handleGetRequest(conn net.Conn, path string, headers map[string]string) {
	if path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nWelcome to the homepage!"))
	} else if strings.HasPrefix(path, "/echo/") {
		message := strings.TrimPrefix(path, "/echo/")
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)
		conn.Write([]byte(response))
	} else if strings.HasPrefix(path, "/user-agent") {
		userAgent := headers["User-Agent"]
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
		conn.Write([]byte(response))
	} else if strings.HasPrefix(path, "/files/") {
		dir := os.Args[2]
		fileName := strings.TrimPrefix(path, "/files/")
		data, err := os.ReadFile(dir + fileName)
		response := ""
		if err != nil {
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
		} else {
			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(data), data)
		}
		conn.Write([]byte(response))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

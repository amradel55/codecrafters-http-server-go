package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	// Logging for debugging
	fmt.Println("Starting server...")
	fmt.Print(os.Args)
	var dir string
	if len(os.Args) < 2 {
		// Use a default directory if no arguments are provided
		dir = "./default_directory"
		fmt.Println("No directory argument provided. Using default directory:", dir)
	} else {
		dir = os.Args[2]
	}

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221:", err)
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Server is listening on port 4221")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			os.Exit(1)
		}
		go handleConnection(conn, dir)
	}
}

func handleConnection(conn net.Conn, dir string) {
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
		handleGetRequest(conn, path, headers, dir)
	case "POST":
		handlePostRequest(conn, path, headers, reader, dir)
	default:
		conn.Write([]byte("HTTP/1.1 405 Method Not Allowed\r\n\r\n"))
	}
}

func handleGetRequest(conn net.Conn, path string, headers map[string]string, dir string) {
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
		fileName := strings.TrimPrefix(path, "/files/")
		data, err := os.ReadFile(filepath.Join(dir, fileName))
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

func handlePostRequest(conn net.Conn, path string, headers map[string]string, reader *bufio.Reader, dir string) {
	if !strings.HasPrefix(path, "/files/") {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return
	}

	contentLengthStr := headers["Content-Length"]
	if contentLengthStr == "" {
		conn.Write([]byte("HTTP/1.1 411 Length Required\r\n\r\n"))
		return
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil || contentLength <= 0 {
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}

	body := make([]byte, contentLength)
	_, err = reader.Read(body)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return
	}

	fileName := strings.TrimPrefix(path, "/files/")
	filePath := filepath.Join(dir, fileName)

	err = os.WriteFile(filePath, body, 0644)
	if err != nil {
		fmt.Print(err)
		conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return
	}

	conn.Write([]byte("HTTP/1.1 201 Created\r\nContent-Type: text/plain\r\n\r\nFile created successfully"))
}

func getContentType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

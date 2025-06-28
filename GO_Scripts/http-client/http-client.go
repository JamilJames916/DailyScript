package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type HTTPClient struct {
	client  *http.Client
	headers map[string]string
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(map[string]string),
	}
}

func (c *HTTPClient) SetHeader(key, value string) {
	c.headers[key] = value
}

func (c *HTTPClient) SetTimeout(timeout time.Duration) {
	c.client.Timeout = timeout
}

func (c *HTTPClient) GET(url string) (*http.Response, error) {
	return c.makeRequest("GET", url, nil)
}

func (c *HTTPClient) POST(url string, body io.Reader) (*http.Response, error) {
	return c.makeRequest("POST", url, body)
}

func (c *HTTPClient) PUT(url string, body io.Reader) (*http.Response, error) {
	return c.makeRequest("PUT", url, body)
}

func (c *HTTPClient) DELETE(url string) (*http.Response, error) {
	return c.makeRequest("DELETE", url, nil)
}

func (c *HTTPClient) makeRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// Set default User-Agent if not provided
	if _, exists := c.headers["User-Agent"]; !exists {
		req.Header.Set("User-Agent", "Go-HTTP-Client/1.0")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (c *HTTPClient) DownloadFile(url, filename string) error {
	resp, err := c.GET(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

func printResponse(resp *http.Response) {
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Headers:\n")
	for key, values := range resp.Header {
		fmt.Printf("  %s: %s\n", key, strings.Join(values, ", "))
	}
	fmt.Println("\nBody:")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return
	}

	fmt.Println(string(body))
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run http-client.go <method> <url> [body]")
		fmt.Println("Methods: GET, POST, PUT, DELETE")
		fmt.Println("Examples:")
		fmt.Println("  go run http-client.go GET https://httpbin.org/get")
		fmt.Println("  go run http-client.go POST https://httpbin.org/post '{\"key\":\"value\"}'")
		fmt.Println("  go run http-client.go download https://httpbin.org/get output.json")
		os.Exit(1)
	}

	method := strings.ToUpper(os.Args[1])
	url := os.Args[2]

	client := NewHTTPClient()
	client.SetHeader("Content-Type", "application/json")

	switch method {
	case "GET":
		resp, err := client.GET(url)
		if err != nil {
			log.Fatalf("GET request failed: %v", err)
		}
		defer resp.Body.Close()
		printResponse(resp)

	case "POST":
		var body io.Reader
		if len(os.Args) > 3 {
			body = strings.NewReader(os.Args[3])
		}
		resp, err := client.POST(url, body)
		if err != nil {
			log.Fatalf("POST request failed: %v", err)
		}
		defer resp.Body.Close()
		printResponse(resp)

	case "PUT":
		var body io.Reader
		if len(os.Args) > 3 {
			body = strings.NewReader(os.Args[3])
		}
		resp, err := client.PUT(url, body)
		if err != nil {
			log.Fatalf("PUT request failed: %v", err)
		}
		defer resp.Body.Close()
		printResponse(resp)

	case "DELETE":
		resp, err := client.DELETE(url)
		if err != nil {
			log.Fatalf("DELETE request failed: %v", err)
		}
		defer resp.Body.Close()
		printResponse(resp)

	case "DOWNLOAD":
		if len(os.Args) < 4 {
			fmt.Println("Usage for download: go run http-client.go download <url> <filename>")
			os.Exit(1)
		}
		filename := os.Args[3]
		err := client.DownloadFile(url, filename)
		if err != nil {
			log.Fatalf("Download failed: %v", err)
		}
		fmt.Printf("File downloaded successfully: %s\n", filename)

	default:
		fmt.Printf("Unsupported method: %s\n", method)
		fmt.Println("Supported methods: GET, POST, PUT, DELETE, DOWNLOAD")
		os.Exit(1)
	}
}

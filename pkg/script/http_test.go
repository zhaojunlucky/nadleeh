package script

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNJSHttp_Request(t *testing.T) {
	njsHttp := &NJSHttp{}

	t.Run("GET_Request", func(t *testing.T) {
		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "success"}`))
		}))
		defer server.Close()

		resp, err := njsHttp.Request("GET", server.URL, nil, nil)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		if resp.Body != `{"message": "success"}` {
			t.Errorf("Expected body %q, got %q", `{"message": "success"}`, resp.Body)
		}

		if resp.ContentType != "application/json" {
			t.Errorf("Expected content type %q, got %q", "application/json", resp.ContentType)
		}
	})

	t.Run("POST_Request_WithBody", func(t *testing.T) {
		expectedBody := `{"name": "test"}`
		
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			
			// Read and verify request body
			body := make([]byte, r.ContentLength)
			r.Body.Read(body)
			if string(body) != expectedBody {
				t.Errorf("Expected request body %q, got %q", expectedBody, string(body))
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": 1, "status": "created"}`))
		}))
		defer server.Close()

		resp, err := njsHttp.Request("POST", server.URL, nil, &expectedBody)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
		}
	})

	t.Run("Request_WithHeaders", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer token123",
			"User-Agent":    "test-agent",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify headers were sent
			if auth := r.Header.Get("Authorization"); auth != "Bearer token123" {
				t.Errorf("Expected Authorization header %q, got %q", "Bearer token123", auth)
			}
			if ua := r.Header.Get("User-Agent"); ua != "test-agent" {
				t.Errorf("Expected User-Agent header %q, got %q", "test-agent", ua)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		defer server.Close()

		resp, err := njsHttp.Request("GET", server.URL, &headers, nil)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Request_InvalidURL", func(t *testing.T) {
		_, err := njsHttp.Request("GET", "invalid-url", nil, nil)
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
	})

	t.Run("Request_CaseInsensitiveMethod", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Test lowercase method gets converted to uppercase
		_, err := njsHttp.Request("post", server.URL, nil, nil)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})
}

func TestNJSHttp_Get(t *testing.T) {
	njsHttp := &NJSHttp{}

	t.Run("Simple_GET", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, World!"))
		}))
		defer server.Close()

		resp, err := njsHttp.Get(server.URL, nil)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		if resp.Body != "Hello, World!" {
			t.Errorf("Expected body %q, got %q", "Hello, World!", resp.Body)
		}
	})

	t.Run("GET_WithHeaders", func(t *testing.T) {
		headers := map[string]string{
			"Accept": "application/json",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if accept := r.Header.Get("Accept"); accept != "application/json" {
				t.Errorf("Expected Accept header %q, got %q", "application/json", accept)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data": "test"}`))
		}))
		defer server.Close()

		resp, err := njsHttp.Get(server.URL, &headers)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if resp.ContentType != "application/json" {
			t.Errorf("Expected content type %q, got %q", "application/json", resp.ContentType)
		}
	})
}

func TestNJSHttp_Post(t *testing.T) {
	njsHttp := &NJSHttp{}

	t.Run("POST_WithBody", func(t *testing.T) {
		body := `{"name": "John", "age": 30}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			
			requestBody := make([]byte, r.ContentLength)
			r.Body.Read(requestBody)
			if string(requestBody) != body {
				t.Errorf("Expected request body %q, got %q", body, string(requestBody))
			}

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": 123}`))
		}))
		defer server.Close()

		resp, err := njsHttp.Post(server.URL, nil, &body)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
		}
	})

	t.Run("POST_WithoutBody", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		resp, err := njsHttp.Post(server.URL, nil, nil)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}

func TestNJSHttp_Put(t *testing.T) {
	njsHttp := &NJSHttp{}

	t.Run("PUT_Request", func(t *testing.T) {
		body := `{"name": "Updated Name"}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		resp, err := njsHttp.Put(server.URL, nil, &body)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}

func TestNJSHttp_Delete(t *testing.T) {
	njsHttp := &NJSHttp{}

	t.Run("DELETE_Request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		resp, err := njsHttp.Delete(server.URL, nil, nil)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
		}
	})

	t.Run("DELETE_WithBody", func(t *testing.T) {
		body := `{"reason": "cleanup"}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE method, got %s", r.Method)
			}
			
			requestBody := make([]byte, r.ContentLength)
			r.Body.Read(requestBody)
			if string(requestBody) != body {
				t.Errorf("Expected request body %q, got %q", body, string(requestBody))
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		resp, err := njsHttp.Delete(server.URL, nil, &body)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}

func TestNJSHttp_Patch(t *testing.T) {
	njsHttp := &NJSHttp{}

	t.Run("PATCH_Request", func(t *testing.T) {
		body := `{"status": "active"}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Note: The original code uses "Patch" instead of "PATCH"
			if r.Method != "PATCH" {
				t.Errorf("Expected PATCH method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		resp, err := njsHttp.Patch(server.URL, nil, &body)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}

func TestNJSHttp_DownloadFile(t *testing.T) {
	njsHttp := &NJSHttp{}

	t.Run("SuccessfulDownload", func(t *testing.T) {
		fileContent := "This is test file content for download"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fileContent))
		}))
		defer server.Close()

		tempDir := t.TempDir()
		downloadPath := filepath.Join(tempDir, "downloaded_file.txt")

		err := njsHttp.DownloadFile("GET", server.URL, downloadPath, nil, nil)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Verify file was created and has correct content
		content, err := os.ReadFile(downloadPath)
		if err != nil {
			t.Fatalf("Failed to read downloaded file: %v", err)
		}

		if string(content) != fileContent {
			t.Errorf("Expected file content %q, got %q", fileContent, string(content))
		}
	})

	t.Run("DownloadWithHeaders", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer token",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if auth := r.Header.Get("Authorization"); auth != "Bearer token" {
				t.Errorf("Expected Authorization header %q, got %q", "Bearer token", auth)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated content"))
		}))
		defer server.Close()

		tempDir := t.TempDir()
		downloadPath := filepath.Join(tempDir, "auth_file.txt")

		err := njsHttp.DownloadFile("GET", server.URL, downloadPath, &headers, nil)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
			t.Error("Downloaded file should exist")
		}
	})

	t.Run("DownloadBadStatus", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		}))
		defer server.Close()

		tempDir := t.TempDir()
		downloadPath := filepath.Join(tempDir, "failed_download.txt")

		err := njsHttp.DownloadFile("GET", server.URL, downloadPath, nil, nil)
		if err == nil {
			t.Error("Expected error for bad status code")
		}

		if !strings.Contains(err.Error(), "bad status") {
			t.Errorf("Expected 'bad status' error, got: %v", err)
		}
	})

	t.Run("DownloadInvalidPath", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("content"))
		}))
		defer server.Close()

		// Try to download to an invalid path
		invalidPath := "/invalid/path/that/does/not/exist/file.txt"

		err := njsHttp.DownloadFile("GET", server.URL, invalidPath, nil, nil)
		if err == nil {
			t.Error("Expected error for invalid download path")
		}
	})

	t.Run("DownloadWithPOSTBody", func(t *testing.T) {
		body := `{"query": "data"}`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			
			requestBody := make([]byte, r.ContentLength)
			r.Body.Read(requestBody)
			if string(requestBody) != body {
				t.Errorf("Expected request body %q, got %q", body, string(requestBody))
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("response data"))
		}))
		defer server.Close()

		tempDir := t.TempDir()
		downloadPath := filepath.Join(tempDir, "post_response.txt")

		err := njsHttp.DownloadFile("POST", server.URL, downloadPath, nil, &body)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Verify file content
		content, err := os.ReadFile(downloadPath)
		if err != nil {
			t.Fatalf("Failed to read downloaded file: %v", err)
		}

		if string(content) != "response data" {
			t.Errorf("Expected file content %q, got %q", "response data", string(content))
		}
	})
}

func TestNJSHttp_decodeContentType(t *testing.T) {
	njsHttp := &NJSHttp{}

	t.Run("SimpleContentType", func(t *testing.T) {
		headers := http.Header{
			"Content-Type": []string{"application/json"},
		}

		contentType, encoding := njsHttp.decodeContentType(headers)
		if contentType != "application/json" {
			t.Errorf("Expected content type %q, got %q", "application/json", contentType)
		}
		if encoding != "" {
			t.Errorf("Expected empty encoding, got %q", encoding)
		}
	})

	t.Run("ContentTypeWithCharset", func(t *testing.T) {
		headers := http.Header{
			"Content-Type": []string{"text/html; charset=utf-8"},
		}

		contentType, encoding := njsHttp.decodeContentType(headers)
		if contentType != "text/html" {
			t.Errorf("Expected content type %q, got %q", "text/html", contentType)
		}
		if encoding != "utf-8" {
			t.Errorf("Expected encoding %q, got %q", "utf-8", encoding)
		}
	})

	t.Run("ContentTypeWithBoundary", func(t *testing.T) {
		headers := http.Header{
			"Content-Type": []string{"multipart/form-data; boundary=something"},
		}

		contentType, encoding := njsHttp.decodeContentType(headers)
		if contentType != "multipart/form-data" {
			t.Errorf("Expected content type %q, got %q", "multipart/form-data", contentType)
		}
		if encoding != "something" {
			t.Errorf("Expected encoding %q, got %q", "something", encoding)
		}
	})

	t.Run("CaseInsensitiveHeader", func(t *testing.T) {
		headers := http.Header{
			"content-type": []string{"application/xml"},
		}

		contentType, encoding := njsHttp.decodeContentType(headers)
		if contentType != "application/xml" {
			t.Errorf("Expected content type %q, got %q", "application/xml", contentType)
		}
		if encoding != "" {
			t.Errorf("Expected empty encoding, got %q", encoding)
		}
	})

	t.Run("NoContentTypeHeader", func(t *testing.T) {
		headers := http.Header{
			"Other-Header": []string{"value"},
		}

		contentType, encoding := njsHttp.decodeContentType(headers)
		if contentType != "" {
			t.Errorf("Expected empty content type, got %q", contentType)
		}
		if encoding != "" {
			t.Errorf("Expected empty encoding, got %q", encoding)
		}
	})

	t.Run("EmptyContentTypeValue", func(t *testing.T) {
		headers := http.Header{
			"Content-Type": []string{},
		}

		contentType, encoding := njsHttp.decodeContentType(headers)
		if contentType != "" {
			t.Errorf("Expected empty content type, got %q", contentType)
		}
		if encoding != "" {
			t.Errorf("Expected empty encoding, got %q", encoding)
		}
	})

	t.Run("ContentTypeWithSemicolonButNoEquals", func(t *testing.T) {
		headers := http.Header{
			"Content-Type": []string{"text/plain; utf-8"},
		}

		contentType, encoding := njsHttp.decodeContentType(headers)
		if contentType != "text/plain" {
			t.Errorf("Expected content type %q, got %q", "text/plain", contentType)
		}
		if encoding != " utf-8" {
			t.Errorf("Expected encoding %q, got %q", " utf-8", encoding)
		}
	})
}

func TestHttpResponse_Structure(t *testing.T) {
	njsHttp := &NJSHttp{}

	t.Run("ResponseFieldsPopulated", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Custom-Header", "custom-value")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"result": "success"}`))
		}))
		defer server.Close()

		resp, err := njsHttp.Get(server.URL, nil)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Verify all fields are populated
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
		}

		if resp.Status != "201 Created" {
			t.Errorf("Expected status %q, got %q", "201 Created", resp.Status)
		}

		if resp.Headers == nil {
			t.Error("Expected headers to be populated")
		}

		if resp.Body != `{"result": "success"}` {
			t.Errorf("Expected body %q, got %q", `{"result": "success"}`, resp.Body)
		}

		if resp.ContentType != "application/json" {
			t.Errorf("Expected content type %q, got %q", "application/json", resp.ContentType)
		}

		if resp.ContentEncoding != "utf-8" {
			t.Errorf("Expected content encoding %q, got %q", "utf-8", resp.ContentEncoding)
		}

		// Check custom header exists
		customHeaders := resp.Headers["Custom-Header"]
		if len(customHeaders) == 0 || customHeaders[0] != "custom-value" {
			t.Errorf("Expected custom header %q, got %v", "custom-value", customHeaders)
		}
	})
}

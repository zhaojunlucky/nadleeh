package script

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

type HttpResponse struct {
	StatusCode    int
	Status        string
	Headers       map[string][]string
	Body          string
	ContentLength int64
	ContentType   string
}

type HttpRequest struct {
	Method  string
	Headers map[string]string
	Url     string
	Body    string
}

func Request(method string, url string, headers *map[string]string, body *string) (*HttpResponse, error) {
	var bodyReader io.Reader = nil
	if body != nil {
		bodyReader = bytes.NewBufferString(*body)
	}
	req, err := http.NewRequest(strings.ToUpper(method), url, bodyReader)
	if err != nil {
		return nil, err
	}
	if headers != nil {
		for name, value := range *headers {
			req.Header.Add(name, value)
		}
	}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	httpResp := &HttpResponse{
		StatusCode:    resp.StatusCode,
		Status:        resp.Status,
		Headers:       resp.Header,
		ContentLength: resp.ContentLength,
		ContentType:   decodeContentType(resp.Header),
		Body:          string(bodyBytes),
	}
	return httpResp, nil
}

func Get(url string, headers *map[string]string, params *map[string]string) (*HttpResponse, error) {

}

func decodeContentType(header http.Header) string {
	for name, value := range header {
		if strings.ToUpper(name) == "content-type" {
			if len(value) > 0 {
				return value[0]
			} else {
				return ""
			}
		}
	}
	return ""
}

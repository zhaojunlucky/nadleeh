package script

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

type NJSHttp struct {
}

type HttpResponse struct {
	StatusCode      int
	Status          string
	Headers         map[string][]string
	Body            string
	ContentLength   int64
	ContentType     string
	ContentEncoding string
}

type HttpRequest struct {
	Method  string
	Headers map[string]string
	Url     string
	Body    string
}

func (js *NJSHttp) Request(method string, url string, headers *map[string]string, body *string) (*HttpResponse, error) {
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
		Body:          string(bodyBytes),
	}
	httpResp.ContentType, httpResp.ContentEncoding = js.decodeContentType(resp.Header)
	return httpResp, nil
}

func (js *NJSHttp) Get(url string, headers *map[string]string) (*HttpResponse, error) {
	return js.Request("GET", url, headers, nil)
}

func (js *NJSHttp) Delete(url string, headers *map[string]string, body *string) (*HttpResponse, error) {
	return js.Request("DELETE", url, headers, body)
}

func (js *NJSHttp) Post(url string, headers *map[string]string, body *string) (*HttpResponse, error) {
	return js.Request("POST", url, headers, body)
}

func (js *NJSHttp) Put(url string, headers *map[string]string, body *string) (*HttpResponse, error) {
	return js.Request("PUT", url, headers, body)
}

func (js *NJSHttp) Patch(url string, headers *map[string]string, body *string) (*HttpResponse, error) {
	return js.Request("Patch", url, headers, body)
}

func (js *NJSHttp) DownloadFile(method string, url string, downloadPath string, headers *map[string]string, body *string) error {
	out, err := os.Create(downloadPath)
	if err != nil {
		return err
	}
	defer out.Close()

	var bodyReader io.Reader = nil
	if body != nil {
		bodyReader = bytes.NewBufferString(*body)
	}
	req, err := http.NewRequest(strings.ToUpper(method), url, bodyReader)
	if err != nil {
		return err
	}
	if headers != nil {
		for name, value := range *headers {
			req.Header.Add(name, value)
		}
	}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	log.Infof("download file to %s", downloadPath)
	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (js *NJSHttp) decodeContentType(header http.Header) (string, string) {
	for name, value := range header {
		if strings.ToLower(name) == "content-type" {
			if len(value) > 0 {
				ct := value[0]
				if strings.ContainsRune(ct, ';') {
					ctEnc := strings.Split(ct, ";")
					if strings.ContainsRune(ctEnc[1], '=') {
						return ctEnc[0], strings.Split(ctEnc[1], "=")[1]
					}
					return ctEnc[0], ctEnc[1]
				} else {
					return ct, ""
				}

			} else {
				return "", ""
			}
		}
	}
	return "", ""
}

package httpclient

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func Call(method string, url string, content interface{}, headers map[string]string) ([]byte, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var body io.Reader
	switch content.(type) {
	case string:
		body = strings.NewReader(content.(string))
	case []byte:
		body = bytes.NewReader(content.([]byte))
	default:
		return nil, errors.New("wrong content type")
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status error: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	return data, err
}

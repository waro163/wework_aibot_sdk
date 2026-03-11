package weworkaibotsdk

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

// ref: https://github.com/WecomTeam/aibot-node-sdk/blob/main/src/api.ts

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ApiClient struct {
	client HttpClient
}

func NewApiClient(client HttpClient) *ApiClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &ApiClient{
		client: client,
	}
}

type DownloadFileResponse struct {
	FileName string
	FileData []byte
}

func (c *ApiClient) DownloadFileRaw(reqUrl string) (*DownloadFileResponse, error) {
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	var filename string
	if contentDisposition != "" {
		// 优先匹配 filename*=UTF-8''xxx 格式（RFC 5987）
		utf8Re := regexp.MustCompile(`(?i)filename\*=UTF-8''([^;\s]+)`)
		if utf8Match := utf8Re.FindStringSubmatch(contentDisposition); utf8Match != nil {
			decoded, err := url.QueryUnescape(utf8Match[1])
			if err == nil {
				filename = decoded
			}
		} else {
			// 匹配 filename="xxx" 或 filename=xxx 格式
			re := regexp.MustCompile(`(?i)filename="?([^";\s]+)"?`)
			if match := re.FindStringSubmatch(contentDisposition); match != nil {
				decoded, err := url.QueryUnescape(match[1])
				if err == nil {
					filename = decoded
				}
			}
		}
	}
	return &DownloadFileResponse{
		FileName: filename,
		FileData: fileData,
	}, nil
}

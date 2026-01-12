package zlm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client ZLMediaKit API 客户端
type Client struct {
	baseURL    string
	secret     string
	httpClient *http.Client
}

// NewClient 创建 ZLMediaKit 客户端
func NewClient(host, port, secret string) *Client {
	return &Client{
		baseURL: fmt.Sprintf("http://%s:%s", host, port),
		secret:  secret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetMediaList 获取流列表
func (c *Client) GetMediaList(app, stream string) (*MediaListResponse, error) {
	params := url.Values{}
	params.Set("secret", c.secret)
	if app != "" {
		params.Set("app", app)
	}
	if stream != "" {
		params.Set("stream", stream)
	}

	resp, err := c.get("/index/api/getMediaList", params)
	if err != nil {
		return nil, err
	}

	var result MediaListResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CloseStreams 关闭流
func (c *Client) CloseStreams(app, stream string, force bool) (*CommonResponse, error) {
	params := url.Values{}
	params.Set("secret", c.secret)
	params.Set("app", app)
	params.Set("stream", stream)
	if force {
		params.Set("force", "1")
	}

	resp, err := c.get("/index/api/close_streams", params)
	if err != nil {
		return nil, err
	}

	var result CommonResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetServerConfig 获取服务器配置
func (c *Client) GetServerConfig() (*ServerConfigResponse, error) {
	params := url.Values{}
	params.Set("secret", c.secret)

	resp, err := c.get("/index/api/getServerConfig", params)
	if err != nil {
		return nil, err
	}

	var result ServerConfigResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) get(path string, params url.Values) ([]byte, error) {
	reqURL := fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// MediaListResponse 流列表响应
type MediaListResponse struct {
	Code int         `json:"code"`
	Data []MediaInfo `json:"data"`
}

// MediaInfo 流信息
type MediaInfo struct {
	App           string  `json:"app"`
	Stream        string  `json:"stream"`
	Schema        string  `json:"schema"`
	ReaderCount   int     `json:"readerCount"`
	TotalReaderCount int  `json:"totalReaderCount"`
	BytesSpeed    int     `json:"bytesSpeed"`
	CreateStamp   int64   `json:"createStamp"`
	AliveSecond   int     `json:"aliveSecond"`
	Tracks        []Track `json:"tracks"`
}

// Track 轨道信息
type Track struct {
	CodecID   int    `json:"codec_id"`
	CodecType int    `json:"codec_type"`
	Ready     bool   `json:"ready"`
	FPS       int    `json:"fps"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

// CommonResponse 通用响应
type CommonResponse struct {
	Code   int    `json:"code"`
	Result int    `json:"result"`
	Msg    string `json:"msg"`
}

// ServerConfigResponse 服务器配置响应
type ServerConfigResponse struct {
	Code int                    `json:"code"`
	Data []map[string]string    `json:"data"`
}

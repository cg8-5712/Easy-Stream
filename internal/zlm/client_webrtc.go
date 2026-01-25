package zlm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// WebRTCPlayResponse WebRTC 播放响应
type WebRTCPlayResponse struct {
	Code int    `json:"code"`
	SDP  string `json:"sdp"`
}

// WebRTCPlay 发送 WebRTC 播放请求到 ZLMediaKit
// app: 应用名，如 "live"
// stream: 流名称（stream_key）
// offerSDP: 客户端的 SDP offer
// 返回 ZLMediaKit 的 SDP answer
func (c *Client) WebRTCPlay(app, stream string, offerSDP string) (*WebRTCPlayResponse, error) {
	params := url.Values{}
	if app != "" {
		params.Set("app", app)
	}
	if stream != "" {
		params.Set("stream", stream)
	}
	params.Set("type", "play")

	reqURL := fmt.Sprintf("%s/index/api/webrtc?%s", c.baseURL, params.Encode())

	// 创建 body reader，这样 Go 会自动计算 Content-Length
	body := strings.NewReader(offerSDP)
	req, err := http.NewRequest("POST", reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/sdp")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result WebRTCPlayResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		// 如果解析 JSON 失败，可能是纯 SDP 文本格式
		return &WebRTCPlayResponse{
			Code: 0,
			SDP:  string(respBody),
		}, nil
	}

	return &result, nil
}

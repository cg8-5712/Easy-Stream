package service

// WebRTCPlayResponse WebRTC 播放响应
type WebRTCPlayResponse struct {
	Code int    `json:"code"`
	SDP  string `json:"sdp"`
}

// WebRTCPlay 发送 WebRTC 播放请求到 ZLMediaKit
// streamKey: 流的 stream_key（推流码）
// offerSDP: 客户端的 SDP offer
// 返回 ZLMediaKit 的 SDP answer
func (s *StreamService) WebRTCPlay(streamKey, offerSDP string) (*WebRTCPlayResponse, error) {
	resp, err := s.zlmClient.WebRTCPlay("live", streamKey, offerSDP)
	if err != nil {
		return nil, err
	}

	return &WebRTCPlayResponse{
		Code: resp.Code,
		SDP:  resp.SDP,
	}, nil
}

package model

// ZLMediaKit Hook 请求结构

// OnPublishRequest 推流开始回调
type OnPublishRequest struct {
	App        string `json:"app"`
	Stream     string `json:"stream"`
	Schema     string `json:"schema"`
	MediaSrvID string `json:"mediaServerId"`
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	Params     string `json:"params"`
}

// OnUnpublishRequest 推流结束回调
type OnUnpublishRequest struct {
	App        string `json:"app"`
	Stream     string `json:"stream"`
	Schema     string `json:"schema"`
	MediaSrvID string `json:"mediaServerId"`
}

// OnFlowReportRequest 流量统计回调
type OnFlowReportRequest struct {
	App           string `json:"app"`
	Stream        string `json:"stream"`
	Schema        string `json:"schema"`
	MediaSrvID    string `json:"mediaServerId"`
	TotalBytes    int64  `json:"totalBytes"`
	Duration      int    `json:"duration"`
	Player        bool   `json:"player"`
	TotalBytesIn  int64  `json:"totalBytesIn"`
	TotalBytesOut int64  `json:"totalBytesOut"`
}

// OnStreamNoneReaderRequest 无人观看回调
type OnStreamNoneReaderRequest struct {
	App        string `json:"app"`
	Stream     string `json:"stream"`
	Schema     string `json:"schema"`
	MediaSrvID string `json:"mediaServerId"`
}

// HookResponse Hook 响应
type HookResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

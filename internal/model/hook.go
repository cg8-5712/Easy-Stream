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

// OnPlayRequest 播放开始回调
type OnPlayRequest struct {
	App        string `json:"app"`
	Stream     string `json:"stream"`
	Schema     string `json:"schema"`
	MediaSrvID string `json:"mediaServerId"`
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	Params     string `json:"params"`
	ID         string `json:"id"` // 播放器唯一标识
}

// OnPlayerDisconnectRequest 播放器断开回调
type OnPlayerDisconnectRequest struct {
	App        string `json:"app"`
	Stream     string `json:"stream"`
	Schema     string `json:"schema"`
	MediaSrvID string `json:"mediaServerId"`
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	ID         string `json:"id"` // 播放器唯一标识
}

// OnRecordMP4Request 录制完成回调
type OnRecordMP4Request struct {
	App        string  `json:"app"`
	Stream     string  `json:"stream"`
	MediaSrvID string  `json:"mediaServerId"`
	FileName   string  `json:"file_name"`   // 文件名
	FilePath   string  `json:"file_path"`   // 文件绝对路径
	FileSize   int64   `json:"file_size"`   // 文件大小（字节）
	Folder     string  `json:"folder"`      // 文件所在目录
	StartTime  int64   `json:"start_time"`  // 录制开始时间戳
	TimeLen    float64 `json:"time_len"`    // 录制时长（秒）
	URL        string  `json:"url"`         // 播放地址
}

// HookResponse Hook 响应
type HookResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

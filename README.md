# Easy-Stream

内网低延迟直播系统，基于 ZLMediaKit + WebRTC，支持多协议推流、多协议播放、实时观看、推流管理。

## 特性

- **超低延迟**: WebRTC 播放，延迟 < 200ms
- **多协议推流**: 支持 RTMP / RTSP / SRT / HTTP-TS / WebRTC
- **多协议播放**: 支持 WebRTC / HLS / HTTP-FLV / WebSocket-FLV
- **推流管理**: 创建/删除推流码、强制断流
- **实时监控**: 流状态、码率、帧率实时更新
- **录制回放**: 支持直播录制（MP4/HLS），动态开关录制
- **私有直播**: 支持密码保护的私有直播
- **推流验证**: 无效推流码自动拒绝
- **超时断流**: 超时自动断流机制
- **Hook 集成**: 与 ZLMediaKit 深度集成
- **云存储**: 支持本地存储、AWS S3、腾讯云 COS、阿里云 OSS

## 技术栈

| 组件 | 技术 |
|------|------|
| 流媒体服务器 | ZLMediaKit |
| 后端 | Go + Gin |
| 前端 | React + TypeScript + Ant Design |
| 数据库 | PostgreSQL |
| 缓存 | Redis |

## 项目结构

```
Easy-Stream/
├── cmd/server/             # 程序入口
├── internal/
│   ├── config/             # 配置管理
│   ├── handler/            # HTTP 处理器
│   ├── middleware/         # 中间件 (JWT、CORS、日志)
│   ├── model/              # 数据模型
│   ├── repository/         # 数据访问层
│   ├── service/            # 业务逻辑层
│   └── zlm/                # ZLMediaKit 客户端
├── pkg/
│   ├── logger/             # 日志工具
│   └── utils/              # 通用工具
├── frontend/               # React 前端
├── deploy/                 # 部署配置
├── scripts/                # 脚本
└── docs/                   # 文档
```

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (可选)

### 1. 克隆项目

```bash
git clone https://github.com/yourusername/easy-stream.git
cd easy-stream
```

### 2. 配置

复制并修改配置文件：

```bash
cp config.yaml.example config.yaml
```

编辑 `config.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: "8080"
  mode: "debug"

database:
  host: "localhost"
  port: "5432"
  user: "easystream"
  password: "your_password"
  dbname: "easystream"
  sslmode: "disable"

redis:
  host: "localhost"
  port: "6379"
  password: ""
  db: 0

jwt:
  secret: "your-jwt-secret-change-in-production"
  expireHour: 24

zlmediakit:
  host: "localhost"
  port: "80"
  secret: "035c73f7-bb6b-4889-a715-d9eb2d1925cc"
  hookBaseURL: "http://localhost:8080/api/v1/hooks"
  httpPort: "80"           # HLS/FLV 播放端口
  httpsPort: "443"         # HTTPS 端口（可选）
  externalHost: "localhost" # 外部访问地址（公网 IP 或域名）
  webrtcPort: "8000"       # WebRTC 端口

log:
  level: "info"
```

### 3. 初始化数据库

```bash
psql -U postgres -c "CREATE DATABASE easystream;"
psql -U postgres -d easystream -f scripts/init-db.sql

# 如果是升级，运行迁移脚本
psql -U postgres -d easystream -f scripts/migrations/002_add_record_fields.sql
```

### 4. 运行后端

```bash
go mod tidy
go run ./cmd/server/
```

### 5. 运行前端

```bash
cd frontend
npm install
npm run dev
```

## Docker 部署

### 一键启动

```bash
cd deploy
docker-compose up -d
```

### 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| 前端 | 3000 | Web 管理控制台 |
| 后端 API | 8081 | REST API |
| ZLMediaKit HTTP | 80 | HLS/FLV 播放、HTTP-TS 推流 |
| ZLMediaKit HTTPS | 443 | HTTPS 播放（可选） |
| ZLMediaKit API | 8080 | 流媒体管理 API |
| RTMP | 1935 | RTMP 推流 |
| RTSP | 554 | RTSP 推流 |
| SRT | 9000/UDP | SRT 推流 |
| WebRTC | 8000/UDP+TCP | WebRTC 播放和推流 |

## API 接口

### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 登录 |
| POST | `/api/v1/auth/logout` | 登出 |
| GET | `/api/v1/auth/profile` | 获取用户信息 |

### 推流管理

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | `/api/v1/streams` | 获取推流列表 | 游客/管理员 |
| POST | `/api/v1/streams` | 创建推流码 | 管理员 |
| GET | `/api/v1/streams/view/:id` | 通过 ID 查看直播（不含 key） | 游客/管理员 |
| GET | `/api/v1/streams/id/:id` | 通过 ID 获取推流详情（含 key） | 管理员 |
| GET | `/api/v1/streams/:key` | 通过推流码获取详情 | 管理员 |
| GET | `/api/v1/streams/:key/play-urls` | 获取播放地址 | 游客/管理员 |
| PUT | `/api/v1/streams/:key` | 更新推流信息 | 管理员 |
| DELETE | `/api/v1/streams/:key` | 删除推流码 | 管理员 |
| POST | `/api/v1/streams/:key/kick` | 强制断流 | 管理员 |
| POST | `/api/v1/streams/:key/end` | 结束直播 | 管理员 |

### WebRTC

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | `/api/v1/streams/webrtc/:id` | WebRTC 播放 | 游客/管理员 |
| GET | `/api/v1/streams/webrtc/:id` | 获取 WebRTC SDP | 游客/管理员 |
| POST | `/api/v1/streams/:key/webrtc-push` | WebRTC 推流 | 无需认证（需有效 stream_key） |

### 系统

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/system/health` | 健康检查 |
| GET | `/api/v1/system/stats` | 系统统计 |

## 推流地址

创建推流码后，使用以下地址推流：

### RTMP 推流（推荐，最稳定）
```
服务器: rtmp://{server}:1935/live
串流密钥: {stream_key}

完整地址: rtmp://{server}:1935/live/{stream_key}
```

### RTSP 推流（摄像头设备）
```
rtsp://{server}:554/live/{stream_key}
```

### SRT 推流（低延迟，抗丢包）
```
srt://{server}:9000?streamid=#!::r=live/{stream_key},m=publish
```

### HTTP-TS 推流（防火墙友好）⭐ 新增
```
http://{server}:80/live/{stream_key}.live.ts
```

### WebRTC 推流（浏览器推流）⭐ 新增
```
POST http://{server}:8080/api/v1/streams/{stream_key}/webrtc-push
Content-Type: application/json

{
  "sdp": "v=0\r\no=- ..."
}
```

## 播放地址

### 方式1：通过 API 获取所有播放地址（推荐）

```bash
# 游客访问公开直播
curl http://{server}:8080/api/v1/streams/{stream_key}/play-urls

# 管理员访问（返回推流地址和播放地址）
curl -H "Authorization: Bearer {token}" \
  http://{server}:8080/api/v1/streams/{stream_key}/play-urls
```

**响应示例**：
```json
{
  "stream_id": 123,
  "stream_key": "abc123",
  "stream_name": "测试直播",
  "status": "pushing",
  "play_urls": {
    "webrtc": "webrtc://192.168.1.9:8000/live/abc123",
    "hls": "http://192.168.1.9:80/live/abc123/hls.m3u8",
    "http_flv": "http://192.168.1.9:80/live/abc123.live.flv",
    "ws_flv": "ws://192.168.1.9:80/live/abc123.live.flv"
  },
  "push_urls": {  // 仅管理员可见
    "rtmp": "rtmp://192.168.1.9:1935/live/abc123",
    "rtsp": "rtsp://192.168.1.9:554/live/abc123",
    "srt": "srt://192.168.1.9:9000?streamid=#!::r=live/abc123,m=publish",
    "http_ts": "http://192.168.1.9:80/live/abc123.live.ts"
  }
}
```

### 方式2：直接使用播放地址

#### WebRTC 播放（超低延迟 < 200ms）
```
webrtc://{server}:8000/live/{stream_key}
```

#### HLS 播放（兼容性最好，延迟 6-15s）⭐ 新增
```
http://{server}:80/live/{stream_key}/hls.m3u8
```

#### HTTP-FLV 播放（延迟 1-3s）⭐ 新增
```
http://{server}:80/live/{stream_key}.live.flv
```

#### WebSocket-FLV 播放（延迟 1-3s）⭐ 新增
```
ws://{server}:80/live/{stream_key}.live.flv
```

### 播放协议对比

| 协议 | 延迟 | 兼容性 | 适用场景 |
|------|------|--------|---------|
| WebRTC | < 200ms | 现代浏览器 | 实时互动、低延迟要求 |
| HTTP-FLV | 1-3s | 需 flv.js | 平衡延迟和兼容性 |
| WebSocket-FLV | 1-3s | 需 flv.js | 实时性要求较高 |
| HLS | 6-15s | 所有设备 | 兼容性优先、移动端 |

## 回放地址

录制开启且直播结束后可用：

```
HTTP: http://{server}:8080/recordings/{record_file}
```

## 默认账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | admin123 | 管理员 |

> ⚠️ 生产环境请务必修改默认密码

## 开发

### 后端开发

```bash
# 运行
go run ./cmd/server/

# 编译
go build -o easy-stream ./cmd/server/

# 测试
go test ./...
```

### 前端开发

```bash
cd frontend

# 开发
npm run dev

# 构建
npm run build

# 预览
npm run preview
```

## 配置 ZLMediaKit Hook

在 ZLMediaKit 配置文件中启用 Hook：

```ini
[hook]
enable=1
on_publish=http://backend:8080/api/v1/hooks/on_publish
on_unpublish=http://backend:8080/api/v1/hooks/on_unpublish
on_flow_report=http://backend:8080/api/v1/hooks/on_flow_report
on_stream_none_reader=http://backend:8080/api/v1/hooks/on_stream_none_reader
on_play=http://backend:8080/api/v1/hooks/on_play
on_player_disconnect=http://backend:8080/api/v1/hooks/on_player_disconnect
```

## 架构图

```
推流端 (RTMP/RTSP/SRT/HTTP-TS/WebRTC)
        │
        ▼
┌──────────────────┐
│   ZLMediaKit     │
│  (流媒体服务器)   │
│                  │
│  支持协议：       │
│  推流: RTMP/RTSP/│
│       SRT/TS/    │
│       WebRTC     │
│  播放: WebRTC/   │
│       HLS/FLV    │
└───────┬──────────┘
        │
   ┌────┴────┐
   │         │
   ▼         ▼
播放端    HTTP API
(多协议)     │
   │        ▼
   │  ┌───────────┐
   │  │  后端服务  │
   │  │   (Go)    │
   │  │           │
   │  │ - 推流验证 │
   │  │ - 状态管理 │
   │  │ - 录制控制 │
   │  └─────┬─────┘
   │        │
   │        ▼
   │  ┌───────────┐
   └─►│  前端控制台 │
      │  (React)  │
      └───────────┘
```

## 使用示例

### OBS 推流配置

1. **RTMP 推流（推荐）**
   - 打开 OBS → 设置 → 推流
   - 服务：自定义
   - 服务器：`rtmp://192.168.1.9:1935/live`
   - 串流密钥：`{你的stream_key}`

2. **SRT 推流（低延迟）**
   - 服务：自定义
   - 服务器：`srt://192.168.1.9:9000?streamid=#!::r=live/{stream_key},m=publish`

### 浏览器播放（HLS.js）

```html
<!DOCTYPE html>
<html>
<head>
  <script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
</head>
<body>
  <video id="video" controls width="800"></video>
  <script>
    const video = document.getElementById('video');
    const hls = new Hls();
    hls.loadSource('http://192.168.1.9:80/live/{stream_key}/hls.m3u8');
    hls.attachMedia(video);
    hls.on(Hls.Events.MANIFEST_PARSED, function() {
      video.play();
    });
  </script>
</body>
</html>
```

### FFmpeg 推流测试

```bash
# RTMP 推流
ffmpeg -re -i input.mp4 -c copy -f flv \
  rtmp://192.168.1.9:1935/live/{stream_key}

# HTTP-TS 推流
ffmpeg -re -i input.mp4 -c copy -f mpegts \
  http://192.168.1.9:80/live/{stream_key}.live.ts

# SRT 推流
ffmpeg -re -i input.mp4 -c copy -f mpegts \
  "srt://192.168.1.9:9000?streamid=#!::r=live/{stream_key},m=publish"
```

## License

[GNU GENERAL PUBLIC LICENSE Version 3](./LICENSE)

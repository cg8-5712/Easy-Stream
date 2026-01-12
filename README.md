# Easy-Stream

内网低延迟直播系统，基于 ZLMediaKit + WebRTC，支持多协议推流、实时观看、推流管理。

## 特性

- **低延迟**: WebRTC 分发，延迟 < 200ms
- **多协议推流**: 支持 RTMP / RTSP / SRT
- **推流管理**: 创建/删除推流码、强制断流
- **实时监控**: 流状态、码率、帧率实时更新
- **空流检测**: 自动检测并销毁无内容流
- **Hook 集成**: 与 ZLMediaKit 深度集成

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

log:
  level: "info"
```

### 3. 初始化数据库

```bash
psql -U postgres -c "CREATE DATABASE easystream;"
psql -U postgres -d easystream -f scripts/init-db.sql
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
| ZLMediaKit HTTP | 8080 | 流媒体 API |
| RTMP | 1935 | RTMP 推流 |
| RTSP | 8554 | RTSP 推流 |
| SRT | 9000 | SRT 推流 |
| WebRTC | 8000 | WebRTC 播放 |

## API 接口

### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 登录 |
| POST | `/api/v1/auth/logout` | 登出 |
| GET | `/api/v1/auth/profile` | 获取用户信息 |

### 推流管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/streams` | 获取推流列表 |
| POST | `/api/v1/streams` | 创建推流码 |
| GET | `/api/v1/streams/:key` | 获取推流详情 |
| PUT | `/api/v1/streams/:key` | 更新推流信息 |
| DELETE | `/api/v1/streams/:key` | 删除推流码 |
| POST | `/api/v1/streams/:key/kick` | 强制断流 |

### 系统

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/system/health` | 健康检查 |
| GET | `/api/v1/system/stats` | 系统统计 |

## 推流地址

创建推流码后，使用以下地址推流：

```
RTMP:  rtmp://{server}:1935/live/{stream_key}
RTSP:  rtsp://{server}:8554/live/{stream_key}
SRT:   srt://{server}:9000?streamid=#!::r=live/{stream_key},m=publish
```

## 播放地址

```
WebRTC: webrtc://{server}:8000/live/{stream_key}
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
```

## 架构图

```
推流端 (RTMP/SRT/RTSP)
        │
        ▼
┌──────────────────┐
│   ZLMediaKit     │
│  (流媒体服务器)   │
└───────┬──────────┘
        │
   ┌────┴────┐
   │         │
   ▼         ▼
WebRTC    HTTP API
播放        │
   │        ▼
   │  ┌───────────┐
   │  │  后端服务  │
   │  │   (Go)    │
   │  └─────┬─────┘
   │        │
   │        ▼
   │  ┌───────────┐
   └─►│  前端控制台 │
      │  (React)  │
      └───────────┘
```

## License

MIT

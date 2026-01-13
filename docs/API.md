# Easy-Stream API 文档

## 目录

- [基本信息](#基本信息)
- [认证接口](#1-认证接口)
- [推流管理接口](#2-推流管理接口)
- [系统接口](#3-系统接口)
- [ZLMediaKit Hook 接口](#4-zlmediakit-hook-接口)
- [数据模型](#数据模型)
- [错误码](#错误码)

---

## 基本信息

- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: JWT Bearer Token
- **Content-Type**: `application/json`
- **字符编码**: UTF-8

### 认证说明

除了公开接口外，所有接口都需要在 HTTP Header 中携带 JWT Token：

```
Authorization: Bearer {your_jwt_token}
```

---

## 1. 认证接口

### 1.1 用户登录

用户通过用户名和密码登录系统，获取 JWT Token。

**接口地址**

```
POST /api/v1/auth/login
```

**请求参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

**请求示例**

```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应示例**

成功 (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin",
    "role": "admin",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

失败 (401 Unauthorized):
```json
{
  "error": "invalid credentials"
}
```

---

### 1.2 用户登出

用户退出登录。

**接口地址**

```
POST /api/v1/auth/logout
```

**请求头**

```
Authorization: Bearer {token}
```

**响应示例**

成功 (200 OK):
```json
{
  "message": "logged out"
}
```

---

### 1.3 获取当前用户信息

获取当前登录用户的详细信息。

**接口地址**

```
GET /api/v1/auth/profile
```

**请求头**

```
Authorization: Bearer {token}
```

**响应示例**

成功 (200 OK):
```json
{
  "id": 1,
  "username": "admin",
  "role": "admin",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

---

## 2. 推流管理接口

> 所有推流管理接口都需要 JWT 认证

### 2.1 获取推流列表

分页获取推流列表，支持按状态过滤。

**接口地址**

```
GET /api/v1/streams
```

**请求头**

```
Authorization: Bearer {token}
```

**查询参数**

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| status | string | 否 | - | 流状态过滤：`idle` / `pushing` / `destroyed` |
| page | integer | 否 | 1 | 页码 |
| pageSize | integer | 否 | 20 | 每页数量 |

**请求示例**

```
GET /api/v1/streams?status=pushing&page=1&pageSize=20
```

**响应示例**

成功 (200 OK):
```json
{
  "total": 100,
  "streams": [
    {
      "id": 1,
      "stream_key": "abc123def456",
      "name": "会议室直播",
      "device_id": "camera-001",
      "status": "pushing",
      "protocol": "rtmp",
      "bitrate": 2500,
      "fps": 30,
      "last_frame_at": "2024-01-01T12:00:00Z",
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    },
    {
      "id": 2,
      "stream_key": "xyz789ghi012",
      "name": "监控摄像头",
      "device_id": "camera-002",
      "status": "idle",
      "protocol": "",
      "bitrate": 0,
      "fps": 0,
      "last_frame_at": null,
      "created_at": "2024-01-01T09:00:00Z",
      "updated_at": "2024-01-01T09:00:00Z"
    }
  ]
}
```

---

### 2.2 创建推流码

创建新的推流码，用于推流认证。

**接口地址**

```
POST /api/v1/streams
```

**请求头**

```
Authorization: Bearer {token}
```

**请求参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 是 | 推流名称 |
| device_id | string | 否 | 设备 ID |

**请求示例**

```json
{
  "name": "会议室直播",
  "device_id": "camera-001"
}
```

**响应示例**

成功 (201 Created):
```json
{
  "id": 1,
  "stream_key": "abc123def456",
  "name": "会议室直播",
  "device_id": "camera-001",
  "status": "idle",
  "protocol": "",
  "bitrate": 0,
  "fps": 0,
  "last_frame_at": null,
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

**推流地址**

创建成功后，可使用以下地址进行推流：

- **RTMP**: `rtmp://{server}:1935/live/{stream_key}`
- **RTSP**: `rtsp://{server}:8554/live/{stream_key}`
- **SRT**: `srt://{server}:9000?streamid=#!::r=live/{stream_key},m=publish`

**播放地址**

- **WebRTC**: `webrtc://{server}:8000/live/{stream_key}`

---

### 2.3 获取推流详情

根据推流密钥获取推流的详细信息。

**接口地址**

```
GET /api/v1/streams/:key
```

**请求头**

```
Authorization: Bearer {token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 (stream_key) |

**请求示例**

```
GET /api/v1/streams/abc123def456
```

**响应示例**

成功 (200 OK):
```json
{
  "id": 1,
  "stream_key": "abc123def456",
  "name": "会议室直播",
  "device_id": "camera-001",
  "status": "pushing",
  "protocol": "rtmp",
  "bitrate": 2500,
  "fps": 30,
  "last_frame_at": "2024-01-01T12:00:00Z",
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

失败 (404 Not Found):
```json
{
  "error": "stream not found"
}
```

---

### 2.4 更新推流信息

更新推流的名称或设备 ID。

**接口地址**

```
PUT /api/v1/streams/:key
```

**请求头**

```
Authorization: Bearer {token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |

**请求参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 否 | 推流名称 |
| device_id | string | 否 | 设备 ID |

**请求示例**

```json
{
  "name": "新的直播名称",
  "device_id": "camera-002"
}
```

**响应示例**

成功 (200 OK):
```json
{
  "id": 1,
  "stream_key": "abc123def456",
  "name": "新的直播名称",
  "device_id": "camera-002",
  "status": "pushing",
  "protocol": "rtmp",
  "bitrate": 2500,
  "fps": 30,
  "last_frame_at": "2024-01-01T12:00:00Z",
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T12:05:00Z"
}
```

失败 (404 Not Found):
```json
{
  "error": "stream not found"
}
```

---

### 2.5 删除推流码

删除指定的推流码。

**接口地址**

```
DELETE /api/v1/streams/:key
```

**请求头**

```
Authorization: Bearer {token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |

**请求示例**

```
DELETE /api/v1/streams/abc123def456
```

**响应示例**

成功 (200 OK):
```json
{
  "message": "deleted"
}
```

---

### 2.6 强制断流

强制断开正在推流的流。

**接口地址**

```
POST /api/v1/streams/:key/kick
```

**请求头**

```
Authorization: Bearer {token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |

**请求示例**

```
POST /api/v1/streams/abc123def456/kick
```

**响应示例**

成功 (200 OK):
```json
{
  "message": "kicked"
}
```

失败 (404 Not Found):
```json
{
  "error": "stream not found"
}
```

---

## 3. 系统接口

### 3.1 健康检查

检查服务是否正常运行。

**接口地址**

```
GET /api/v1/system/health
```

**无需认证**

**响应示例**

成功 (200 OK):
```json
{
  "status": "ok"
}
```

---

### 3.2 系统统计

获取系统的统计信息。

**接口地址**

```
GET /api/v1/system/stats
```

**请求头**

```
Authorization: Bearer {token}
```

**响应示例**

成功 (200 OK):
```json
{
  "online_streams": 5,
  "total_streams": 100
}
```

**字段说明**

| 字段名 | 类型 | 说明 |
|--------|------|------|
| online_streams | integer | 当前在线推流数 |
| total_streams | integer | 总推流数 |

---

## 4. ZLMediaKit Hook 接口

> 这些接口由 ZLMediaKit 流媒体服务器调用，用于推流事件通知，无需认证。

### 4.1 推流开始回调

当有推流开始时，ZLMediaKit 会调用此接口进行鉴权。

**接口地址**

```
POST /api/v1/hooks/on_publish
```

**请求参数**

| 参数名 | 类型 | 说明 |
|--------|------|------|
| app | string | 应用名 (通常为 "live") |
| stream | string | 流 ID (stream_key) |
| schema | string | 推流协议 (rtmp/rtsp/srt) |
| mediaServerId | string | 媒体服务器 ID |
| ip | string | 推流客户端 IP |
| port | integer | 推流客户端端口 |
| params | string | 推流参数 |

**请求示例**

```json
{
  "app": "live",
  "stream": "abc123def456",
  "schema": "rtmp",
  "mediaServerId": "zlm-server-1",
  "ip": "192.168.1.100",
  "port": 12345,
  "params": ""
}
```

**响应示例**

允许推流 (200 OK):
```json
{
  "code": 0,
  "msg": "success"
}
```

拒绝推流 (200 OK, code != 0):
```json
{
  "code": -1,
  "msg": "invalid stream key"
}
```

**说明**

- 返回 `code: 0` 表示允许推流
- 返回 `code: -1` 或其他非 0 值表示拒绝推流

---

### 4.2 推流结束回调

当推流结束时，ZLMediaKit 会调用此接口通知。

**接口地址**

```
POST /api/v1/hooks/on_unpublish
```

**请求参数**

| 参数名 | 类型 | 说明 |
|--------|------|------|
| app | string | 应用名 |
| stream | string | 流 ID |
| schema | string | 推流协议 |
| mediaServerId | string | 媒体服务器 ID |

**请求示例**

```json
{
  "app": "live",
  "stream": "abc123def456",
  "schema": "rtmp",
  "mediaServerId": "zlm-server-1"
}
```

**响应示例**

成功 (200 OK):
```json
{
  "code": 0,
  "msg": "success"
}
```

---

### 4.3 流量统计回调

定期上报流量统计信息。

**接口地址**

```
POST /api/v1/hooks/on_flow_report
```

**请求参数**

| 参数名 | 类型 | 说明 |
|--------|------|------|
| app | string | 应用名 |
| stream | string | 流 ID |
| schema | string | 协议 |
| mediaServerId | string | 媒体服务器 ID |
| totalBytes | integer | 总流量 (字节) |
| duration | integer | 持续时间 (秒) |
| player | boolean | 是否为播放器 |
| totalBytesIn | integer | 上行流量 (字节) |
| totalBytesOut | integer | 下行流量 (字节) |

**请求示例**

```json
{
  "app": "live",
  "stream": "abc123def456",
  "schema": "rtmp",
  "mediaServerId": "zlm-server-1",
  "totalBytes": 1048576,
  "duration": 60,
  "player": false,
  "totalBytesIn": 1048576,
  "totalBytesOut": 0
}
```

**响应示例**

成功 (200 OK):
```json
{
  "code": 0,
  "msg": "success"
}
```

---

### 4.4 无人观看回调

当流无人观看时，ZLMediaKit 会调用此接口询问是否关闭流。

**接口地址**

```
POST /api/v1/hooks/on_stream_none_reader
```

**请求参数**

| 参数名 | 类型 | 说明 |
|--------|------|------|
| app | string | 应用名 |
| stream | string | 流 ID |
| schema | string | 协议 |
| mediaServerId | string | 媒体服务器 ID |

**请求示例**

```json
{
  "app": "live",
  "stream": "abc123def456",
  "schema": "rtmp",
  "mediaServerId": "zlm-server-1"
}
```

**响应示例**

不关闭流 (200 OK):
```json
{
  "code": 0,
  "close": false
}
```

关闭流 (200 OK):
```json
{
  "code": 0,
  "close": true
}
```

**说明**

- `close: false` 表示保持流继续运行
- `close: true` 表示关闭流

---

## 数据模型

### Stream (推流)

```typescript
{
  id: number              // 推流 ID
  stream_key: string      // 推流密钥 (唯一标识)
  name: string            // 推流名称
  device_id: string       // 设备 ID
  status: string          // 状态: idle (空闲) / pushing (推流中) / destroyed (已销毁)
  protocol: string        // 推流协议: rtmp / rtsp / srt
  bitrate: number         // 码率 (kbps)
  fps: number             // 帧率
  last_frame_at: string   // 最后一帧时间 (ISO 8601 格式)
  created_at: string      // 创建时间 (ISO 8601 格式)
  updated_at: string      // 更新时间 (ISO 8601 格式)
}
```

### User (用户)

```typescript
{
  id: number              // 用户 ID
  username: string        // 用户名
  role: string            // 角色: admin (管理员) / operator (操作员) / viewer (观察者)
  created_at: string      // 创建时间 (ISO 8601 格式)
  updated_at: string      // 更新时间 (ISO 8601 格式)
}
```

### StreamListResponse (推流列表响应)

```typescript
{
  total: number           // 总记录数
  streams: Stream[]       // 推流列表
}
```

### LoginResponse (登录响应)

```typescript
{
  token: string           // JWT Token
  user: User              // 用户信息
}
```

### HookResponse (Hook 响应)

```typescript
{
  code: number            // 响应码: 0 表示成功，非 0 表示失败
  msg: string             // 响应消息
}
```

---

## 错误码

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 / Token 无效或过期 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

### 业务错误码

错误响应格式：

```json
{
  "error": "错误描述信息"
}
```

常见错误信息：

| 错误信息 | 说明 |
|---------|------|
| invalid credentials | 用户名或密码错误 |
| stream not found | 推流不存在 |
| invalid token | Token 无效 |
| token expired | Token 已过期 |

---

## 使用示例

### cURL 示例

**1. 登录获取 Token**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

**2. 创建推流码**

```bash
curl -X POST http://localhost:8080/api/v1/streams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "会议室直播",
    "device_id": "camera-001"
  }'
```

**3. 获取推流列表**

```bash
curl -X GET "http://localhost:8080/api/v1/streams?status=pushing&page=1&pageSize=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**4. 强制断流**

```bash
curl -X POST http://localhost:8080/api/v1/streams/abc123def456/kick \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

### JavaScript 示例

**使用 Fetch API**

```javascript
// 登录
async function login(username, password) {
  const response = await fetch('http://localhost:8080/api/v1/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username, password }),
  });

  const data = await response.json();
  return data.token;
}

// 获取推流列表
async function getStreams(token, status = '', page = 1, pageSize = 20) {
  const params = new URLSearchParams({
    status,
    page: page.toString(),
    pageSize: pageSize.toString(),
  });

  const response = await fetch(`http://localhost:8080/api/v1/streams?${params}`, {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });

  return await response.json();
}

// 创建推流码
async function createStream(token, name, deviceId) {
  const response = await fetch('http://localhost:8080/api/v1/streams', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({
      name,
      device_id: deviceId,
    }),
  });

  return await response.json();
}

// 使用示例
(async () => {
  const token = await login('admin', 'admin123');
  const streams = await getStreams(token, 'pushing');
  console.log(streams);
})();
```

---

### Python 示例

**使用 requests 库**

```python
import requests

BASE_URL = "http://localhost:8080/api/v1"

# 登录
def login(username, password):
    response = requests.post(
        f"{BASE_URL}/auth/login",
        json={"username": username, "password": password}
    )
    return response.json()["token"]

# 获取推流列表
def get_streams(token, status="", page=1, page_size=20):
    headers = {"Authorization": f"Bearer {token}"}
    params = {"status": status, "page": page, "pageSize": page_size}

    response = requests.get(
        f"{BASE_URL}/streams",
        headers=headers,
        params=params
    )
    return response.json()

# 创建推流码
def create_stream(token, name, device_id=""):
    headers = {"Authorization": f"Bearer {token}"}
    data = {"name": name, "device_id": device_id}

    response = requests.post(
        f"{BASE_URL}/streams",
        headers=headers,
        json=data
    )
    return response.json()

# 使用示例
if __name__ == "__main__":
    token = login("admin", "admin123")
    streams = get_streams(token, status="pushing")
    print(streams)
```

---

## 配置 ZLMediaKit Hook

在 ZLMediaKit 配置文件中启用 Hook 回调：

```ini
[hook]
enable=1
on_publish=http://backend:8080/api/v1/hooks/on_publish
on_unpublish=http://backend:8080/api/v1/hooks/on_unpublish
on_flow_report=http://backend:8080/api/v1/hooks/on_flow_report
on_stream_none_reader=http://backend:8080/api/v1/hooks/on_stream_none_reader
```

---

## 默认账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | admin123 | 管理员 |

> ⚠️ **安全提示**: 生产环境请务必修改默认密码！

---

## 常见问题

### Q: Token 有效期是多久？

A: 默认 24 小时，可在配置文件中通过 `jwt.expireHour` 修改。

### Q: 如何刷新 Token？

A: 当前版本需要重新登录获取新 Token。

### Q: 推流地址中的 {server} 应该填什么？

A: 填写 ZLMediaKit 服务器的 IP 地址或域名。

### Q: 支持哪些推流协议？

A: 支持 RTMP、RTSP、SRT 三种协议。

### Q: 播放延迟有多低？

A: 使用 WebRTC 播放，延迟通常在 200ms 以内。

---

## 更新日志

### v1.0.0 (2024-01-01)

- 初始版本发布
- 支持推流管理
- 支持 ZLMediaKit Hook 集成
- 支持 JWT 认证

---

## 联系方式

- GitHub: https://github.com/yourusername/easy-stream
- Issues: https://github.com/yourusername/easy-stream/issues

---

**文档版本**: v1.0.0
**最后更新**: 2024-01-01

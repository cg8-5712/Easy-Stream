# Easy-Stream API 文档 v2.0

## 目录

- [基本信息](#基本信息)
- [认证机制](#认证机制)
- [认证接口](#1-认证接口)
- [推流管理接口](#2-推流管理接口)
- [分享链接接口](#3-分享链接接口)
- [系统接口](#4-系统接口)
- [ZLMediaKit Hook 接口](#5-zlmediakit-hook-接口)
- [数据模型](#数据模型)
- [错误码](#错误码)

---

## 基本信息

| 项目 | 说明 |
|------|------|
| Base URL | `http://localhost:8080/api/v1` |
| 认证方式 | JWT Bearer Token (Access Token + Refresh Token) |
| Content-Type | `application/json` |
| 字符编码 | UTF-8 |

---

## 认证机制

### Token 说明

本系统采用双 Token 机制：

| Token 类型 | 有效期 | 用途 |
|-----------|--------|------|
| Access Token | 2 小时 | 访问 API 接口 |
| Refresh Token | 7 天 | 刷新 Access Token |

### 认证流程

```
1. 用户登录 → 获取 access_token + refresh_token
2. 使用 access_token 访问 API
3. access_token 过期前，使用 refresh_token 获取新的 token 对
4. refresh_token 过期后，需要重新登录
```

### 请求头格式

```
Authorization: Bearer {access_token}
```

---

## 1. 认证接口

### 1.1 用户登录

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

**响应示例** (200 OK)
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "1_a1b2c3d4e5f6...",
  "expires_in": 7200,
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "phone": "13800138000",
    "real_name": "系统管理员",
    "avatar": "",
    "last_login_at": "2024-01-01T12:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

**错误响应** (401 Unauthorized)
```json
{
  "error": "invalid credentials"
}
```

---

### 1.2 刷新令牌

**接口地址**
```
POST /api/v1/auth/refresh
```

**请求参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| refresh_token | string | 是 | 刷新令牌 |

**请求示例**
```json
{
  "refresh_token": "1_a1b2c3d4e5f6..."
}
```

**响应示例** (200 OK)
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "1_new_token...",
  "expires_in": 7200
}
```

**错误响应** (401 Unauthorized)
```json
{
  "error": "invalid or expired refresh token"
}
```

---

### 1.3 用户登出

**接口地址**
```
POST /api/v1/auth/logout
```

**请求参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| refresh_token | string | 否 | 刷新令牌（用于撤销） |

**请求示例**
```json
{
  "refresh_token": "1_a1b2c3d4e5f6..."
}
```

**响应示例** (200 OK)
```json
{
  "message": "logged out"
}
```

---

### 1.4 获取当前用户信息

**接口地址**
```
GET /api/v1/auth/profile
```

**请求头**
```
Authorization: Bearer {access_token}
```

**响应示例** (200 OK)
```json
{
  "id": 1,
  "username": "admin",
  "email": "admin@example.com",
  "phone": "13800138000",
  "real_name": "系统管理员",
  "avatar": "",
  "last_login_at": "2024-01-01T12:00:00Z",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

---

## 2. 推流管理接口

### 2.1 获取推流列表

> 游客可访问，但只能看到公开且正在直播的内容（不含 stream_key）；管理员可看到所有直播记录（包括过去、正在进行和将来的）

**接口地址**
```
GET /api/v1/streams
```

**查询参数**

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| status | string | 否 | - | 状态过滤：`idle`/`pushing`/`ended`（仅管理员有效） |
| visibility | string | 否 | - | 可见性过滤：`public`/`private`（仅管理员有效） |
| time_range | string | 否 | - | 时间范围：`past`/`current`/`future`（仅管理员有效） |
| access_token | string | 否 | - | 私有直播访问令牌（游客可通过此参数获取已授权的私有直播） |
| page | int | 否 | 1 | 页码 |
| pageSize | int | 否 | 20 | 每页数量 |

**时间范围说明**

| 值 | 说明 |
|----|------|
| past | 已结束的直播（actual_end_time 不为空） |
| current | 正在进行的直播（status = pushing） |
| future | 未开始的直播（status = idle 且 scheduled_start_time > 当前时间） |

**请求示例**
```
GET /api/v1/streams?status=pushing&page=1&pageSize=20
GET /api/v1/streams?time_range=past&page=1&pageSize=20
GET /api/v1/streams?access_token=xyz789abc123...  (游客携带访问令牌)
```

**游客响应示例** (200 OK) - 不含 stream_key 和敏感信息
```json
{
  "total": 100,
  "streams": [
    {
      "id": 1,
      "name": "技术分享会",
      "description": "每周技术分享直播",
      "status": "pushing",
      "visibility": "public",
      "record_enabled": true,
      "record_status": "idle",
      "streamer_name": "张三",
      "streamer_contact": "13800138000",
      "scheduled_start_time": "2024-01-01T14:00:00Z",
      "scheduled_end_time": "2024-01-01T16:00:00Z",
      "actual_start_time": "2024-01-01T14:05:00Z",
      "actual_end_time": null,
      "current_viewers": 128,
      "total_viewers": 1520,
      "peak_viewers": 256,
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T14:05:00Z"
    }
  ]
}
```

**管理员响应示例** (200 OK) - 含 stream_key
```json
{
  "total": 100,
  "streams": [
    {
      "id": 1,
      "stream_key": "abc123def456",
      "name": "技术分享会",
      ...
    }
  ]
}
```

---

### 2.2 通过 ID 获取推流详情（游客/管理员）

> 游客可访问公开直播（不含 stream_key），私有直播需要 access_token；管理员返回完整信息

**接口地址**
```
GET /api/v1/streams/view/:id
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int | 是 | 直播 ID |

**查询参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| access_token | string | 否 | 私有直播访问令牌（游客访问私有直播时必填） |

**游客响应示例** (200 OK) - 不含 stream_key 和敏感信息
```json
{
  "id": 1,
  "name": "技术分享会",
  "description": "每周技术分享直播",
  "status": "pushing",
  "visibility": "public",
  "record_enabled": true,
  "record_status": "idle",
  "streamer_name": "张三",
  "streamer_contact": "13800138000",
  "scheduled_start_time": "2024-01-01T14:00:00Z",
  "scheduled_end_time": "2024-01-01T16:00:00Z",
  "actual_start_time": "2024-01-01T14:05:00Z",
  "actual_end_time": null,
  "current_viewers": 128,
  "total_viewers": 1520,
  "peak_viewers": 256,
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T14:05:00Z"
}
```

**错误响应** (403 Forbidden) - 私有直播无权限
```json
{
  "error": "private stream requires access token"
}
```

---

### 2.3 创建推流码（管理员）

**接口地址**
```
POST /api/v1/streams
```

**请求头**
```
Authorization: Bearer {access_token}
```

**请求参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 是 | 直播名称 |
| description | string | 否 | 直播描述 |
| device_id | string | 否 | 设备 ID |
| visibility | string | 是 | 可见性：`public`/`private` |
| record_enabled | bool | 否 | 是否开启录制，默认 false |
| streamer_name | string | 是 | 直播人员姓名 |
| streamer_contact | string | 否 | 直播人员联系方式 |
| scheduled_start_time | datetime | 是 | 预计开始时间 (ISO 8601) |
| scheduled_end_time | datetime | 是 | 预计结束时间 (ISO 8601) |
| auto_kick_delay | int | 否 | 超时断流延迟（分钟），默认 30 |

**请求示例**
```json
{
  "name": "技术分享会",
  "description": "每周技术分享直播",
  "device_id": "camera-001",
  "visibility": "public",
  "record_enabled": true,
  "streamer_name": "张三",
  "streamer_contact": "13800138000",
  "scheduled_start_time": "2024-01-01T14:00:00Z",
  "scheduled_end_time": "2024-01-01T16:00:00Z",
  "auto_kick_delay": 30
}
```

**私有直播请求示例**
```json
{
  "name": "内部会议",
  "visibility": "private",
  "streamer_name": "李四",
  "scheduled_start_time": "2024-01-01T14:00:00Z",
  "scheduled_end_time": "2024-01-01T16:00:00Z"
}
```

> 私有直播创建后会自动生成分享码，可通过分享码或分享链接访问。

**响应示例** (201 Created)
```json
{
  "id": 1,
  "stream_key": "abc123def456",
  "name": "技术分享会",
  "description": "每周技术分享直播",
  "device_id": "camera-001",
  "status": "idle",
  "visibility": "public",
  "record_enabled": true,
  "record_files": [],
  "protocol": "",
  "bitrate": 0,
  "fps": 0,
  "streamer_name": "张三",
  "streamer_contact": "13800138000",
  "scheduled_start_time": "2024-01-01T14:00:00Z",
  "scheduled_end_time": "2024-01-01T16:00:00Z",
  "auto_kick_delay": 30,
  "actual_start_time": null,
  "actual_end_time": null,
  "last_frame_at": null,
  "current_viewers": 0,
  "total_viewers": 0,
  "peak_viewers": 0,
  "created_by": 1,
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

**推流地址**
- RTMP: `rtmp://{server}:1935/live/{stream_key}`
- RTSP: `rtsp://{server}:554/live/{stream_key}`
- SRT: `srt://{server}:9000?streamid=#!::r=live/{stream_key},m=publish`
- HTTP-TS: `http://{server}:80/live/{stream_key}.live.ts` ⭐ 新增
- WebRTC: `POST http://{server}:8080/api/v1/streams/{stream_key}/webrtc-push` ⭐ 新增

**播放地址**
- WebRTC: `webrtc://{server}:8000/live/{stream_key}`
- HLS: `http://{server}:80/live/{stream_key}/hls.m3u8` ⭐ 新增
- HTTP-FLV: `http://{server}:80/live/{stream_key}.live.flv` ⭐ 新增
- WebSocket-FLV: `ws://{server}:80/live/{stream_key}.live.flv` ⭐ 新增

**回放地址**（录制开启且直播结束后可用）
- HTTP: `http://{server}:8080/recordings/{record_file}`

> 多次开关录制会生成多个文件，通过 `record_files` 数组获取所有录制文件路径。

---

### 2.3 通过 ID 获取推流详情（管理员）

**接口地址**
```
GET /api/v1/streams/id/:id
```

**请求头**
```
Authorization: Bearer {access_token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int | 是 | 直播 ID |

**请求示例**
```
GET /api/v1/streams/id/1
```

**响应示例** (200 OK)
```json
{
  "id": 1,
  "stream_key": "abc123def456",
  "name": "技术分享会",
  "description": "每周技术分享直播",
  "device_id": "camera-001",
  "status": "ended",
  "visibility": "public",
  "record_enabled": true,
  "record_files": [
    "/recordings/2024/01/01/abc123def456_001.mp4",
    "/recordings/2024/01/01/abc123def456_002.mp4"
  ],
  "protocol": "rtmp",
  "bitrate": 2500,
  "fps": 30,
  "streamer_name": "张三",
  "streamer_contact": "13800138000",
  "scheduled_start_time": "2024-01-01T14:00:00Z",
  "scheduled_end_time": "2024-01-01T16:00:00Z",
  "auto_kick_delay": 30,
  "actual_start_time": "2024-01-01T14:05:00Z",
  "actual_end_time": "2024-01-01T16:10:00Z",
  "last_frame_at": "2024-01-01T16:10:00Z",
  "current_viewers": 0,
  "total_viewers": 3280,
  "peak_viewers": 512,
  "created_by": 1,
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T16:10:00Z"
}
```

---

### 2.4 通过推流码获取推流详情

> 游客可访问公开直播；私有直播需要 access_token 或管理员权限

**接口地址**
```
GET /api/v1/streams/:key
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |

**查询参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| access_token | string | 否 | 私有直播访问令牌（通过分享码或分享链接获取） |

**请求示例**
```
GET /api/v1/streams/abc123def456
GET /api/v1/streams/abc123def456?access_token=xyz789...
```

**响应示例** (200 OK)
```json
{
  "id": 1,
  "stream_key": "abc123def456",
  "name": "技术分享会",
  "description": "每周技术分享直播",
  "device_id": "camera-001",
  "status": "pushing",
  "visibility": "public",
  "record_enabled": true,
  "record_files": ["/recordings/2024/01/01/abc123def456_001.mp4"],
  "protocol": "rtmp",
  "bitrate": 2500,
  "fps": 30,
  "streamer_name": "张三",
  "streamer_contact": "13800138000",
  "scheduled_start_time": "2024-01-01T14:00:00Z",
  "scheduled_end_time": "2024-01-01T16:00:00Z",
  "auto_kick_delay": 30,
  "actual_start_time": "2024-01-01T14:05:00Z",
  "actual_end_time": null,
  "last_frame_at": "2024-01-01T15:30:00Z",
  "current_viewers": 128,
  "total_viewers": 1520,
  "peak_viewers": 256,
  "created_by": 1,
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T14:05:00Z"
}
```

**错误响应**

404 Not Found:
```json
{
  "error": "stream not found"
}
```

403 Forbidden (私有直播无权限):
```json
{
  "error": "private stream requires access token"
}
```

---

### 2.5 验证分享码（游客）

> 游客通过此接口验证分享码，获取访问令牌

**接口地址**
```
POST /api/v1/shares/verify-code
```

**请求参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| share_code | string | 是 | 分享码（6位） |

**请求示例**
```json
{
  "share_code": "AB3XK9"
}
```

**响应示例** (200 OK)
```json
{
  "stream_id": 1,
  "access_token": "xyz789abc123...",
  "expires_at": "2024-01-01T16:00:00Z"
}
```

**使用方式**

获取 token 后，可通过以下方式访问私有直播：
1. 查询参数：`GET /api/v1/streams/view/:id?access_token={token}`
2. 获取列表时携带：`GET /api/v1/streams?access_token={token}`

**错误响应**

404 Not Found:
```json
{
  "error": "invalid share code"
}
```

410 Gone:
```json
{
  "error": "stream has ended"
}
```

403 Forbidden:
```json
{
  "error": "share code max uses reached"
}
```

---

### 2.6 添加分享码（管理员）

> 管理员为私有直播添加分享码

**接口地址**
```
POST /api/v1/streams/:key/share-code
```

**请求头**
```
Authorization: Bearer {access_token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |

**请求参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| max_uses | int | 否 | 最大使用次数（0表示不限制） |

**请求示例**
```json
{
  "max_uses": 10
}
```

**响应示例** (200 OK)
```json
{
  "id": 1,
  "stream_key": "abc123def456",
  "share_code": "AB3XK9",
  ...
}
```

---

### 2.7 重新生成分享码（管理员）

> 管理员可以重新生成私有直播的分享码，旧分享码将失效

**接口地址**
```
PUT /api/v1/streams/:key/share-code
```

**请求头**
```
Authorization: Bearer {access_token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |

**请求参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| max_uses | int | 否 | 最大使用次数（0表示不限制） |

**响应示例** (200 OK)
```json
{
  "id": 1,
  "stream_key": "abc123def456",
  "share_code": "XY7PQ2",
  ...
}
```

---

### 2.8 更新分享码使用次数（管理员）

**接口地址**
```
PATCH /api/v1/streams/:key/share-code
```

**请求头**
```
Authorization: Bearer {access_token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |

**请求参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| max_uses | int | 是 | 最大使用次数（0表示不限制） |

**请求示例**
```json
{
  "max_uses": 20
}
```

**响应示例** (200 OK)
```json
{
  "id": 1,
  "stream_key": "abc123def456",
  "share_code": "AB3XK9",
  "share_code_max_uses": 20,
  ...
}
```

---

### 2.9 删除分享码（管理员）

**接口地址**
```
DELETE /api/v1/streams/:key/share-code
```

**请求头**
```
Authorization: Bearer {access_token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |

**响应示例** (200 OK)
```json
{
  "message": "share code deleted"
}
```

---

### 2.10 更新推流信息（管理员）

**接口地址**
```
PUT /api/v1/streams/:key
```

**请求头**
```
Authorization: Bearer {access_token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |

**请求参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 否 | 直播名称 |
| description | string | 否 | 直播描述 |
| device_id | string | 否 | 设备 ID |
| visibility | string | 否 | 可见性：`public`/`private` |
| share_code_max_uses | int | 否 | 分享码最大使用次数（0表示不限制） |
| record_enabled | bool | 否 | 是否开启录制（支持推流中动态修改） |
| streamer_name | string | 否 | 直播人员姓名 |
| streamer_contact | string | 否 | 直播人员联系方式 |
| scheduled_start_time | datetime | 否 | 预计开始时间 |
| scheduled_end_time | datetime | 否 | 预计结束时间 |
| auto_kick_delay | int | 否 | 超时断流延迟（分钟） |

**动态录制说明**

`record_enabled` 参数支持在推流过程中动态修改：

| 场景 | 行为 |
|------|------|
| 推流中开启录制 | 从当前时间点开始录制，不包含之前的内容 |
| 推流中关闭录制 | 立即停止录制，已录制内容保留 |
| 多次开关录制 | 每次开启会生成新的录制文件 |

> ⚠️ **注意**: 推流中途开启录制，录制文件只包含开启后的内容。如需完整录制，请在创建直播时开启。

**请求示例**
```json
{
  "name": "技术分享会（更新）",
  "scheduled_end_time": "2024-01-01T17:00:00Z",
  "auto_kick_delay": 60
}
```

**响应示例** (200 OK)
```json
{
  "id": 1,
  "stream_key": "abc123def456",
  "name": "技术分享会（更新）",
  ...
}
```

---

### 2.11 删除推流码（管理员）

**接口地址**
```
DELETE /api/v1/streams/:key
```

**请求头**
```
Authorization: Bearer {access_token}
```

**响应示例** (200 OK)
```json
{
  "message": "deleted"
}
```

---

### 2.12 强制断流（管理员）

**接口地址**
```
POST /api/v1/streams/:key/kick
```

**请求头**
```
Authorization: Bearer {access_token}
```

**响应示例** (200 OK)
```json
{
  "message": "kicked"
}
```

---

## 3. 分享链接接口

> 管理员可以为私有直播创建分享链接，用户通过分享链接可以直接获取访问权限

### 3.1 获取分享链接列表（管理员）

**接口地址**
```
GET /api/v1/streams/:key/share-links
```

**请求头**
```
Authorization: Bearer {access_token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |

**响应示例** (200 OK)
```json
{
  "total": 1,
  "links": [
    {
      "id": 1,
      "stream_key": "test-stream-004",
      "token": "abc123xyz789def456...",
      "max_uses": 100,
      "used_count": 25,
      "created_by": 1,
      "created_at": "2024-01-01T10:00:00Z"
    }
  ]
}
```

---

### 3.2 创建分享链接（管理员）

**接口地址**
```
POST /api/v1/streams/:key/share-links
```

**请求头**
```
Authorization: Bearer {access_token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |

**请求参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| max_uses | int | 否 | 最大使用次数（0表示不限制） |

**请求示例**
```json
{
  "max_uses": 100
}
```

**响应示例** (201 Created)
```json
{
  "id": 1,
  "token": "abc123xyz789def456...",
  "share_url": "/share/abc123xyz789def456...",
  "max_uses": 100,
  "used_count": 0
}
```

---

### 3.3 更新分享链接使用次数（管理员）

**接口地址**
```
PATCH /api/v1/streams/share-links/:id
```

**请求头**
```
Authorization: Bearer {access_token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int | 是 | 分享链接 ID |

**请求参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| max_uses | int | 是 | 最大使用次数（0表示不限制） |

**请求示例**
```json
{
  "max_uses": 200
}
```

**响应示例** (200 OK)
```json
{
  "id": 1,
  "stream_key": "test-stream-004",
  "token": "abc123xyz789...",
  "max_uses": 200,
  "used_count": 25,
  "created_by": 1,
  "created_at": "2024-01-01T10:00:00Z"
}
```

---

### 3.4 删除分享链接（管理员）

**接口地址**
```
DELETE /api/v1/streams/share-links/:id
```

**请求头**
```
Authorization: Bearer {access_token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int | 是 | 分享链接 ID |

**响应示例** (200 OK)
```json
{
  "message": "share link deleted"
}
```

---

### 3.5 验证分享链接（游客）

> 用户通过分享链接访问时，验证 token 获取访问权限

**接口地址**
```
GET /api/v1/shares/link/:token
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| token | string | 是 | 分享链接 token |

**响应示例** (200 OK)
```json
{
  "stream_id": 5,
  "access_token": "xyz789abc123...",
  "expires_at": "2024-01-01T18:00:00Z"
}
```

**错误响应**

404 Not Found:
```json
{
  "error": "invalid share link"
}
```

410 Gone:
```json
{
  "error": "stream has ended"
}
```

403 Forbidden:
```json
{
  "error": "share link max uses reached"
}
```

---

## 4. 系统接口

### 4.1 健康检查

**接口地址**
```
GET /api/v1/system/health
```

**无需认证**

**功能说明**

返回系统详细健康状态，包括：
- 主应用状态（healthy/degraded/unhealthy）
- PostgreSQL 数据库连接状态
- Redis 连接状态
- ZLMediaKit 流媒体服务器状态
- 网络连接状态（DNS 和外网）

**状态说明**

| 总体状态 | 说明 |
|---------|------|
| healthy | 所有服务正常 |
| degraded | 非关键服务异常（ZLMediaKit、网络） |
| unhealthy | 关键服务异常（PostgreSQL、Redis） |

| 服务状态 | 说明 |
|---------|------|
| up | 服务正常，连接和认证均成功 |
| down | 服务不可用，无法连接 |
| auth_failed | 认证失败，密码或密钥错误 |

**响应示例** (200 OK)
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "uptime": "2h 30m 15s",
  "version": "2.0",
  "services": {
    "postgresql": {
      "status": "up",
      "latency": "1.2ms"
    },
    "redis": {
      "status": "up",
      "latency": "0.5ms"
    },
    "zlmediakit": {
      "status": "up",
      "latency": "15.3ms"
    }
  },
  "network": {
    "dns": {
      "status": "up",
      "latency": "25ms"
    },
    "internet": {
      "status": "up",
      "latency": "120ms"
    }
  }
}
```

**降级响应示例** (200 OK)
```json
{
  "status": "degraded",
  "timestamp": "2024-01-01T12:00:00Z",
  "uptime": "2h 30m 15s",
  "version": "2.0",
  "services": {
    "postgresql": {
      "status": "up",
      "latency": "1.2ms"
    },
    "redis": {
      "status": "up",
      "latency": "0.5ms"
    },
    "zlmediakit": {
      "status": "down",
      "latency": "5s",
      "message": "connection failed: dial tcp 127.0.0.1:80: connect: connection refused"
    }
  },
  "network": {
    "dns": {
      "status": "up",
      "latency": "25ms"
    },
    "internet": {
      "status": "up",
      "latency": "120ms"
    }
  }
}
```

**不可用响应示例** (503 Service Unavailable)
```json
{
  "status": "unhealthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "uptime": "2h 30m 15s",
  "version": "2.0",
  "services": {
    "postgresql": {
      "status": "down",
      "latency": "5s",
      "message": "connection failed: dial tcp 127.0.0.1:5432: connect: connection refused"
    },
    "redis": {
      "status": "up",
      "latency": "0.5ms"
    },
    "zlmediakit": {
      "status": "up",
      "latency": "15.3ms"
    }
  },
  "network": {
    "dns": {
      "status": "up",
      "latency": "25ms"
    },
    "internet": {
      "status": "up",
      "latency": "120ms"
    }
  }
}
```

**认证失败响应示例** (503 Service Unavailable)
```json
{
  "status": "unhealthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "uptime": "2h 30m 15s",
  "version": "2.0",
  "services": {
    "postgresql": {
      "status": "up",
      "latency": "1.2ms"
    },
    "redis": {
      "status": "auth_failed",
      "latency": "0.8ms",
      "message": "authentication failed: invalid password"
    },
    "zlmediakit": {
      "status": "auth_failed",
      "latency": "10ms",
      "message": "authentication failed: invalid secret"
    }
  },
  "network": {
    "dns": {
      "status": "up",
      "latency": "25ms"
    },
    "internet": {
      "status": "up",
      "latency": "120ms"
    }
  }
}
```

**验证内容说明**

| 服务 | 验证内容 |
|------|---------|
| PostgreSQL | 执行 `SELECT 1` 查询验证连接和用户名/密码 |
| Redis | 执行 SET/GET/DEL 操作验证连接和密码 |
| ZLMediaKit | 调用 `getServerConfig` API 验证连接和 Secret |

---

### 4.2 系统统计（管理员）

**接口地址**
```
GET /api/v1/system/stats
```

**请求头**
```
Authorization: Bearer {access_token}
```

**响应示例** (200 OK)
```json
{
  "online_streams": 5,
  "total_streams": 100
}
```

---

## 5. ZLMediaKit Hook 接口

> 这些接口由 ZLMediaKit 流媒体服务器调用，无需认证

### 5.1 推流开始回调

```
POST /api/v1/hooks/on_publish
```

**请求示例**
```json
{
  "app": "live",
  "stream": "abc123def456",
  "schema": "rtmp",
  "mediaServerId": "zlm-server-1",
  "ip": "192.168.1.100",
  "port": 12345
}
```

**验证逻辑**

系统会对推流请求进行以下验证：

| 验证项 | 说明 |
|-------|------|
| stream_key 存在性 | 推流码必须是通过管理接口创建的有效推流码 |
| 状态检查 | 推流码状态不能为 `ended` |

**响应示例**

验证通过 (允许推流):
```json
{
  "code": 0,
  "msg": "success"
}
```

验证失败 (拒绝推流):
```json
{
  "code": -1,
  "msg": "stream not found"
}
```

| 错误信息 | 说明 |
|---------|------|
| stream not found | 推流码不存在 |
| stream expired | 推流码已结束 |

> ⚠️ **安全说明**: 无效或不存在的推流码将被拒绝，ZLMediaKit 会自动断开该推流连接。

### 5.2 推流结束回调

```
POST /api/v1/hooks/on_unpublish
```

> 推流结束时，系统会自动清理该直播的所有访问令牌（分享码和分享链接生成的令牌都会失效）

### 5.3 流量统计回调

```
POST /api/v1/hooks/on_flow_report
```

### 5.4 无人观看回调

```
POST /api/v1/hooks/on_stream_none_reader
```

### 5.5 播放开始回调

```
POST /api/v1/hooks/on_play
```

**请求示例**
```json
{
  "app": "live",
  "stream": "abc123def456",
  "schema": "rtmp",
  "mediaServerId": "zlm-server-1",
  "ip": "192.168.1.100",
  "port": 12345,
  "id": "player-unique-id"
}
```

**说明**: 当有观众开始观看直播时，ZLMediaKit 会调用此接口。系统会自动增加当前观看人数和累计观看人次。

### 5.6 播放器断开回调

```
POST /api/v1/hooks/on_player_disconnect
```

**请求示例**
```json
{
  "app": "live",
  "stream": "abc123def456",
  "schema": "rtmp",
  "mediaServerId": "zlm-server-1",
  "ip": "192.168.1.100",
  "port": 12345,
  "id": "player-unique-id"
}
```

**说明**: 当观众离开直播时，ZLMediaKit 会调用此接口。系统会自动减少当前观看人数。

---

## 数据模型

### User (用户)

```typescript
{
  id: number              // 用户 ID
  username: string        // 用户名
  role: string            // 角色: admin / operator / viewer
  email: string           // 邮箱
  phone: string           // 电话
  real_name: string       // 真实姓名
  avatar: string          // 头像 URL
  department: string      // 部门
  status: string          // 状态: active / disabled
  last_login_at: string   // 最后登录时间
  created_at: string      // 创建时间
  updated_at: string      // 更新时间
}
```

### Stream (推流)

```typescript
{
  id: number                    // 推流 ID
  stream_key: string            // 推流密钥
  name: string                  // 直播名称
  description: string           // 直播描述
  device_id: string             // 设备 ID
  status: string                // 状态: idle / pushing / ended
  visibility: string            // 可见性: public / private
  share_code: string            // 分享码（私有直播自动生成，6位）
  share_code_max_uses: number   // 分享码最大使用次数（0表示不限制）
  share_code_used_count: number // 分享码已使用次数
  record_enabled: boolean       // 是否开启录制
  record_status: string         // 录制状态: idle / recording / stopped / failed
  record_files: RecordFile[]    // 录制文件列表（包含完整元数据）
  protocol: string              // 协议: rtmp / rtsp / srt / webrtc
  bitrate: number               // 码率 (kbps)
  fps: number                   // 帧率
  streamer_name: string         // 直播人员姓名
  streamer_contact: string      // 直播人员联系方式
  scheduled_start_time: string  // 预计开始时间
  scheduled_end_time: string    // 预计结束时间
  auto_kick_delay: number       // 超时断流延迟（分钟）
  actual_start_time: string     // 实际开始时间
  actual_end_time: string       // 实际结束时间
  last_frame_at: string         // 最后一帧时间
  // 观看统计
  current_viewers: number       // 当前观看人数
  total_viewers: number         // 累计观看人次
  peak_viewers: number          // 峰值观看人数
  created_by: number            // 创建者用户 ID
  created_at: string            // 创建时间
  updated_at: string            // 更新时间
}
```

### RecordFile (录制文件)

```typescript
{
  file_name: string             // 文件名
  file_path: string             // 文件路径
  file_size: number             // 文件大小（字节）
  duration: number              // 录制时长（秒）
  start_time: number            // 录制开始时间戳
  time_len: number              // 录制时长（秒，ZLM 返回）
  created_at: string            // 文件创建时间
  urls: {                       // 各存储的访问URL
    local?: string              // 本地路径
    s3?: string                 // S3 URL
    cos?: string                // 腾讯云 COS URL
    oss?: string                // 阿里云 OSS URL
  }
}
```

### StreamPublicView (游客可见的直播信息)

游客访问公开直播时，返回的字段子集（不含敏感信息）：

```typescript
{
  id: number                    // 推流 ID
  name: string                  // 直播名称
  description: string           // 直播描述
  status: string                // 状态: idle / pushing / ended
  visibility: string            // 可见性: public / private
  record_enabled: boolean       // 是否开启录制
  record_status: string         // 录制状态: idle / recording / stopped / failed
  streamer_name: string         // 直播人员姓名
  streamer_contact: string      // 直播人员联系方式
  scheduled_start_time: string  // 预计开始时间
  scheduled_end_time: string    // 预计结束时间
  actual_start_time: string     // 实际开始时间
  actual_end_time: string       // 实际结束时间
  current_viewers: number       // 当前观看人数
  total_viewers: number         // 累计观看人次
  peak_viewers: number          // 峰值观看人数
  created_at: string            // 创建时间
  updated_at: string            // 更新时间
}
```

**注意**：游客不可见的字段包括：
- `stream_key` - 推流密钥
- `device_id` - 设备 ID
- `record_files` - 录制文件列表
- `protocol` - 推流协议
- `bitrate` - 码率
- `fps` - 帧率
- `auto_kick_delay` - 超时延迟
- `last_frame_at` - 最后一帧时间
- `created_by` - 创建者 ID

### ShareLink (分享链接)

```typescript
{
  id: number              // 分享链接 ID
  stream_key: string      // 关联的直播 stream_key
  token: string           // 分享链接 token（64位）
  max_uses: number        // 最大使用次数（0表示不限制）
  used_count: number      // 已使用次数
  created_by: number      // 创建者用户 ID
  created_at: string      // 创建时间
  stream: Stream          // 关联的直播信息（可选）
}
```

### StreamAccessToken (私有直播访问令牌)

```typescript
{
  stream_id: number       // 直播 ID
  access_token: string    // 访问令牌
  expires_at: string      // 过期时间
}
```

---

## 错误码

| HTTP 状态码 | 说明 |
|------------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 / Token 无效 |
| 403 | 禁止访问（如私有直播无权限） |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

### 常见错误信息

| 错误信息 | 说明 |
|---------|------|
| invalid credentials | 用户名或密码错误 |
| invalid or expired refresh token | 刷新令牌无效或已过期 |
| stream not found | 推流不存在 |
| private stream requires access token | 私有直播需要访问令牌 |
| invalid share code | 分享码无效 |
| invalid share link | 分享链接无效 |
| share code max uses reached | 分享码使用次数已达上限 |
| share link max uses reached | 分享链接使用次数已达上限 |
| stream has ended | 直播已结束 |
| only private streams support sharing | 仅私有直播支持分享功能 |

---

## 超时自动断流机制

系统会每分钟检查正在推流的直播，当满足以下条件时自动断流：

```
当前时间 > 预计结束时间 + 超时延迟时间
```

例如：
- 预计结束时间：16:00
- 超时延迟：30 分钟
- 自动断流时间：16:30

管理员可在创建或更新直播时设置 `auto_kick_delay` 参数。

---

## 私有直播访问机制

私有直播支持两种访问方式：

### 1. 分享码访问

- 私有直播创建时自动生成 8 位分享码
- 用户输入分享码验证后获取访问令牌
- 支持设置最大使用次数限制
- 直播结束后分享码自动失效
- 管理员可重新生成分享码（旧分享码失效）

**访问流程**:
```
1. 用户获取分享码（由管理员提供）
2. 调用 POST /api/v1/shares/verify-code 验证分享码
3. 获取 access_token
4. 使用 access_token 访问直播内容
```

### 2. 分享链接访问

- 管理员手动创建分享链接（64位 token）
- 用户通过链接直接获取访问权限
- 支持设置最大使用次数限制
- 直播结束后分享链接自动失效
- 可创建多个分享链接

**访问流程**:
```
1. 管理员创建分享链接
2. 用户点击分享链接（包含 token 参数）
3. 前端调用 GET /api/v1/shares/link/:token 验证 token
4. 获取 access_token
5. 使用 access_token 访问直播内容
```

---

## 默认账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | admin123 | 管理员 |

> ⚠️ **安全提示**: 生产环境请务必修改默认密码！

---

**文档版本**: v2.2
**最后更新**: 2026-02-04

---

## 6. 播放地址接口 ⭐ 新增

### 6.1 获取播放地址

**接口地址**
```
GET /api/v1/streams/:key/play-urls
```

**权限说明**
- 游客：可访问公开直播的播放地址
- 管理员：可访问所有直播的播放地址和推流地址

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |

**请求头（可选）**
```
Authorization: Bearer {access_token}
```

**响应示例** (200 OK) - 游客访问公开直播

```json
{
  "stream_id": 123,
  "stream_key": "abc123def456",
  "stream_name": "测试直播",
  "status": "pushing",
  "play_urls": {
    "webrtc": "webrtc://192.168.1.9:8000/live/abc123def456",
    "hls": "http://192.168.1.9:80/live/abc123def456/hls.m3u8",
    "http_flv": "http://192.168.1.9:80/live/abc123def456.live.flv",
    "ws_flv": "ws://192.168.1.9:80/live/abc123def456.live.flv"
  }
}
```

**响应示例** (200 OK) - 管理员访问

```json
{
  "stream_id": 123,
  "stream_key": "abc123def456",
  "stream_name": "测试直播",
  "status": "pushing",
  "play_urls": {
    "webrtc": "webrtc://192.168.1.9:8000/live/abc123def456",
    "hls": "http://192.168.1.9:80/live/abc123def456/hls.m3u8",
    "http_flv": "http://192.168.1.9:80/live/abc123def456.live.flv",
    "ws_flv": "ws://192.168.1.9:80/live/abc123def456.live.flv"
  },
  "push_urls": {
    "rtmp": "rtmp://192.168.1.9:1935/live/abc123def456",
    "rtsp": "rtsp://192.168.1.9:554/live/abc123def456",
    "srt": "srt://192.168.1.9:9000?streamid=#!::r=live/abc123def456,m=publish",
    "http_ts": "http://192.168.1.9:80/live/abc123def456.live.ts"
  }
}
```

**播放协议说明**

| 协议 | 延迟 | 兼容性 | 适用场景 |
|------|------|--------|---------|
| webrtc | < 200ms | 现代浏览器 | 实时互动、低延迟要求 |
| hls | 6-15s | 所有设备 | 兼容性优先、移动端 |
| http_flv | 1-3s | 需 flv.js | 平衡延迟和兼容性 |
| ws_flv | 1-3s | 需 flv.js | 实时性要求较高 |

**推流协议说明**（仅管理员可见）

| 协议 | 端口 | 适用场景 |
|------|------|---------|
| rtmp | 1935 | OBS 推流（推荐） |
| rtsp | 554 | 摄像头设备 |
| srt | 9000/UDP | 低延迟、抗丢包 |
| http_ts | 80 | 防火墙友好 |

**错误响应**

404 Not Found:
```json
{
  "error": "stream not found"
}
```

---

## 7. WebRTC 推流接口 ⭐ 新增

### 7.1 WebRTC 推流

**接口地址**
```
POST /api/v1/streams/:key/webrtc-push
```

**权限说明**
- 无需管理员认证
- 只需要有效的 stream_key（与 RTMP/RTSP/SRT 推流一致）

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |

**请求参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| sdp | string | 是 | WebRTC SDP Offer |

**请求示例**
```json
{
  "sdp": "v=0\r\no=- 1234567890 1234567890 IN IP4 0.0.0.0\r\ns=WebRTC Push\r\nt=0 0\r\na=group:BUNDLE 0 1\r\na=msid-semantic: WMS stream\r\nm=audio 9 UDP/TLS/RTP/SAVPF 111\r\nc=IN IP4 0.0.0.0\r\n..."
}
```

**响应示例** (200 OK)
```json
{
  "sdp": "v=0\r\no=- 9876543210 9876543210 IN IP4 192.168.1.9\r\ns=ZLMediaKit\r\nt=0 0\r\na=group:BUNDLE 0 1\r\na=msid-semantic: WMS stream\r\nm=audio 9 UDP/TLS/RTP/SAVPF 111\r\nc=IN IP4 192.168.1.9\r\n..."
}
```

**错误响应**

404 Not Found:
```json
{
  "error": "stream not found"
}
```

403 Forbidden:
```json
{
  "error": "stream has expired"
}
```

**使用示例（JavaScript）**

```javascript
async function startWebRTCPush(streamKey) {
  // 获取本地媒体流
  const localStream = await navigator.mediaDevices.getUserMedia({
    video: true,
    audio: true
  });

  // 创建 RTCPeerConnection
  const pc = new RTCPeerConnection({
    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
  });

  // 添加本地流
  localStream.getTracks().forEach(track => {
    pc.addTrack(track, localStream);
  });

  // 创建 Offer
  const offer = await pc.createOffer();
  await pc.setLocalDescription(offer);

  // 发送 Offer 到服务器
  const response = await fetch(`http://localhost:8080/api/v1/streams/${streamKey}/webrtc-push`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ sdp: offer.sdp })
  });

  const { sdp: answerSDP } = await response.json();

  // 设置远端 SDP Answer
  await pc.setRemoteDescription({
    type: 'answer',
    sdp: answerSDP
  });

  console.log('WebRTC 推流成功！');
}
```

**说明**
- WebRTC 推流会自动触发 `on_publish` Hook 进行验证
- 推流协议字段 `schema` 将显示为 `webrtc`
- 推流成功后,观众可通过任意播放协议观看

---

## 8. 录制文件管理接口 ⭐ 新增

### 8.1 获取所有录制文件列表（管理员）

**接口地址**
```
GET /api/v1/records
```

**请求头**
```
Authorization: Bearer {access_token}
```

**查询参数**

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| pageSize | int | 否 | 20 | 每页数量（最大100） |

**请求示例**
```
GET /api/v1/records?page=1&pageSize=20
```

**响应示例** (200 OK)
```json
{
  "total": 50,
  "streams": [
    {
      "id": 1,
      "stream_key": "abc123def456",
      "name": "技术分享会",
      "record_status": "stopped",
      "record_files": [
        {
          "file_name": "2024-01-15/12-30-45.mp4",
          "file_path": "live/abc123def456/2024-01-15/12-30-45.mp4",
          "file_size": 104857600,
          "duration": 3600.5,
          "start_time": 1705315845,
          "created_at": "2024-01-15T12:30:45Z",
          "urls": {
            "download": "/api/v1/records/abc123def456/2024-01-15/12-30-45.mp4/download",
            "s3": "https://bucket.s3.amazonaws.com/records/abc123def456/2024-01-15/12-30-45.mp4",
            "cos": "https://bucket.cos.ap-guangzhou.myqcloud.com/records/abc123def456/2024-01-15/12-30-45.mp4"
          }
        }
      ]
    }
  ]
}
```

**字段说明**

| 字段 | 说明 |
|------|------|
| file_name | 文件名（相对路径） |
| file_path | 完整文件路径 |
| file_size | 文件大小（字节） |
| duration | 录制时长（秒） |
| start_time | 录制开始时间戳 |
| urls.download | Main Server 代理下载URL（推荐） |
| urls.s3/cos/oss | 对象存储直接访问URL |

---

### 8.2 获取指定直播的录制文件（管理员）

**接口地址**
```
GET /api/v1/records/:key
```

**请求头**
```
Authorization: Bearer {access_token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |

**请求示例**
```
GET /api/v1/records/abc123def456
```

**响应示例** (200 OK)
```json
{
  "stream_id": 1,
  "stream_key": "abc123def456",
  "stream_name": "技术分享会",
  "record_files": [
    {
      "file_name": "2024-01-15/12-30-45.mp4",
      "file_path": "live/abc123def456/2024-01-15/12-30-45.mp4",
      "file_size": 104857600,
      "duration": 3600.5,
      "start_time": 1705315845,
      "created_at": "2024-01-15T12:30:45Z",
      "urls": {
        "download": "/api/v1/records/abc123def456/2024-01-15/12-30-45.mp4/download",
        "s3": "https://bucket.s3.amazonaws.com/records/abc123def456/2024-01-15/12-30-45.mp4"
      }
    }
  ],
  "record_status": "stopped"
}
```

**错误响应**

404 Not Found:
```json
{
  "error": "stream not found"
}
```

---

### 8.3 播放录制文件（管理员）⭐ 新增

**接口地址**
```
GET /api/v1/records/:key/play/*filepath
```

**功能说明**
- 支持在线播放录制文件（video/mp4）
- 支持 HTTP Range 请求，可拖拽进度条
- 适合前端 `<video>` 标签直接播放

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |
| filepath | string | 是 | 文件路径（支持多级目录） |

**请求头**
```
Authorization: Bearer {access_token}
```

**请求示例**
```
GET /api/v1/records/abc123def456/play/2024-01-15/12-30-45.mp4
GET /api/v1/records/abc123def456/play/2026-02-12-10-46-24-0.mp4
```

**Range 请求示例**
```
GET /api/v1/records/abc123def456/play/2024-01-15/12-30-45.mp4
Range: bytes=0-1023
```

**响应说明**

| 模式 | 行为 |
|------|------|
| 本地模式 | 直接返回文件内容，自动支持 Range 请求 |
| 远程模式 | 代理转发 Range 请求到远程服务器 |

**完整响应** (200 OK)
```
Content-Type: video/mp4
Accept-Ranges: bytes
Content-Length: 104857600

[文件二进制内容]
```

**Range 响应** (206 Partial Content)
```
Content-Type: video/mp4
Accept-Ranges: bytes
Content-Range: bytes 0-1023/104857600
Content-Length: 1024

[指定范围的文件内容]
```

**前端使用示例**
```html
<video controls width="800">
  <source src="/api/v1/records/abc123def456/play/2024-01-15/12-30-45.mp4" type="video/mp4">
</video>
```

**错误响应**

404 Not Found:
```json
{
  "error": "stream not found"
}
```

404 Not Found:
```json
{
  "error": "file not found"
}
```

401 Unauthorized:
```json
{
  "error": "authentication required"
}
```

---

### 8.4 下载录制文件（管理员）

**接口地址**
```
GET /api/v1/records/:key/download/*filepath
```

**功能说明**
- 强制下载录制文件（浏览器会提示保存）
- 不支持 Range 请求
- 适合用户保存文件到本地

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |
| filepath | string | 是 | 文件路径（支持多级目录） |

**请求头**
```
Authorization: Bearer {access_token}
```

**请求示例**
```
GET /api/v1/records/abc123def456/download/2024-01-15/12-30-45.mp4
GET /api/v1/records/abc123def456/download/2026-02-12-10-46-24-0.mp4
```

**响应说明**

| 模式 | 行为 |
|------|------|
| 本地模式 | 直接返回文件内容（200 OK） |
| 远程模式 | 代理转发文件内容（200 OK） |

**响应** (200 OK)
```
Content-Type: application/octet-stream
Content-Disposition: attachment; filename=2024-01-15/12-30-45.mp4
Content-Transfer-Encoding: binary

[文件二进制内容]
```

**错误响应**

404 Not Found:
```json
{
  "error": "stream not found"
}
```

404 Not Found:
```json
{
  "error": "file not found"
}
```

401 Unauthorized (私有直播未认证):
```json
{
  "error": "authentication required"
}
```

403 Forbidden (私有直播非创建者):
```json
{
  "error": "access denied"
}
```

**前端使用示例**

```javascript
// 公开直播下载
function downloadPublicRecord(streamKey, fileName) {
  const url = `/api/v1/records/${streamKey}/download/${fileName}`;
  window.location.href = url;
}

// 私有直播下载（需要token）
function downloadPrivateRecord(streamKey, fileName, token) {
  const url = `/api/v1/records/${streamKey}/download/${fileName}`;

  fetch(url, {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  })
  .then(response => {
    if (response.redirected) {
      // 远程模式：重定向到对象存储
      window.location.href = response.url;
    } else {
      // 本地模式：直接下载
      return response.blob();
    }
  })
  .then(blob => {
    if (blob) {
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = fileName;
      a.click();
    }
  });
}
```

---

### 8.5 删除录制文件（管理员）

**接口地址**
```
DELETE /api/v1/records/:key/delete/*filepath
```

**请求头**
```
Authorization: Bearer {access_token}
```

**路径参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key | string | 是 | 推流密钥 |
| filepath | string | 是 | 文件路径（支持多级目录） |

**请求示例**
```
DELETE /api/v1/records/abc123def456/delete/2024-01-15/12-30-45.mp4
DELETE /api/v1/records/abc123def456/delete/2026-02-12-10-46-24-0.mp4
```

**删除逻辑**

系统会依次删除：
1. **ZLM原始文件**（本地模式）或标记删除（远程模式）
2. **所有对象存储备份**（S3/COS/OSS等）
3. **数据库记录**

**响应示例** (200 OK)
```json
{
  "message": "record file deleted successfully"
}
```

**错误响应**

404 Not Found:
```json
{
  "error": "stream not found"
}
```

500 Internal Server Error:
```json
{
  "error": "failed to delete file"
}
```

**注意事项**

| 模式 | 删除行为 |
|------|---------|
| 本地模式 | 删除ZLM本地文件 + 对象存储备份 + 数据库记录 |
| 远程模式 | 只删除对象存储备份 + 数据库记录（ZLM文件需手动清理） |

> ⚠️ **远程模式说明**：由于ZLM不提供删除文件的API，远程模式下无法删除ZLM服务器上的物理文件。建议配置定时任务自动清理过期文件。

---

## 录制文件URL说明

每个录制文件包含多个访问URL：

| URL类型 | 说明 | 优先级 |
|---------|------|--------|
| download | Main Server代理下载（推荐） | 最高 |
| s3/cos/oss | 对象存储直接访问 | 中 |
| 无URL | 仅ZLM本地存储 | 最低 |

**推荐使用download URL的原因**：
- ✅ 统一的权限控制（自动检查私有直播权限）
- ✅ 智能源选择（优先对象存储，降低ZLM负载）
- ✅ 安全性高（不直接暴露ZLM或对象存储）
- ✅ 灵活性好（可随时切换下载源）

---

**文档版本**: v2.3
**最后更新**: 2026-02-13

---

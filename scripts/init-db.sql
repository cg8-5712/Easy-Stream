-- Easy-Stream 数据库初始化脚本

-- 创建推流表
CREATE TABLE IF NOT EXISTS streams (
    id              SERIAL PRIMARY KEY,
    stream_key      VARCHAR(64) UNIQUE NOT NULL,
    name            VARCHAR(128),
    device_id       VARCHAR(64),
    status          VARCHAR(16) DEFAULT 'idle',
    protocol        VARCHAR(16),
    bitrate         INTEGER DEFAULT 0,
    fps             INTEGER DEFAULT 0,
    last_frame_at   TIMESTAMP,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_streams_status ON streams(status);
CREATE INDEX IF NOT EXISTS idx_streams_device_id ON streams(device_id);

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id              SERIAL PRIMARY KEY,
    username        VARCHAR(64) UNIQUE NOT NULL,
    password_hash   VARCHAR(256) NOT NULL,
    role            VARCHAR(16) DEFAULT 'viewer',
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

-- 创建操作日志表
CREATE TABLE IF NOT EXISTS operation_logs (
    id              SERIAL PRIMARY KEY,
    user_id         INTEGER REFERENCES users(id),
    action          VARCHAR(64) NOT NULL,
    target_type     VARCHAR(32),
    target_id       VARCHAR(64),
    detail          JSONB,
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_logs_user_id ON operation_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_logs_created_at ON operation_logs(created_at);

-- 插入默认管理员用户 (密码: admin123)
-- 密码使用 bcrypt 加密
INSERT INTO users (username, password_hash, role)
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin')
ON CONFLICT (username) DO NOTHING;

-- Easy-Stream 数据库初始化脚本

-- 创建版本迁移表（必须最先创建）
CREATE TABLE IF NOT EXISTS schema_migrations (
    version         INTEGER PRIMARY KEY,
    description     VARCHAR(256) NOT NULL,
    applied_at      TIMESTAMP DEFAULT NOW()
);

COMMENT ON TABLE schema_migrations IS '数据库版本迁移记录表';
COMMENT ON COLUMN schema_migrations.version IS '版本号';
COMMENT ON COLUMN schema_migrations.description IS '迁移描述';
COMMENT ON COLUMN schema_migrations.applied_at IS '应用时间';

-- 插入初始版本记录
INSERT INTO schema_migrations (version, description)
VALUES (1, '初始化数据库结构')
ON CONFLICT (version) DO NOTHING;

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id              SERIAL PRIMARY KEY,
    username        VARCHAR(64) UNIQUE NOT NULL,
    password_hash   VARCHAR(256) NOT NULL,
    email           VARCHAR(128),
    phone           VARCHAR(32),
    real_name       VARCHAR(64),
    avatar          VARCHAR(256),
    last_login_at   TIMESTAMP,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- 创建推流表
CREATE TABLE IF NOT EXISTS streams (
    id                      SERIAL PRIMARY KEY,
    stream_key              VARCHAR(64) UNIQUE NOT NULL,
    name                    VARCHAR(128) NOT NULL,
    description             TEXT,
    device_id               VARCHAR(64),
    status                  VARCHAR(16) DEFAULT 'idle',
    visibility              VARCHAR(16) DEFAULT 'public',
    share_code              VARCHAR(8),
    share_code_max_uses     INTEGER DEFAULT 0,
    share_code_used_count   INTEGER DEFAULT 0,
    record_enabled          BOOLEAN DEFAULT FALSE,
    record_files            JSONB DEFAULT '[]',
    protocol                VARCHAR(16),
    bitrate                 INTEGER DEFAULT 0,
    fps                     INTEGER DEFAULT 0,
    streamer_name           VARCHAR(64) NOT NULL,
    streamer_contact        VARCHAR(128),
    scheduled_start_time    TIMESTAMP NOT NULL,
    scheduled_end_time      TIMESTAMP NOT NULL,
    auto_kick_delay         INTEGER DEFAULT 30,
    actual_start_time       TIMESTAMP,
    actual_end_time         TIMESTAMP,
    last_frame_at           TIMESTAMP,
    current_viewers         INTEGER DEFAULT 0,
    total_viewers           INTEGER DEFAULT 0,
    peak_viewers            INTEGER DEFAULT 0,
    created_by              INTEGER REFERENCES users(id),
    created_at              TIMESTAMP DEFAULT NOW(),
    updated_at              TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_streams_status ON streams(status);
CREATE INDEX IF NOT EXISTS idx_streams_visibility ON streams(visibility);
CREATE INDEX IF NOT EXISTS idx_streams_device_id ON streams(device_id);
CREATE INDEX IF NOT EXISTS idx_streams_created_by ON streams(created_by);
CREATE INDEX IF NOT EXISTS idx_streams_scheduled_start ON streams(scheduled_start_time);
CREATE INDEX IF NOT EXISTS idx_streams_scheduled_end ON streams(scheduled_end_time);
CREATE INDEX IF NOT EXISTS idx_streams_record_enabled ON streams(record_enabled);
CREATE UNIQUE INDEX IF NOT EXISTS idx_streams_share_code ON streams(share_code) WHERE share_code IS NOT NULL;

-- 创建分享链接表
CREATE TABLE IF NOT EXISTS share_links (
    id              SERIAL PRIMARY KEY,
    stream_key      VARCHAR(64) NOT NULL REFERENCES streams(stream_key) ON DELETE CASCADE,
    token           VARCHAR(64) UNIQUE NOT NULL,
    max_uses        INTEGER DEFAULT 0,
    used_count      INTEGER DEFAULT 0,
    created_by      INTEGER REFERENCES users(id),
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_share_links_stream_key ON share_links(stream_key);
CREATE INDEX IF NOT EXISTS idx_share_links_token ON share_links(token);

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
CREATE INDEX IF NOT EXISTS idx_logs_action ON operation_logs(action);

-- 插入默认管理员用户 (密码: admin123)
-- 密码使用 bcrypt 加密
INSERT INTO users (username, password_hash, real_name)
VALUES ('admin', '$2a$10$wxWds7XBPNLDPJu2/Fiaj.ryW0ym01KKiYHtrb.56NLsExEewNQxS', '系统管理员')
ON CONFLICT (username) DO NOTHING;

-- 插入测试操作员用户 (密码: operator123)
INSERT INTO users (username, password_hash, real_name)
VALUES ('operator', '$2a$10$YourHashHere', '操作员')
ON CONFLICT (username) DO NOTHING;

-- 添加注释
COMMENT ON TABLE users IS '用户表';
COMMENT ON COLUMN users.id IS '用户ID';
COMMENT ON COLUMN users.username IS '用户名';
COMMENT ON COLUMN users.password_hash IS '密码哈希';
COMMENT ON COLUMN users.email IS '邮箱';
COMMENT ON COLUMN users.phone IS '电话';
COMMENT ON COLUMN users.real_name IS '真实姓名';
COMMENT ON COLUMN users.avatar IS '头像URL';
COMMENT ON COLUMN users.last_login_at IS '最后登录时间';

COMMENT ON TABLE streams IS '推流表';
COMMENT ON COLUMN streams.id IS '推流ID';
COMMENT ON COLUMN streams.stream_key IS '推流密钥';
COMMENT ON COLUMN streams.name IS '推流名称';
COMMENT ON COLUMN streams.description IS '推流描述';
COMMENT ON COLUMN streams.device_id IS '设备ID';
COMMENT ON COLUMN streams.status IS '状态：idle/pushing/ended';
COMMENT ON COLUMN streams.visibility IS '可见性：public/private';
COMMENT ON COLUMN streams.share_code IS '分享码（私有直播自动生成）';
COMMENT ON COLUMN streams.share_code_max_uses IS '分享码最大使用次数（0表示无限制）';
COMMENT ON COLUMN streams.share_code_used_count IS '分享码已使用次数';
COMMENT ON COLUMN streams.record_enabled IS '是否开启录制';
COMMENT ON COLUMN streams.record_files IS '录制文件路径列表（JSON数组）';
COMMENT ON COLUMN streams.protocol IS '推流协议：rtmp/rtsp/srt';
COMMENT ON COLUMN streams.bitrate IS '码率（kbps）';
COMMENT ON COLUMN streams.fps IS '帧率';
COMMENT ON COLUMN streams.streamer_name IS '直播人员姓名';
COMMENT ON COLUMN streams.streamer_contact IS '直播人员联系方式';
COMMENT ON COLUMN streams.scheduled_start_time IS '预计开始时间';
COMMENT ON COLUMN streams.scheduled_end_time IS '预计结束时间';
COMMENT ON COLUMN streams.auto_kick_delay IS '超时自动断流延迟（分钟）';
COMMENT ON COLUMN streams.actual_start_time IS '实际开始时间';
COMMENT ON COLUMN streams.actual_end_time IS '实际结束时间';
COMMENT ON COLUMN streams.last_frame_at IS '最后一帧时间';
COMMENT ON COLUMN streams.current_viewers IS '当前观看人数';
COMMENT ON COLUMN streams.total_viewers IS '累计观看人次';
COMMENT ON COLUMN streams.peak_viewers IS '峰值观看人数';
COMMENT ON COLUMN streams.created_by IS '创建者用户ID';

COMMENT ON TABLE share_links IS '分享链接表';
COMMENT ON COLUMN share_links.id IS '链接ID';
COMMENT ON COLUMN share_links.stream_key IS '关联的直播stream_key';
COMMENT ON COLUMN share_links.token IS '分享链接token';
COMMENT ON COLUMN share_links.max_uses IS '最大使用次数（0表示无限制）';
COMMENT ON COLUMN share_links.used_count IS '已使用次数';
COMMENT ON COLUMN share_links.created_by IS '创建者用户ID';

COMMENT ON TABLE operation_logs IS '操作日志表';
COMMENT ON COLUMN operation_logs.id IS '日志ID';
COMMENT ON COLUMN operation_logs.user_id IS '操作用户ID';
COMMENT ON COLUMN operation_logs.action IS '操作动作';
COMMENT ON COLUMN operation_logs.target_type IS '目标类型';
COMMENT ON COLUMN operation_logs.target_id IS '目标ID';
COMMENT ON COLUMN operation_logs.detail IS '详细信息（JSON）';

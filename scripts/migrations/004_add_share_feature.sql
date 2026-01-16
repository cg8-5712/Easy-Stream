-- 迁移脚本: 添加分享功能
-- 移除密码机制，添加分享码和分享链接功能

-- 移除 password 字段
ALTER TABLE streams DROP COLUMN IF EXISTS password;

-- 添加分享码字段
ALTER TABLE streams ADD COLUMN IF NOT EXISTS share_code VARCHAR(8);
ALTER TABLE streams ADD COLUMN IF NOT EXISTS share_code_max_uses INTEGER DEFAULT 0;
ALTER TABLE streams ADD COLUMN IF NOT EXISTS share_code_used_count INTEGER DEFAULT 0;

-- 为分享码创建唯一索引（仅对非空值）
CREATE UNIQUE INDEX IF NOT EXISTS idx_streams_share_code ON streams(share_code) WHERE share_code IS NOT NULL;

-- 创建分享链接表
CREATE TABLE IF NOT EXISTS share_links (
    id SERIAL PRIMARY KEY,
    stream_id INTEGER NOT NULL REFERENCES streams(id) ON DELETE CASCADE,
    token VARCHAR(64) UNIQUE NOT NULL,
    max_uses INTEGER DEFAULT 0,
    used_count INTEGER DEFAULT 0,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_share_links_stream_id ON share_links(stream_id);
CREATE INDEX IF NOT EXISTS idx_share_links_token ON share_links(token);

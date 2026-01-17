-- 迁移脚本: 添加断流时间字段
-- 用于记录最后断流时间，配合 auto_kick_delay 实现自动结束功能

-- 添加最后断流时间字段
ALTER TABLE streams ADD COLUMN IF NOT EXISTS last_unpublish_at TIMESTAMP;

-- 添加注释
COMMENT ON COLUMN streams.last_unpublish_at IS '最后断流时间，用于计算自动结束';

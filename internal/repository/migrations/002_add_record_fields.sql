-- 迁移脚本：添加录制相关字段
-- 版本：002
-- 日期：2026-01-13

-- 添加录制开关字段
ALTER TABLE streams ADD COLUMN IF NOT EXISTS record_enabled BOOLEAN DEFAULT FALSE;

-- 添加录制文件列表字段（JSONB 数组）
ALTER TABLE streams ADD COLUMN IF NOT EXISTS record_files JSONB DEFAULT '[]';

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_streams_record_enabled ON streams(record_enabled);

-- 更新注释
COMMENT ON COLUMN streams.status IS '状态：idle/pushing/ended';
COMMENT ON COLUMN streams.record_enabled IS '是否开启录制';
COMMENT ON COLUMN streams.record_files IS '录制文件路径列表（JSON数组）';

-- 将旧的 destroyed 状态更新为 ended
UPDATE streams SET status = 'ended' WHERE status = 'destroyed';

-- 添加录制状态字段和扩展录制文件元数据
-- 迁移脚本 006

-- 添加 record_status 字段
ALTER TABLE streams ADD COLUMN IF NOT EXISTS record_status VARCHAR(20) DEFAULT 'idle';

-- 更新 record_files 字段注释（从简单数组改为对象数组）
COMMENT ON COLUMN streams.record_files IS '录制文件列表（包含完整元数据：文件名、大小、时长等）';

-- 更新现有数据：将 record_enabled 为 false 的设置为 idle
UPDATE streams SET record_status = 'idle' WHERE record_enabled = false;

-- 更新现有数据：将正在推流且 record_enabled 为 true 的设置为 recording
UPDATE streams SET record_status = 'recording' WHERE status = 'pushing' AND record_enabled = true;

-- 更新现有数据：将已结束且 record_enabled 为 true 的设置为 stopped
UPDATE streams SET record_status = 'stopped' WHERE status = 'ended' AND record_enabled = true;

-- 添加检查约束
ALTER TABLE streams DROP CONSTRAINT IF EXISTS streams_record_status_check;
ALTER TABLE streams ADD CONSTRAINT streams_record_status_check
    CHECK (record_status IN ('idle', 'recording', 'stopped', 'failed'));

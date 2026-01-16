-- 迁移脚本：添加观看统计字段
-- 版本：003
-- 日期：2026-01-13

-- 添加观看统计字段
ALTER TABLE streams ADD COLUMN IF NOT EXISTS current_viewers INTEGER DEFAULT 0;
ALTER TABLE streams ADD COLUMN IF NOT EXISTS total_viewers INTEGER DEFAULT 0;
ALTER TABLE streams ADD COLUMN IF NOT EXISTS peak_viewers INTEGER DEFAULT 0;

-- 添加注释
COMMENT ON COLUMN streams.current_viewers IS '当前观看人数';
COMMENT ON COLUMN streams.total_viewers IS '累计观看人次';
COMMENT ON COLUMN streams.peak_viewers IS '峰值观看人数';

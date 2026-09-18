-- 迁移 76：MySQL-only（放宽 #74/#75 的 VARCHAR(19) 时间列至 32，
-- gtime 微秒超长问题——见 mysql/76）。SQLite 侧该列为 TEXT 天然不受限，
-- 本文件仅为双轨同号占位，无操作
SELECT 1;

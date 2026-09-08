-- 同 sqlite/64：MySQL 侧时区对齐说明性迁移（无实际 DDL 需求）。
-- MySQL baseline 的 DATETIME DEFAULT CURRENT_TIMESTAMP / ON UPDATE
-- CURRENT_TIMESTAMP 均按会话时区生成；DSN 强制 loc=Local（部署样例已含）
-- → 会话时区=服务器本地 → 与统一口径天然一致，无需改表。
-- 此文件存在的意义：双轨编号对齐 + 把口径决策留在迁移历史里可追溯。
SELECT 1;

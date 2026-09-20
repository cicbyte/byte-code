-- 同 sqlite/83：测试表排序规则对齐（生产 MySQL 实报 Error 1267）
-- 背景：迁移 77 建表只写了 DEFAULT CHARSET=utf8mb4 未写 COLLATE，MySQL 8
-- 服务器/库默认给 utf8mb4_0900_ai_ci；基线与其余迁移均为 utf8mb4_unicode_ci。
-- test_run_cases.external_key 与 test_cases.external_key 的跨表等值比较
-- （用例执行统计 join）是全库首处两表字符串直接比较，暴露排序规则混用。
-- 本迁移把两张表整体 CONVERT 对齐到基线口径（列与索引随之重建）；
-- 77 的建表语句已补 COLLATE（新装库不再偏航）。
ALTER TABLE `test_runs` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
ALTER TABLE `test_run_cases` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

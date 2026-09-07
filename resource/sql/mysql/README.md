# MySQL 迁移目录（双轨维护说明）

- `01_baseline.sql`：全新库一次性建表 + 种子数据，等价于 sqlite 目录迁移 01-62 的终态。
  由 SQLite 实际 schema 机械转换 + 人工审校生成（生成脚本为一次性工具未入库）。
- **后续增量迁移必须双轨**：`resource/sql/sqlite/63_xxx.sql` 与 `resource/sql/mysql/63_xxx.sql`
  同号同名成对出现，两边语义一致、各方言合法。
- `_migrations` 按库独立记账：SQLite 库记录 sqlite 目录文件、MySQL 库记录 mysql 目录文件，
  互不干扰；MySQL 基线记录为 `01_baseline.sql`，之后从 63 起续编。
- 版本要求：MySQL >= 8.0.13 / MariaDB >= 10.2（TEXT 函数式默认值）。
- 与 SQLite 的有意差异（详见基线头部注释）：updated_at 用列属性 ON UPDATE 而非触发器、
  不迁移外键声明（「0=无引用」约定）、FTS5 表不迁移（搜索走 LIKE）、
  索引涉及列用 VARCHAR、字符集 utf8mb4_unicode_ci。
- MySQL 迁移文件内**不要写触发器/存储过程**（执行器按分号切分逐条执行，无 DELIMITER 支持）。

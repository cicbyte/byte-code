package dbinit

import (
	"strings"
	"testing"
)

func TestMigrationNumber(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"01_categories.sql", 1},
		{"09_seed_roles.sql", 9},
		{"38_migrate.sql", 38},
		{"40_drop.sql", 40},
		{"41_db_hardening.sql", 41},
		// 字典序陷阱：字符串比较会把 100 排在 40 之前，数字排序必须纠正
		{"100_future.sql", 100},
		{"9_before_ten.sql", 9},
		{"no_prefix.sql", 0x7fffffff}, // 无数字前缀排最后
	}
	for _, c := range cases {
		if got := migrationNumber(c.in); got != c.want {
			t.Errorf("migrationNumber(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestSplitSQLStatements(t *testing.T) {
	input := "-- 行注释; 含分号\n" +
		"CREATE TABLE `t` (`a` INT, `b` VARCHAR(10) DEFAULT 'x;y');\n" +
		"/* 块注释; 分号 */\n" +
		"INSERT INTO t VALUES (1, 'a''b;c');\n" +
		"INSERT INTO t VALUES (2, 'it''s;ok');\n" +
		"CREATE INDEX idx ON t (`b`);\n"
	stmts := splitSQLStatements(input)
	want := []string{
		"CREATE TABLE `t` (`a` INT, `b` VARCHAR(10) DEFAULT 'x;y')",
		"INSERT INTO t VALUES (1, 'a''b;c')",
		"INSERT INTO t VALUES (2, 'it''s;ok')",
		"CREATE INDEX idx ON t (`b`)",
	}
	if len(stmts) != len(want) {
		t.Fatalf("语句数 %d != %d: %q", len(stmts), len(want), stmts)
	}
	for i := range want {
		if stmts[i] != want[i] {
			t.Errorf("第 %d 条不匹配:\n got: %s\nwant: %s", i, stmts[i], want[i])
		}
	}
	for _, s := range stmts {
		if strings.Contains(s, "--") || strings.Contains(s, "/*") {
			t.Errorf("语句残留注释: %s", s)
		}
	}
}

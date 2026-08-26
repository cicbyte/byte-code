package dbinit

import "testing"

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

package escape

import "testing"

func TestLike(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"plain", "plain"},
		{"100%", `100|%`},
		{"a_b", `a|_b`},
		{"pipe|char", `pipe||char`},
		// 字面量反斜杠不再转义：ESCAPE '|' 下它就是普通字符（双方言一致）
		{`back\slash`, `back\slash`},
		{`%_\|`, `|%|_\||`},
		{"", ""},
		{"中文关键词", "中文关键词"},
	}
	for _, c := range cases {
		if got := Like(c.in); got != c.want {
			t.Errorf("Like(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

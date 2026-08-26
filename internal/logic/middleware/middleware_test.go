package middleware

import "testing"

func TestParseTarget(t *testing.T) {
	cases := []struct {
		path       string
		entityType string
		entity     int
	}{
		{"/api/v1/tasks/456", "tasks", 456},
		// 嵌套路径取最深的实体（task 456 属于 project 123）
		{"/api/v1/projects/123/tasks/456", "tasks", 456},
		{"/api/v1/tasks/456/claim", "tasks", 456},
		{"/api/v1/sprints/9/burndown", "sprints", 9},
		{"/api/v1/projects", "", 0},
	}
	for _, c := range cases {
		got := parseTarget(c.path)
		if got.entityType != c.entityType || got.entityId != c.entity {
			t.Errorf("parseTarget(%q) = {%s %d}, want {%s %d}",
				c.path, got.entityType, got.entityId, c.entityType, c.entity)
		}
	}
}

func TestLastResourceSegment(t *testing.T) {
	cases := []struct {
		path, want string
	}{
		{"/api/v1/projects", "projects"},
		{"/api/v1/projects/5/tasks", "tasks"},
		{"/api/v1/projects/5/db-tables", "db-tables"},
		// 边缘输入：纯前缀路径返回最后的非数字段（审计 target_type 记为 v1，无害）
		{"/api/v1/", "v1"},
	}
	for _, c := range cases {
		if got := lastResourceSegment(c.path); got != c.want {
			t.Errorf("lastResourceSegment(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestCreatedIdFromBody(t *testing.T) {
	cases := []struct {
		name, body string
		want       int
	}{
		{"正常创建", `{"code":0,"message":"OK","data":{"id":42}}`, 42},
		{"失败响应", `{"code":50,"message":"err","data":null}`, 0},
		{"无id", `{"code":0,"data":{}}`, 0},
		{"空体", ``, 0},
		{"非JSON", `Not Found`, 0},
	}
	for _, c := range cases {
		if got := createdIdFromBody([]byte(c.body)); got != c.want {
			t.Errorf("%s: createdIdFromBody(%q) = %d, want %d", c.name, c.body, got, c.want)
		}
	}
}

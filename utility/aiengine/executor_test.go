package aiengine

import (
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestEvictOldToolResults(t *testing.T) {
	// 构造消息序列：sys + user + [assistant(toolcall) + tool] × 4 轮
	var msgs []*schema.Message
	msgs = append(msgs,
		&schema.Message{Role: schema.System, Content: "sys"},
		&schema.Message{Role: schema.User, Content: "task"})
	for r := 0; r < 4; r++ {
		msgs = append(msgs, &schema.Message{Role: schema.Assistant, Content: "",
			ToolCalls: []schema.ToolCall{{ID: string(rune('a' + r)), Function: schema.FunctionCall{Name: "x"}}}})
		msgs = append(msgs, &schema.Message{Role: schema.Tool, Content: "BIG-RESULT-ROUND-" + string(rune('0'+r)), ToolCallID: string(rune('a' + r))})
	}

	evictOldToolResults(msgs, 2)

	for i, m := range msgs {
		if m.Role != schema.Tool {
			continue
		}
		r := (i - 2) / 2 // 轮次
		if r >= 2 {       // 最近两轮保留
			if !strings.Contains(m.Content, "BIG-RESULT-ROUND") {
				t.Fatalf("第 %d 轮工具结果不应被截断: %q", r, m.Content)
			}
		} else {
			if strings.Contains(m.Content, "BIG-RESULT") {
				t.Fatalf("第 %d 轮工具结果应被占位替换: %q", r, m.Content)
			}
			if !strings.Contains(m.Content, "truncated") {
				t.Fatalf("占位符缺失: %q", m.Content)
			}
		}
	}
	// 非 Tool 消息不动
	if msgs[0].Content != "sys" || msgs[1].Content != "task" {
		t.Fatal("非工具消息被误改")
	}

	// 轮数不足 keepN：全保留
	var few []*schema.Message
	few = append(few,
		&schema.Message{Role: schema.Assistant, ToolCalls: []schema.ToolCall{{ID: "a"}}},
		&schema.Message{Role: schema.Tool, Content: "ONLY", ToolCallID: "a"})
	evictOldToolResults(few, 2)
	if few[1].Content != "ONLY" {
		t.Fatal("不足保留轮数时不应截断")
	}
}

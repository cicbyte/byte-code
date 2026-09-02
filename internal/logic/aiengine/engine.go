package aiengine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cicbyte/byte-code/utility/aiengine"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
)

// ==================== AI 执行引擎 ====================
// 定时扫描待认领任务 → AI Agent 分析并执行 → 提交结果

const (
	engineScanInterval = "0 */2 * * * *" // 每 2 分钟扫描一次
	engineMaxTasks     = 3               // 每轮最多处理 3 个任务
)

var (
	engineOnce sync.Once
	engineStop = make(chan struct{})
	isRunning = false
	runMu     sync.Mutex
)

// StartEngine 启动 AI 执行引擎
func StartEngine(ctx context.Context) {
	engineOnce.Do(func() {
		isRunning = true
		g.Log().Info(ctx, "AI 执行引擎启动")

		_, err := gcron.AddSingleton(ctx, engineScanInterval, func(ctx context.Context) {
			scanAndExecute(ctx)
		})
		if err != nil {
			g.Log().Errorf(ctx, "AI 引擎定时任务注册失败: %v", err)
		}

		// 启动立即执行一次
		go scanAndExecute(ctx)
	})
}

// StopEngine 停止引擎
func StopEngine(ctx context.Context) {
	runMu.Lock()
	defer runMu.Unlock()
	if isRunning {
		isRunning = false
		close(engineStop)
		g.Log().Info(ctx, "AI 执行引擎停止")
	}
}

// IsRunning 引擎是否在运行
func IsRunning() bool {
	runMu.Lock()
	defer runMu.Unlock()
	return isRunning
}

// scanAndExecute 扫描并执行一轮
func scanAndExecute(ctx context.Context) {
	runMu.Lock()
	if !isRunning {
		runMu.Unlock()
		return
	}
	runMu.Unlock()

	// 1. 获取引擎配置（模型地址/Key）
	cfg, err := getEngineConfig(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "AI 引擎: 获取配置失败: %v", err)
		return
	}
	if cfg == nil {
		g.Log().Info(ctx, "AI 引擎: 未配置模型，跳过本轮")
		return
	}

	// 2. 找到所有 open 状态且分配给 AI 用户的任务
	tasks, err := getClaimableTasks(ctx, engineMaxTasks)
	if err != nil {
		g.Log().Warningf(ctx, "AI 引擎: 查询可认领任务失败: %v", err)
		return
	}
	if len(tasks) == 0 {
		return
	}

	g.Log().Infof(ctx, "AI 引擎: 本轮发现 %d 个可处理任务", len(tasks))

	// 3. 逐个处理
	for _, task := range tasks {
		processTask(ctx, cfg, task)
	}
}

// EngineConfig 引擎配置
type EngineConfig struct {
	BaseURL string // OpenAI-compatible API 地址
	ApiKey  string
	Model   string // 模型名
}

func getEngineConfig(ctx context.Context) (*EngineConfig, error) {
	var cfgs []struct{ Key, Value string }
	err := g.DB().Model("sys_config").Ctx(ctx).
		WhereIn("key", g.Slice{"ai_engine_base_url", "ai_engine_api_key", "ai_engine_model"}).
		Scan(&cfgs)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for _, c := range cfgs {
		m[c.Key] = c.Value
	}

	if m["ai_engine_base_url"] == "" || m["ai_engine_model"] == "" {
		return nil, nil // 未配置
	}

	return &EngineConfig{
		BaseURL: m["ai_engine_base_url"],
		ApiKey:  m["ai_engine_api_key"],
		Model:   m["ai_engine_model"],
	}, nil
}

// ClaimableTask 可被 AI 认领的任务
type ClaimableTask struct {
	Id        int
	ProjectId int
	Title     string
	Desc      string
	AssigneeId int // AI 用户 ID
	AssigneeName string
}

func getClaimableTasks(ctx context.Context, limit int) ([]ClaimableTask, error) {
	var tasks []ClaimableTask
	err := g.DB().Model("tasks t").Ctx(ctx).
		LeftJoin("sys_users u", "t.assignee_id = u.id").
		Fields("t.id, t.project_id, t.title, t.description, t.assignee_id, COALESCE(u.real_name, u.username) as assignee_name").
		Where("t.status", "open").
		Where("t.assignee_id > 0").
		Where("u.type", "ai").
		Where("u.status", 1).
		Order("t.created_at ASC").
		Limit(limit).
		Scan(&tasks)
	return tasks, err
}

// processTask 处理单个任务
func processTask(ctx context.Context, cfg *EngineConfig, task ClaimableTask) {
	g.Log().Infof(ctx, "AI 引擎: 处理任务 #%d「%s」（AI: %s）", task.Id, task.Title, task.AssigneeName)

	// 标记 in_progress（条件更新防竞态）
	result, err := g.DB().Model("tasks").Ctx(ctx).
		Where("id", task.Id).
		Where("status", "open").
		Data(g.Map{"status": "in_progress"}).
		Update()
	if err != nil {
		g.Log().Warningf(ctx, "AI 引擎: 认领任务 %d 失败: %v", task.Id, err)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		g.Log().Info(ctx, "AI 引擎: 任务已被其他进程认领，跳过")
		return
	}

	// 记录 AI 执行日志
	logId := recordAiLog(ctx, task.AssigneeId, task.Id, "claim", fmt.Sprintf("AI 认领任务「%s」", task.Title), "success")

	// 调用 AI Agent 执行
	output, err := aiengine.ExecuteTask(ctx, &aiengine.ExecuteConfig{
		BaseURL:  cfg.BaseURL,
		ApiKey:   cfg.ApiKey,
		Model:    cfg.Model,
		AIUserId: task.AssigneeId,
		Task: aiengine.TaskContext{
			Id:        task.Id,
			ProjectId: task.ProjectId,
			Title:     task.Title,
			Desc:      task.Desc,
		},
	})

	if err != nil {
		g.Log().Errorf(ctx, "AI 引擎: 任务 %d 执行失败: %v", task.Id, err)
		recordAiLog(ctx, task.AssigneeId, task.Id, "error", err.Error(), "failed")
		// 执行失败回退状态
		g.DB().Model("tasks").Ctx(ctx).
			Where("id", task.Id).
			Where("status", "in_progress").
			Data(g.Map{"status": "open"}).Update()
		return
	}

	// 提交结果
	g.DB().Model("tasks").Ctx(ctx).
		Where("id", task.Id).
		Data(g.Map{
			"status":       "review",
			"artifacts":    output,
			"completed_at": time.Now().Format("2006-01-02 15:04:05"),
		}).Update()

	recordAiLog(ctx, task.AssigneeId, task.Id, "complete", output, "success")
	g.Log().Infof(ctx, "AI 引擎: 任务 %d 完成，已提交审核", task.Id)
	_ = logId
}

func recordAiLog(ctx context.Context, aiUserId, taskId int, action, detail, status string) int {
	result, err := g.DB().Model("ai_execution_logs").Ctx(ctx).Insert(g.Map{
		"ai_user_id": aiUserId,
		"task_id":    taskId,
		"action":     action,
		"detail":     detail,
		"status":     status,
		"created_at": time.Now().Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		return 0
	}
	id, _ := result.LastInsertId()
	return int(id)
}

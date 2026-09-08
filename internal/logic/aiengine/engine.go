package aiengine

import (
	"context"
	"fmt"
	"github.com/cicbyte/byte-code/utility/dbutil"
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
	// enabledConfigKey 运行开关持久化：重启进程后据此自启（此前只存内存，重启即丢）
	enabledConfigKey = "ai_engine_enabled"
)

var (
	engineMu     sync.Mutex
	engineEntry  *gcron.Entry // gcron 具名条目：start=注册 / stop=移除，可反复启停
	engineCancel context.CancelFunc
	engineCtx    context.Context
	isRunning    bool
)

// StartEngine 启动引擎：注册扫描条目（已注册则忽略），持久化开关。
// 旧实现用 sync.Once 导致停后再启 cron 永不注册（引擎静默死亡），已弃用
func StartEngine(ctx context.Context) {
	engineMu.Lock()
	defer engineMu.Unlock()
	if isRunning {
		return
	}
	engineCtx, engineCancel = context.WithCancel(context.Background())
	if _, err := gcron.AddSingleton(engineCtx, engineScanInterval, func(ctx context.Context) {
		scanAndExecute(ctx)
	}); err != nil {
		g.Log().Errorf(ctx, "AI 引擎定时任务注册失败: %v", err)
		return
	}
	isRunning = true
	persistEnabled(ctx, true)
	g.Log().Info(ctx, "AI 执行引擎启动")

	// 启动立即执行一次
	go scanAndExecute(engineCtx)
}

// StopEngine 停止引擎：移除条目 + 取消执行 ctx（进行中的 LLM 调用随 ctx 中断）
func StopEngine(ctx context.Context) {
	engineMu.Lock()
	defer engineMu.Unlock()
	if !isRunning {
		return
	}
	if engineCancel != nil {
		engineCancel()
	}
	if engineEntry != nil {
		engineEntry.Stop()
		engineEntry = nil
	}
	isRunning = false
	persistEnabled(ctx, false)
	g.Log().Info(ctx, "AI 执行引擎停止")
}

// IsRunning 引擎是否在运行
func IsRunning() bool {
	engineMu.Lock()
	defer engineMu.Unlock()
	return isRunning
}

// RestoreEngine 服务启动时调用：开关持久化为开则自启（此前重启后不自启）
func RestoreEngine(ctx context.Context) {
	v, err := g.DB().Model("sys_config").Ctx(ctx).Where("key", enabledConfigKey).Value("value")
	if err != nil {
		g.Log().Warningf(ctx, "AI 引擎自启状态读取失败: %v", err)
		return
	}
	if v.String() == "1" {
		StartEngine(ctx)
	}
}

// persistEnabled 开关落库（失败只记日志，不影响启停动作）
func persistEnabled(ctx context.Context, on bool) {
	val := "0"
	if on {
		val = "1"
	}
	if err := dbutil.UpsertConfig(ctx, enabledConfigKey, val); err != nil {
		g.Log().Warningf(ctx, "persist engine enabled failed: %v", err)
	}
}

// scanAndExecute 扫描并执行一轮
func scanAndExecute(ctx context.Context) {
	engineMu.Lock()
	if !isRunning {
		engineMu.Unlock()
		return
	}
	engineMu.Unlock()

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

	// 3. 逐个处理（ctx 已绑定引擎取消：StopEngine 时中断进行中的调用）
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
	Id           int
	ProjectId    int
	Title        string
	Desc         string
	AssigneeId   int // AI 用户 ID
	AssigneeName string
	Attempts     int // 已失败次数（退避排程用）
}

// aiBackoff 失败退避间隔：attempts 超出表长后按末档，达 maxAttempts 进死信
var aiBackoff = []time.Duration{5 * time.Minute, 30 * time.Minute, 2 * time.Hour, 8 * time.Hour}

const (
	aiMaxAttempts = 5
	aiExecTimeout = 10 * time.Minute // 单任务执行上限：挂死的模型端点不得拖垮整个引擎
)

func getClaimableTasks(ctx context.Context, limit int) ([]ClaimableTask, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	var tasks []ClaimableTask
	err := g.DB().Model("tasks t").Ctx(ctx).
		LeftJoin("sys_users u", "t.assignee_id = u.id").
		Fields("t.id, t.project_id, t.title, t.description, t.assignee_id, COALESCE(u.real_name, u.username) as assignee_name, t.ai_attempts").
		Where("t.status", "open").
		Where("t.assignee_id > 0").
		Where("u.type", "ai").
		Where("u.status", 1).
		// 退避窗口未到的任务不认领（毒任务不再 2 分钟无限重试烧 API 费用）
		Where("(t.ai_next_attempt_at = '' OR t.ai_next_attempt_at <= ?)", now).
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

	// 单任务执行超时：挂死的模型端点会永久阻塞 processTask，gcron AddSingleton
	// 下后续所有轮次被跳过——一个任务拖垮整个引擎
	execCtx, cancel := context.WithTimeout(ctx, aiExecTimeout)
	defer cancel()

	// 调用 AI Agent 执行
	output, err := aiengine.ExecuteTask(execCtx, &aiengine.ExecuteConfig{
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
		g.Log().Errorf(ctx, "AI 引擎: 任务 %d 执行失败（第 %d 次）: %v", task.Id, task.Attempts+1, err)
		recordAiLog(ctx, task.AssigneeId, task.Id, "error", err.Error(), "failed")
		// 失败退避：按次数指数拉开重试间隔，超限进死信（closed），不再无限重试烧 API
		attempts := task.Attempts + 1
		if attempts >= aiMaxAttempts {
			g.DB().Model("tasks").Ctx(ctx).
				Where("id", task.Id).
				Where("status", "in_progress").
				Data(g.Map{
					"status":       "closed",
					"artifacts":    fmt.Sprintf("AI 执行连续失败 %d 次，已自动关闭：\n\n%s", attempts, err.Error()),
					"completed_at": time.Now().Format("2006-01-02 15:04:05"),
				}).Update()
			recordAiLog(ctx, task.AssigneeId, task.Id, "dead", "连续失败超限，任务转入死信", "failed")
			return
		}
		backoff := aiBackoff[len(aiBackoff)-1]
		if attempts-1 < len(aiBackoff) {
			backoff = aiBackoff[attempts-1]
		}
		g.DB().Model("tasks").Ctx(ctx).
			Where("id", task.Id).
			Where("status", "in_progress").
			Data(g.Map{
				"status":             "open",
				"ai_attempts":        attempts,
				"ai_next_attempt_at": time.Now().Add(backoff).Format("2006-01-02 15:04:05"),
			}).Update()
		return
	}

	// 提交结果
	g.DB().Model("tasks").Ctx(ctx).
		Where("id", task.Id).
		Data(g.Map{
			"status":             "review",
			"artifacts":          output,
			"completed_at":       time.Now().Format("2006-01-02 15:04:05"),
			"ai_attempts":        0,
			"ai_next_attempt_at": "",
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

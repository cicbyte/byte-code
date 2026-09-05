package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/agent"
	"github.com/cicbyte/byte-code/internal/logic/agent"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// AgentCtl 外部 Agent 接入协议（dev-docs/agent-protocol.md）。
// Register 挂公开组（无认证，纯身份零权限）；其余挂认证组
// （bc key 或人类 token——Join/Session 业务层限定 agent 身份）
var AgentCtl = agentController{}

type agentController struct{}

func (c *agentController) Register(ctx context.Context, req *api.RegisterReq) (*api.RegisterRes, error) {
	return agent.Register(ctx, req)
}

func (c *agentController) JoinCodeCreate(ctx context.Context, req *api.JoinCodeCreateReq) (*api.JoinCodeCreateRes, error) {
	return agent.JoinCodeCreate(ctx, req.ProjectId)
}

func (c *agentController) Join(ctx context.Context, req *api.JoinReq) (*api.JoinRes, error) {
	if !isAgentCtx(ctx) {
		return nil, gerror.New("仅 Agent（bc key）可加入项目")
	}
	return agent.Join(ctx, req.Code)
}

func (c *agentController) SessionCreate(ctx context.Context, req *api.SessionCreateReq) (*api.SessionCreateRes, error) {
	if !isAgentCtx(ctx) {
		return nil, gerror.New("仅 Agent（bc key）可建立工作会话")
	}
	return agent.SessionCreate(ctx, req)
}

func (c *agentController) AgentTasks(ctx context.Context, req *api.AgentTasksReq) (*api.AgentTasksRes, error) {
	if !isAgentCtx(ctx) {
		return nil, gerror.New("仅 Agent（bc key）可调用")
	}
	session := g.RequestFromCtx(ctx).Header.Get("X-Session")
	uid := ctxUserId(ctx)
	return agent.AgentTasks(ctx, uid, session, req.Status, req.Keyword)
}

// isAgentCtx 当前认证主体是 agent（sys_users.type=ai）
func isAgentCtx(ctx context.Context) bool {
	v, err := g.DB().Model("sys_users").Ctx(ctx).
		Where("id", ctxUserId(ctx)).Fields("type").Value()
	return err == nil && v != nil && v.String() == "ai"
}

func ctxUserId(ctx context.Context) int {
	if v, ok := ctx.Value("userId").(int); ok {
		return v
	}
	return 0
}

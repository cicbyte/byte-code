package project

// 项目服务实现按领域拆分：project(注册与类型)/project_crud(项目)/member(成员)/
// task+task_ops(任务)/comment(评论与AI日志)/sprint+burndown(迭代与燃尽)/
// requirement(需求)/milestone(里程碑)/activity(活动辅助)

import (
	service "github.com/cicbyte/byte-code/internal/service"
)

func init() {
	service.RegisterProject(New())
}

func New() *sProject {
	return &sProject{}
}

type sProject struct{}

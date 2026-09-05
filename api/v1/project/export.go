// Package api project 导出：项目全量 JSON（数据所有权出口）
package project

import (
	"github.com/gogf/gf/v2/frame/g"
)

type ProjectExportReq struct {
	g.Meta    `path:"/projects/{projectId}/export" method:"get" tags:"项目管理" summary:"导出项目全量数据（JSON；owner/超管）"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
}

// v1.1：补齐 members/attachments/entity_tags/activities/ai_execution_logs/
// 数据库模型四表与文档正文（v1.0 仅主实体，且 documentIndex 因排序 bug 恒为空）
type ProjectExportRes struct {
	g.Meta        `mime:"application/json"`
	Version       string                   `json:"version"`
	ExportedAt    string                   `json:"exportedAt"`
	Project       map[string]interface{}   `json:"project"`
	Members       []map[string]interface{} `json:"members"`
	Requirements  []map[string]interface{} `json:"requirements"`
	Milestones    []map[string]interface{} `json:"milestones"`
	Sprints       []map[string]interface{} `json:"sprints"`
	Tasks         []map[string]interface{} `json:"tasks"`
	Comments      []map[string]interface{} `json:"comments"`
	TestPlans     []map[string]interface{} `json:"testPlans"`
	TestCases     []map[string]interface{} `json:"testCases"`
	TestPlanCases []map[string]interface{} `json:"testPlanCases"`
	Memories      []map[string]interface{} `json:"memories"`
	DocumentIndex []map[string]interface{} `json:"documentIndex"`
	// 文档正文（vault 当前版本；.history 历史快照不含。key 为正斜杠相对
	// 路径，与 documentIndex.path 对齐）
	DocContents     map[string]string        `json:"docContents"`
	Attachments     []map[string]interface{} `json:"attachments"`
	EntityTags      []map[string]interface{} `json:"entityTags"`
	Activities      []map[string]interface{} `json:"activities"`
	AiExecutionLogs []map[string]interface{} `json:"aiExecutionLogs"`
	Databases       []map[string]interface{} `json:"databases"`
	DbTables        []map[string]interface{} `json:"dbTables"`
	DbColumns       []map[string]interface{} `json:"dbColumns"`
	SchemaVersions  []map[string]interface{} `json:"schemaVersions"`
}

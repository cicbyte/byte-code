// Package api project 导出：项目全量 JSON（数据所有权出口）
package project

import (
	"github.com/gogf/gf/v2/frame/g"
)

type ProjectExportReq struct {
	g.Meta    `path:"/projects/{projectId}/export" method:"get" tags:"项目管理" summary:"导出项目全量数据（JSON；owner/超管）"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
}

type ProjectExportRes struct {
	g.Meta        `mime:"application/json"`
	Version       string                   `json:"version"`
	ExportedAt    string                   `json:"exportedAt"`
	Project       map[string]interface{}   `json:"project"`
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
}

package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/review"
)

type IReview interface {
	// ListPending 跨项目待审聚合（管理员全量，成员限所在项目）
	ListPending(ctx context.Context, req *api.PendingListReq) (res *api.PendingListRes, err error)
	// BatchReview 批量审核：逐条复用单任务 ReviewTask 规则，收集逐条成败
	BatchReview(ctx context.Context, req *api.BatchReviewReq) (res *api.BatchReviewRes, err error)
}

var localReview IReview

func RegisterReview(i IReview) {
	localReview = i
}

func Review() IReview {
	if localReview == nil {
		panic("implementation not found for interface IReview")
	}
	return localReview
}

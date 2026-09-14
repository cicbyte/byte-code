package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/review"
	service "github.com/cicbyte/byte-code/internal/service"
)

var Review = reviewController{}

type reviewController struct {
	BaseController
}

func (c *reviewController) ListPending(ctx context.Context, req *api.PendingListReq) (res *api.PendingListRes, err error) {
	return service.Review().ListPending(ctx, req)
}

func (c *reviewController) BatchReview(ctx context.Context, req *api.BatchReviewReq) (res *api.BatchReviewRes, err error) {
	return service.Review().BatchReview(ctx, req)
}

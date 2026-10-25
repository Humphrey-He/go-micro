package logic

import (
	"context"

	"ecomm-micro/internal/generated/rpc/refund/internal/svc"
	"ecomm-micro/proto/refundpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type StatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StatusLogic {
	return &StatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *StatusLogic) Status(in *refundpb.StatusRequest) (*refundpb.Refund, error) {
	// todo: add your logic here and delete this line

	return &refundpb.Refund{}, nil
}

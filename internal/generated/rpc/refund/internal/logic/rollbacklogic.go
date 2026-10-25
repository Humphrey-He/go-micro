package logic

import (
	"context"

	"ecomm-micro/internal/generated/rpc/refund/internal/svc"
	"ecomm-micro/proto/refundpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RollbackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRollbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RollbackLogic {
	return &RollbackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RollbackLogic) Rollback(in *refundpb.RollbackRequest) (*refundpb.Refund, error) {
	// todo: add your logic here and delete this line

	return &refundpb.Refund{}, nil
}

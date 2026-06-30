package logic

import (
	"context"

	"go-micro/internal/generated/rpc/refund/internal/svc"
	"go-micro/proto/refundpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitiateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitiateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitiateLogic {
	return &InitiateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *InitiateLogic) Initiate(in *refundpb.InitiateRequest) (*refundpb.Refund, error) {
	// todo: add your logic here and delete this line

	return &refundpb.Refund{}, nil
}

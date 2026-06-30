package logic

import (
	"context"

	"go-micro/internal/generated/rpc/refund/internal/svc"
	"go-micro/proto/refundpb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListLogic) List(in *refundpb.ListRequest) (*refundpb.ListResponse, error) {
	// todo: add your logic here and delete this line

	return &refundpb.ListResponse{}, nil
}

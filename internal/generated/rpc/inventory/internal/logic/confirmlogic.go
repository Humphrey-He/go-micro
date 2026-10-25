package logic

import (
	"context"

	"ecomm-micro/internal/generated/rpc/inventory/internal/svc"
	"ecomm-micro/internal/generated/rpc/proto/inventorypb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfirmLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmLogic {
	return &ConfirmLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConfirmLogic) Confirm(in *inventorypb.ConfirmRequest) (*inventorypb.SimpleResponse, error) {
	// todo: add your logic here and delete this line

	return &inventorypb.SimpleResponse{}, nil
}

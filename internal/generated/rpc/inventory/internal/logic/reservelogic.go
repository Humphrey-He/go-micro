package logic

import (
	"context"

	"go-micro/internal/generated/rpc/inventory/internal/svc"
	"go-micro/internal/generated/rpc/proto/inventorypb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReserveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReserveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReserveLogic {
	return &ReserveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReserveLogic) Reserve(in *inventorypb.ReserveRequest) (*inventorypb.ReserveResponse, error) {
	// todo: add your logic here and delete this line

	return &inventorypb.ReserveResponse{}, nil
}

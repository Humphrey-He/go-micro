package logic

import (
	"context"

	"ecomm-micro/internal/generated/rpc/inventory/internal/svc"
	"ecomm-micro/internal/generated/rpc/proto/inventorypb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReleaseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReleaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseLogic {
	return &ReleaseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReleaseLogic) Release(in *inventorypb.ReleaseRequest) (*inventorypb.SimpleResponse, error) {
	// todo: add your logic here and delete this line

	return &inventorypb.SimpleResponse{}, nil
}

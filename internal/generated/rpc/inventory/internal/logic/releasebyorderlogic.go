package logic

import (
	"context"

	"go-micro/internal/generated/rpc/inventory/internal/svc"
	"go-micro/internal/generated/rpc/proto/inventorypb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReleaseByOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReleaseByOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseByOrderLogic {
	return &ReleaseByOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReleaseByOrderLogic) ReleaseByOrder(in *inventorypb.ReleaseByOrderRequest) (*inventorypb.SimpleResponse, error) {
	// todo: add your logic here and delete this line

	return &inventorypb.SimpleResponse{}, nil
}

package logic

import (
	"context"

	"ecomm-micro/internal/generated/rpc/inventory/internal/svc"
	"ecomm-micro/internal/generated/rpc/proto/inventorypb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListInventoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListInventoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListInventoryLogic {
	return &ListInventoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListInventoryLogic) ListInventory(in *inventorypb.ListInventoryRequest) (*inventorypb.ListInventoryResponse, error) {
	// todo: add your logic here and delete this line

	return &inventorypb.ListInventoryResponse{}, nil
}

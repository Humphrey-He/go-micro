package logic

import (
	"context"

	"go-micro/internal/generated/rpc/inventory/internal/svc"
	"go-micro/internal/generated/rpc/proto/inventorypb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReservationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetReservationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReservationLogic {
	return &GetReservationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetReservationLogic) GetReservation(in *inventorypb.GetReservationRequest) (*inventorypb.Reservation, error) {
	// todo: add your logic here and delete this line

	return &inventorypb.Reservation{}, nil
}

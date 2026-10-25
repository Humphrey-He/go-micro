package main

import (
	"flag"
	"fmt"

	"ecomm-micro/internal/generated/rpc/refund/internal/config"
	"ecomm-micro/internal/generated/rpc/refund/internal/server"
	"ecomm-micro/internal/generated/rpc/refund/internal/svc"
	"ecomm-micro/proto/refundpb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/refund.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		refundpb.RegisterRefundServiceServer(grpcServer, server.NewRefundServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

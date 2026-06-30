package main

import (
	"flag"
	"fmt"

	"go-micro/internal/generated/rpc/inventory/internal/config"
	"go-micro/internal/generated/rpc/inventory/internal/server"
	"go-micro/internal/generated/rpc/inventory/internal/svc"
	"go-micro/internal/generated/rpc/proto/inventorypb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/inventory.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		inventorypb.RegisterInventoryServiceServer(grpcServer, server.NewInventoryServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

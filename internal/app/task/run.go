package taskapp

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go-micro/internal/inventory"
	"go-micro/internal/order"
	"go-micro/internal/task"
	"go-micro/pkg/config"
	"go-micro/pkg/db"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
	"go-micro/pkg/mq"
	"go-micro/pkg/resilience"
	"go-micro/proto/taskpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Run() error {
	logger := logx.L()
	defer logx.Sync()

	dbx, err := db.NewMySQL()
	if err != nil {
		logger.Error("mysql connect failed", zap.Error(err))
		return err
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	svc := task.NewService(dbx)
	h := task.NewHandler(svc)
	h.Register(r)

	grpcAddr := config.GetEnv("TASK_GRPC_ADDR", ":9084")
	grpcLis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error("task-grpc listen failed", zap.Error(err))
		return err
	}
	grpcServer := grpc.NewServer()
	taskpb.RegisterTaskServiceServer(grpcServer, task.NewGRPCServer(svc))
	go func() {
		if err := grpcServer.Serve(grpcLis); err != nil {
			logger.Error("task-grpc serve failed", zap.Error(err))
		}
	}()

	mqURL := config.GetEnv("MQ_URL", "amqp://guest:guest@localhost:5672/")
	exchange := config.GetEnv("MQ_EXCHANGE", "order.events")
	queue := config.GetEnv("MQ_QUEUE", "order_reserved")
	routeKey := config.GetEnv("MQ_ROUTING_KEY", "order_reserved")
	dlx := config.GetEnv("MQ_DLX", "order.events.dlx")
	dlq := config.GetEnv("MQ_DLQ", "order_reserved.dlq")
	consumer, err := mq.NewRabbit(mqURL, exchange, queue, routeKey, dlx, dlq)
	if err != nil {
		logger.Error("mq connect failed", zap.Error(err))
		return err
	}
	defer consumer.Close()

	invTarget := config.GetEnv("INVENTORY_GRPC_TARGET", "localhost:9082")
	invClient, invConn, err := inventory.NewGRPCClient(invTarget)
	if err != nil {
		logger.Error("inventory grpc dial failed", zap.Error(err))
		return err
	}
	defer invConn.Close()

	orderTarget := config.GetEnv("ORDER_GRPC_TARGET", "localhost:9081")
	orderClient, orderConn, err := order.NewGRPCClient(orderTarget)
	if err != nil {
		logger.Error("order grpc dial failed", zap.Error(err))
		return err
	}
	defer orderConn.Close()

	cb := newBreakerFromEnv()
	go task.StartRabbitConsumer(consumer, svc, &inventoryReleaseAdapter{c: invClient, cb: cb}, &orderUpdateAdapter{c: orderClient, cb: cb})
	workerStop := make(chan struct{})
	go task.StartRetryWorker(svc, &orderUpdateAdapter{c: orderClient, cb: cb}, workerStop)
	go task.StartTimeoutWorker(svc, &orderReaderAdapter{c: orderClient, cb: cb}, &orderCancelAdapter{c: orderClient, cb: cb}, &inventoryReleaseByOrderAdapter{c: invClient, cb: cb}, svc, workerStop)

	addr := config.GetEnv("TASK_ADDR", ":8084")
	srv := &http.Server{Addr: addr, Handler: r}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("task-service start failed", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	close(workerStop)
	grpcServer.GracefulStop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

type inventoryReleaseAdapter struct {
	c  *inventory.GRPCClient
	cb *resilience.CircuitBreaker
}

func (a *inventoryReleaseAdapter) Release(ctx context.Context, reservedID string) error {
	return a.cb.Execute(func() error {
		return a.c.Release(ctx, reservedID)
	})
}

type orderUpdateAdapter struct {
	c  *order.GRPCClient
	cb *resilience.CircuitBreaker
}

func (a *orderUpdateAdapter) UpdateStatus(ctx context.Context, orderID, from, to string) error {
	return a.cb.Execute(func() error {
		return a.c.UpdateStatus(ctx, orderID, from, to)
	})
}

type orderReaderAdapter struct {
	c  *order.GRPCClient
	cb *resilience.CircuitBreaker
}

func (a *orderReaderAdapter) GetStatus(ctx context.Context, orderID string) (string, error) {
	var status string
	err := a.cb.Execute(func() error {
		ord, callErr := a.c.Get(ctx, orderID)
		if callErr != nil {
			return callErr
		}
		status = ord.Status
		return nil
	})
	if err != nil {
		return "", err
	}
	return status, nil
}

type orderCancelAdapter struct {
	c  *order.GRPCClient
	cb *resilience.CircuitBreaker
}

func (a *orderCancelAdapter) Cancel(ctx context.Context, orderID string) error {
	return a.cb.Execute(func() error {
		return a.c.Cancel(ctx, orderID)
	})
}

type inventoryReleaseByOrderAdapter struct {
	c  *inventory.GRPCClient
	cb *resilience.CircuitBreaker
}

func (a *inventoryReleaseByOrderAdapter) ReleaseByOrder(ctx context.Context, orderID string) error {
	return a.cb.Execute(func() error {
		return a.c.ReleaseByOrder(ctx, orderID)
	})
}

func newBreakerFromEnv() *resilience.CircuitBreaker {
	fail := getInt("CB_FAIL_THRESHOLD", 5)
	reset := getInt("CB_RESET_SECONDS", 10)
	half := getInt("CB_HALF_OPEN_SUCCESS", 1)
	return resilience.NewCircuitBreaker(fail, time.Duration(reset)*time.Second, half)
}

func getInt(key string, def int) int {
	v := config.GetEnv(key, "")
	if v == "" {
		return def
	}
	n := 0
	for i := 0; i < len(v); i++ {
		ch := v[i]
		if ch < '0' || ch > '9' {
			return def
		}
		n = n*10 + int(ch-'0')
	}
	if n == 0 {
		return def
	}
	return n
}

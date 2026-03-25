package gateway

import (
	"context"
	"time"

	"go-micro/internal/inventory"
	"go-micro/internal/order"
	"go-micro/internal/task"
	"go-micro/internal/user"
	"go-micro/pkg/config"
	"go-micro/pkg/httpx"
	"go-micro/pkg/resilience"
	"go-micro/proto/orderpb"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	order     *order.GRPCClient
	user      *user.GRPCClient
	inventory *inventory.GRPCClient
	task      *task.GRPCClient
	cbOrder   *resilience.CircuitBreaker
	cbUser    *resilience.CircuitBreaker
	cbInv     *resilience.CircuitBreaker
	cbTask    *resilience.CircuitBreaker
}

func NewService(orderClient *order.GRPCClient, userClient *user.GRPCClient, invClient *inventory.GRPCClient, taskClient *task.GRPCClient) *Service {
	cb := newBreakerFromEnv()
	return &Service{
		order:     orderClient,
		user:      userClient,
		inventory: invClient,
		task:      taskClient,
		cbOrder:   cb,
		cbUser:    cb,
		cbInv:     cb,
		cbTask:    cb,
	}
}

func (s *Service) CreateOrder(req CreateOrderRequest, requestID string) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	items := make([]order.Item, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, order.Item{SkuID: it.SkuID, Quantity: it.Quantity, Price: it.Price})
	}
	var resp *orderpb.CreateOrderResponse
	err := s.cbOrder.Execute(func() error {
		var callErr error
		resp, callErr = s.order.Create(ctx, order.CreateOrderRequest{
			RequestID: req.RequestID,
			UserID:    req.UserID,
			Items:     items,
			Remark:    req.Remark,
		})
		return callErr
	})
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: resp}, nil
}

func (s *Service) GetOrder(orderID, requestID string) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var resp *orderpb.Order
	err := s.cbOrder.Execute(func() error {
		var callErr error
		resp, callErr = s.order.Get(ctx, orderID)
		return callErr
	})
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: resp}, nil
}

func (s *Service) GetOrderView(bizNo string) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var ord *orderpb.Order
	err := s.cbOrder.Execute(func() error {
		var callErr error
		ord, callErr = s.order.GetByBizNo(ctx, bizNo)
		return callErr
	})
	if err != nil {
		return httpx.Response{}, err
	}

	resvStatus := "UNKNOWN"
	if s.inventory != nil {
		_ = s.cbInv.Execute(func() error {
			resv, callErr := s.inventory.GetReservation(ctx, ord.OrderId)
			if callErr == nil && resv != nil {
				resvStatus = resv.Status
			}
			return nil
		})
	}

	taskStatus := "NOT_FOUND"
	taskType := ""
	if s.task != nil {
		_ = s.cbTask.Execute(func() error {
			t, callErr := s.task.GetByOrder(ctx, ord.OrderId)
			if callErr == nil && t != nil {
				taskStatus = t.Status
				taskType = t.Type
			}
			return nil
		})
	}

	viewStatus, cancelReason := computeViewStatus(ord.Status, taskStatus, taskType, resvStatus)
	view := OrderViewResponse{
		OrderNo:           ord.BizNo,
		ViewStatus:        viewStatus,
		OrderStatus:       ord.Status,
		TaskStatus:        taskStatus,
		ReservationStatus: resvStatus,
		CancelReason:      cancelReason,
	}
	return httpx.Response{Code: 0, Message: "OK", Data: view}, nil
}

func computeViewStatus(orderStatus, taskStatus, taskType, resvStatus string) (string, string) {
	if orderStatus == "CANCELED" {
		if taskType == "TIMEOUT_CANCEL" {
			return "TIMEOUT", "timeout"
		}
		return "CANCELED", ""
	}
	if taskStatus == "DEAD" {
		return "DEAD", ""
	}
	if taskStatus == "FAILED" {
		return "FAILED", ""
	}
	if orderStatus == "SUCCESS" {
		return "SUCCESS", ""
	}
	if taskStatus == "RUNNING" {
		return "PROCESSING", ""
	}
	if orderStatus == "RESERVED" && (taskStatus == "PENDING" || taskStatus == "NOT_FOUND") {
		return "PENDING", ""
	}
	if orderStatus != "" {
		return orderStatus, ""
	}
	return "UNKNOWN", ""
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string     `json:"token"`
	User  *user.User `json:"user"`
}

func (s *Service) Login(username, password string) (*LoginResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var u *user.User
	err := s.cbUser.Execute(func() error {
		var callErr error
		u, callErr = s.user.Authenticate(ctx, username, password)
		return callErr
	})
	if err != nil {
		return nil, err
	}
	secret := []byte(config.GetEnv("JWT_SECRET", "dev-secret"))
	claims := jwt.MapClaims{
		"user_id":  u.UserId,
		"username": u.Username,
		"role":     u.Role,
		"exp":      time.Now().Add(2 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return nil, err
	}
	return &LoginResponse{
		Token: signed,
		User:  &user.User{UserID: u.UserId, Username: u.Username, Role: u.Role, Status: int(u.Status)},
	}, nil
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

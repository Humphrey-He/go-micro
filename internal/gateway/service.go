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

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	order     *order.GRPCClient
	user      *user.GRPCClient
	inventory *inventory.GRPCClient
	task      *task.GRPCClient
}

func NewService(orderClient *order.GRPCClient, userClient *user.GRPCClient, invClient *inventory.GRPCClient, taskClient *task.GRPCClient) *Service {
	return &Service{order: orderClient, user: userClient, inventory: invClient, task: taskClient}
}

func (s *Service) CreateOrder(req CreateOrderRequest, requestID string) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	items := make([]order.Item, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, order.Item{SkuID: it.SkuID, Quantity: it.Quantity, Price: it.Price})
	}
	resp, err := s.order.Create(ctx, order.CreateOrderRequest{
		RequestID: req.RequestID,
		UserID:    req.UserID,
		Items:     items,
		Remark:    req.Remark,
	})
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: resp}, nil
}

func (s *Service) GetOrder(orderID, requestID string) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := s.order.Get(ctx, orderID)
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: resp}, nil
}

func (s *Service) GetOrderView(bizNo string) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ord, err := s.order.GetByBizNo(ctx, bizNo)
	if err != nil {
		return httpx.Response{}, err
	}

	invStatus := "UNKNOWN"
	if s.inventory != nil {
		if resv, err := s.inventory.GetReservation(ctx, ord.OrderId); err == nil {
			invStatus = resv.Status
		}
	}

	taskStatus := "UNKNOWN"
	if s.task != nil {
		if t, err := s.task.GetByOrder(ctx, ord.OrderId); err == nil {
			taskStatus = t.Status
		}
	}

	view := OrderViewResponse{
		OrderNo:         ord.BizNo,
		Status:          ord.Status,
		InventoryStatus: invStatus,
		TaskStatus:      taskStatus,
	}
	return httpx.Response{Code: 0, Message: "OK", Data: view}, nil
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
	u, err := s.user.Authenticate(ctx, username, password)
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

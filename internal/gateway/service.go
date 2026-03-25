package gateway

import (
	"context"
	"errors"
	"time"

	"go-micro/internal/activity"
	"go-micro/internal/inventory"
	"go-micro/internal/order"
	"go-micro/internal/payment"
	"go-micro/internal/price"
	"go-micro/internal/refund"
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
	refund    *refund.GRPCClient
	activity  *activity.GRPCClient
	price     *price.GRPCClient
	payment   *payment.Service
	cbOrder   *resilience.CircuitBreaker
	cbUser    *resilience.CircuitBreaker
	cbInv     *resilience.CircuitBreaker
	cbTask    *resilience.CircuitBreaker
	cbRefund  *resilience.CircuitBreaker
	cbAct     *resilience.CircuitBreaker
	cbPrice   *resilience.CircuitBreaker
}

func NewService(orderClient *order.GRPCClient, userClient *user.GRPCClient, invClient *inventory.GRPCClient, taskClient *task.GRPCClient, refundClient *refund.GRPCClient, activityClient *activity.GRPCClient, priceClient *price.GRPCClient) *Service {
	cb := newBreakerFromEnv()
	return &Service{
		order:     orderClient,
		user:      userClient,
		inventory: invClient,
		task:      taskClient,
		refund:    refundClient,
		activity:  activityClient,
		price:     priceClient,
		payment:   nil,
		cbOrder:   cb,
		cbUser:    cb,
		cbInv:     cb,
		cbTask:    cb,
		cbRefund:  cb,
		cbAct:     cb,
		cbPrice:   cb,
	}
}

func (s *Service) SetPayment(p *payment.Service) {
	s.payment = p
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

func (s *Service) CreatePayment(req payment.CreatePaymentRequest) (httpx.Response, error) {
	if s.payment == nil {
		return httpx.Response{}, errPaymentNotInit
	}
	p, err := s.payment.Create(req)
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: p}, nil
}

func (s *Service) GetPayment(paymentID string) (httpx.Response, error) {
	if s.payment == nil {
		return httpx.Response{}, errPaymentNotInit
	}
	p, err := s.payment.Get(paymentID)
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: p}, nil
}

func (s *Service) MarkPaymentSuccess(paymentID string) (httpx.Response, error) {
	if s.payment == nil {
		return httpx.Response{}, errPaymentNotInit
	}
	if err := s.payment.MarkSuccess(paymentID); err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: map[string]bool{"success": true}}, nil
}

func (s *Service) MarkPaymentFailed(paymentID string) (httpx.Response, error) {
	if s.payment == nil {
		return httpx.Response{}, errPaymentNotInit
	}
	if err := s.payment.MarkFailed(paymentID); err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: map[string]bool{"success": true}}, nil
}

func (s *Service) MarkPaymentTimeout(paymentID string) (httpx.Response, error) {
	if s.payment == nil {
		return httpx.Response{}, errPaymentNotInit
	}
	if err := s.payment.MarkTimeout(paymentID); err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: map[string]bool{"success": true}}, nil
}

func (s *Service) RefundInitiate(req refund.InitiateRequest) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var resp *refund.Refund
	err := s.cbRefund.Execute(func() error {
		var callErr error
		resp, callErr = s.refund.Initiate(ctx, req)
		return callErr
	})
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: resp}, nil
}

func (s *Service) RefundStatus(req refund.StatusRequest) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var resp *refund.Refund
	err := s.cbRefund.Execute(func() error {
		var callErr error
		resp, callErr = s.refund.Status(ctx, req)
		return callErr
	})
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: resp}, nil
}

func (s *Service) RefundRollback(req refund.RollbackRequest) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var resp *refund.Refund
	err := s.cbRefund.Execute(func() error {
		var callErr error
		resp, callErr = s.refund.Rollback(ctx, req)
		return callErr
	})
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: resp}, nil
}

func (s *Service) IssueCoupon(req activity.CouponRequest) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var resp *activity.Coupon
	err := s.cbAct.Execute(func() error {
		var callErr error
		resp, callErr = s.activity.IssueCoupon(ctx, req)
		return callErr
	})
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: resp}, nil
}

func (s *Service) Seckill(req activity.SeckillRequest) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var resp *activity.SeckillOrder
	err := s.cbAct.Execute(func() error {
		var callErr error
		resp, callErr = s.activity.Seckill(ctx, req)
		return callErr
	})
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: resp}, nil
}

func (s *Service) GetActivityStatus(couponID, skuID string) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if couponID != "" {
		var resp *activity.Coupon
		err := s.cbAct.Execute(func() error {
			var callErr error
			resp, callErr = s.activity.GetCoupon(ctx, couponID)
			return callErr
		})
		if err != nil {
			return httpx.Response{}, err
		}
		return httpx.Response{Code: 0, Message: "OK", Data: resp}, nil
	}
	if skuID != "" {
		var resp *activity.Seckill
		err := s.cbAct.Execute(func() error {
			var callErr error
			resp, callErr = s.activity.GetSeckill(ctx, skuID)
			return callErr
		})
		if err != nil {
			return httpx.Response{}, err
		}
		return httpx.Response{Code: 0, Message: "OK", Data: resp}, nil
	}
	return httpx.Response{}, errors.New("coupon_id or sku_id required")
}

func (s *Service) CalculatePrice(req price.CalculateRequest) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var resp *price.CalculateResponse
	err := s.cbPrice.Execute(func() error {
		var callErr error
		resp, callErr = s.price.Calculate(ctx, req)
		return callErr
	})
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: resp}, nil
}

func (s *Service) GetPriceHistory(skuID string, limit int) (httpx.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var resp []price.History
	err := s.cbPrice.Execute(func() error {
		var callErr error
		resp, callErr = s.price.History(ctx, skuID, limit)
		return callErr
	})
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Response{Code: 0, Message: "OK", Data: resp}, nil
}

var errPaymentNotInit = errors.New("payment service not initialized")

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
		upb, callErr := s.user.Authenticate(ctx, username, password)
		if callErr == nil && upb != nil {
			u = &user.User{
				UserID:   upb.UserId,
				Username: upb.Username,
				Role:     upb.Role,
				Status:   int(upb.Status),
			}
		}
		return callErr
	})
	if err != nil {
		return nil, err
	}
	secret := []byte(config.GetEnv("JWT_SECRET", "dev-secret"))
	claims := jwt.MapClaims{
		"user_id":  u.UserID,
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
		User:  &user.User{UserID: u.UserID, Username: u.Username, Role: u.Role, Status: int(u.Status)},
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

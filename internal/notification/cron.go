package notification

import (
    "context"
    "log"
    "strconv"

    "github.com/robfig/cron/v3"
)

type CronService struct {
    notificationSvc *Service
    cron            *cron.Cron
}

func NewCronService(notificationSvc *Service) *CronService {
    return &CronService{
        notificationSvc: notificationSvc,
        cron:            cron.New(),
    }
}

func (s *CronService) Start() error {
    // 检查退款告警 - 每5分钟
    _, err := s.cron.AddFunc("*/5 * * * *", func() {
        s.checkRefundAlert()
    })
    if err != nil {
        return err
    }

    // 检查库存告警 - 每5分钟
    _, err = s.cron.AddFunc("*/5 * * * *", func() {
        s.checkLowStockAlert()
    })
    if err != nil {
        return err
    }

    // 每日报告 - 每天早上9点
    _, err = s.cron.AddFunc("0 9 * * *", func() {
        if err := s.notificationSvc.SendDailyReport(); err != nil {
            log.Printf("Failed to send daily report: %v", err)
        }
    })
    if err != nil {
        return err
    }

    // 每周报告 - 每周一早上9点
    _, err = s.cron.AddFunc("0 9 * * 1", func() {
        if err := s.notificationSvc.SendWeeklyReport(); err != nil {
            log.Printf("Failed to send weekly report: %v", err)
        }
    })
    if err != nil {
        return err
    }

    s.cron.Start()
    log.Println("Notification cron service started")
    return nil
}

func (s *CronService) Stop() {
    s.cron.Stop()
}

func (s *CronService) checkRefundAlert() {
    // TODO: 从数据库或服务获取当前待处理退款数量
    // 这里需要接入实际的退款服务
    pendingCount := 0 // 暂时设为0，后续接入实际服务

    if pendingCount > 10 {
        err := s.notificationSvc.CreateNotification(
            context.Background(),
            "admin",
            TypeRefundPending,
            "退款告警",
            "待处理退款数量超过阈值，当前："+strconv.Itoa(pendingCount)+"件",
        )
        if err != nil {
            log.Printf("Failed to create refund alert: %v", err)
        }
    }
}

func (s *CronService) checkLowStockAlert() {
    // TODO: 从库存服务获取低库存SKU数量
    lowStockCount := 0

    if lowStockCount > 5 {
        err := s.notificationSvc.CreateNotification(
            context.Background(),
            "admin",
            TypeLowStock,
            "库存告警",
            "低库存SKU数量超过阈值，当前："+strconv.Itoa(lowStockCount)+"个",
        )
        if err != nil {
            log.Printf("Failed to create low stock alert: %v", err)
        }
    }
}

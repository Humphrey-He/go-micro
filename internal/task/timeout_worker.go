package task

import (
	"context"
	"time"
)

func StartTimeoutWorker(svc *Service, ord OrderReader, canceler OrderCanceler, stop <-chan struct{}) {
	if svc == nil {
		return
	}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			tasks, err := svc.ListTimeoutTasks(20)
			if err != nil {
				continue
			}
			for i := range tasks {
				t := tasks[i]
				ok, err := svc.MarkRunningIfStatus(t.TaskID, statusPending)
				if err != nil || !ok {
					continue
				}
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				if ord == nil {
					_ = svc.DelayTask(t.TaskID, t.RetryCount+1, nextRetryAt(t.RetryCount+1))
					cancel()
					continue
				}
				status, err := ord.GetStatus(ctx, t.OrderID)
				if err != nil {
					_ = svc.DelayTask(t.TaskID, t.RetryCount+1, nextRetryAt(t.RetryCount+1))
					cancel()
					continue
				}
				if status == orderStatusSuccess || status == orderStatusCanceled {
					_ = svc.MarkSuccess(t.TaskID)
					cancel()
					continue
				}
				if canceler != nil {
					if err := canceler.Cancel(ctx, t.OrderID); err == nil {
						_ = svc.MarkSuccess(t.TaskID)
						cancel()
						continue
					}
				}
				retryCount := t.RetryCount + 1
				if retryCount >= maxRetryCount {
					_ = svc.MarkDead(t.TaskID, retryCount)
					cancel()
					continue
				}
				_ = svc.DelayTask(t.TaskID, retryCount, nextRetryAt(retryCount))
				cancel()
			}
		}
	}
}

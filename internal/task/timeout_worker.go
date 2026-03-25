package task

import (
	"context"
	"time"
)

type InventoryReleaserByOrder interface {
	ReleaseByOrder(ctx context.Context, orderID string) error
}

type SagaUpdater interface {
	MarkSagaCompensated(sagaID, reason string) error
}

type TimeoutTaskStore interface {
	ListTimeoutTasks(limit int) ([]Task, error)
	MarkRunningIfStatus(taskID, status string) (bool, error)
	MarkSuccess(taskID string) error
	DelayTask(taskID string, retryCount int, nextRetryAt time.Time) error
	MarkDead(taskID string, retryCount int) error
}

func StartTimeoutWorker(svc TimeoutTaskStore, ord OrderReader, canceler OrderCanceler, inv InventoryReleaserByOrder, saga SagaUpdater, stop <-chan struct{}) {
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
			processTimeoutTasks(svc, ord, canceler, inv, saga)
		}
	}
}

func processTimeoutTasks(svc TimeoutTaskStore, ord OrderReader, canceler OrderCanceler, inv InventoryReleaserByOrder, saga SagaUpdater) {
	tasks, err := svc.ListTimeoutTasks(20)
	if err != nil {
		return
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
				if inv != nil {
					_ = inv.ReleaseByOrder(ctx, t.OrderID)
				}
				if saga != nil {
					_ = saga.MarkSagaCompensated(t.OrderID, "timeout")
				}
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

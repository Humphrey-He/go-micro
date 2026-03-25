package task

import (
	"context"
	"time"
)

const (
	orderStatusReserved   = "RESERVED"
	orderStatusProcessing = "PROCESSING"
	orderStatusSuccess    = "SUCCESS"
	orderStatusFailed     = "FAILED"
)

var retryBackoff = []time.Duration{
	1 * time.Minute,
	5 * time.Minute,
	15 * time.Minute,
}

func processFulfillment(ctx context.Context, svc *Service, ord OrderUpdater, t *Task) error {
	if err := svc.MarkRunning(t.TaskID); err != nil {
		return err
	}
	if ord != nil {
		if err := ord.UpdateStatus(ctx, t.OrderID, orderStatusReserved, orderStatusProcessing); err != nil {
			return err
		}
		if err := ord.UpdateStatus(ctx, t.OrderID, orderStatusProcessing, orderStatusSuccess); err != nil {
			return err
		}
	}
	return svc.MarkSuccess(t.TaskID)
}

func handleTaskFailure(ctx context.Context, svc *Service, ord OrderUpdater, t *Task) {
	retryCount := t.RetryCount + 1
	if retryCount >= maxRetryCount {
		_ = svc.MarkDead(t.TaskID, retryCount)
		if ord != nil {
			_ = ord.UpdateStatus(ctx, t.OrderID, orderStatusProcessing, orderStatusFailed)
		}
		return
	}
	next := nextRetryAt(retryCount)
	_ = svc.MarkFailed(t.TaskID, retryCount, next)
}

func nextRetryAt(retryCount int) time.Time {
	if retryCount <= 0 {
		return time.Now().Add(retryBackoff[0])
	}
	if retryCount-1 >= len(retryBackoff) {
		return time.Now().Add(retryBackoff[len(retryBackoff)-1])
	}
	return time.Now().Add(retryBackoff[retryCount-1])
}

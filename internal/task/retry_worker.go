package task

import (
	"context"
	"time"
)

func StartRetryWorker(svc *Service, ord OrderUpdater, stop <-chan struct{}) {
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
			tasks, err := svc.ListRetryTasks(20)
			if err != nil {
				continue
			}
			for i := range tasks {
				t := tasks[i]
				ok, err := svc.MarkRunningIfFailed(t.TaskID)
				if err != nil || !ok {
					continue
				}
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				if err := processFulfillment(ctx, svc, ord, &t); err != nil {
					handleTaskFailure(ctx, svc, ord, &t)
				}
				cancel()
			}
		}
	}
}

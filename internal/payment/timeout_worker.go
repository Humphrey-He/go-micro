package payment

import (
	"time"
)

func StartTimeoutWorker(svc *Service, stop <-chan struct{}) {
	if svc == nil {
		return
	}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	timeout := getTimeout()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			before := time.Now().Add(-timeout)
			rows, err := svc.ListTimeoutPayments(before, 20)
			if err != nil {
				continue
			}
			for i := range rows {
				_ = svc.MarkTimeout(rows[i].PaymentID)
			}
		}
	}
}

func getTimeout() time.Duration {
	return 15 * time.Minute
}

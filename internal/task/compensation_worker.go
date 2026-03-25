package task

import (
	"context"
	"encoding/json"
	"time"
)

type CompensationOrderUpdater interface {
	UpdateStatus(ctx context.Context, orderID, from, to string) error
	Cancel(ctx context.Context, orderID string) error
}

type CompensationInventory interface {
	Release(ctx context.Context, reservedID string) error
	ReleaseByOrder(ctx context.Context, orderID string) error
}

type StepPayload struct {
	OrderID    string `json:"order_id"`
	ReservedID string `json:"reserved_id"`
	From       string `json:"from"`
	To         string `json:"to"`
}

func StartCompensationWorker(svc *Service, ord CompensationOrderUpdater, inv CompensationInventory, stop <-chan struct{}) {
	if svc == nil {
		return
	}
	// Poll saga_steps and execute compensation asynchronously.
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			processCompensations(svc, ord, inv)
		}
	}
}

func processCompensations(svc *Service, ord CompensationOrderUpdater, inv CompensationInventory) {
	steps, err := svc.ListPendingSagaSteps(20)
	if err != nil {
		return
	}
	for i := range steps {
		step := steps[i]
		ok, err := svc.MarkSagaStepRunning(step.SagaID, step.Step)
		if err != nil || !ok {
			continue
		}

		var payload StepPayload
		_ = json.Unmarshal([]byte(step.Payload), &payload)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = executeCompensation(ctx, ord, inv, step.Step, payload)
		cancel()

		if err != nil {
			_ = svc.MarkSagaStepFailed(step.SagaID, step.Step, err.Error())
			continue
		}

		_ = svc.MarkSagaStepCompleted(step.SagaID, step.Step)
		if step.NextStep != "" {
			_ = svc.CreateSagaStep(step.SagaID, step.NextStep, "", step.Reason, step.Payload)
			continue
		}
		_ = svc.MarkSagaCompensated(step.SagaID, step.Reason)
	}
}

func executeCompensation(ctx context.Context, ord CompensationOrderUpdater, inv CompensationInventory, step string, payload StepPayload) error {
	switch step {
	case stepOrderFail:
		if ord != nil {
			return ord.UpdateStatus(ctx, payload.OrderID, payload.From, payload.To)
		}
	case stepOrderCancel:
		if ord != nil {
			return ord.Cancel(ctx, payload.OrderID)
		}
	case stepInvRelease:
		if inv != nil {
			return inv.Release(ctx, payload.ReservedID)
		}
	case stepInvReleaseOrder:
		if inv != nil {
			return inv.ReleaseByOrder(ctx, payload.OrderID)
		}
	}
	return nil
}

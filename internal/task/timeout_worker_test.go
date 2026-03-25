package task

import (
	"context"
	"testing"
	"time"
)

type fakeTimeoutStore struct {
	tasks     []Task
	processed map[string]bool
	marked    map[string]string
}

func newFakeTimeoutStore(tasks []Task) *fakeTimeoutStore {
	return &fakeTimeoutStore{tasks: tasks, processed: map[string]bool{}, marked: map[string]string{}}
}

func (f *fakeTimeoutStore) ListTimeoutTasks(limit int) ([]Task, error) {
	return f.tasks, nil
}

func (f *fakeTimeoutStore) MarkRunningIfStatus(taskID, status string) (bool, error) {
	if f.processed[taskID] {
		return false, nil
	}
	f.processed[taskID] = true
	return true, nil
}

func (f *fakeTimeoutStore) MarkSuccess(taskID string) error {
	f.marked[taskID] = "SUCCESS"
	return nil
}

func (f *fakeTimeoutStore) DelayTask(taskID string, retryCount int, nextRetryAt time.Time) error {
	f.marked[taskID] = "DELAY"
	return nil
}

func (f *fakeTimeoutStore) MarkDead(taskID string, retryCount int) error {
	f.marked[taskID] = "DEAD"
	return nil
}

type fakeOrderReader struct {
	status string
	err    error
}

func (f *fakeOrderReader) GetStatus(ctx context.Context, orderID string) (string, error) {
	return f.status, f.err
}

type fakeOrderCanceler struct {
	calls int
	err   error
}

func (f *fakeOrderCanceler) Cancel(ctx context.Context, orderID string) error {
	f.calls++
	return f.err
}

type fakeInventoryReleaser struct {
	calls int
}

func (f *fakeInventoryReleaser) ReleaseByOrder(ctx context.Context, orderID string) error {
	f.calls++
	return nil
}

func TestTimeoutCancelChain(t *testing.T) {
	store := newFakeTimeoutStore([]Task{{TaskID: "T-1", OrderID: "O-1", RetryCount: 0, Status: statusPending}})
	reader := &fakeOrderReader{status: orderStatusReserved}
	canceler := &fakeOrderCanceler{}
	releaser := &fakeInventoryReleaser{}

	processTimeoutTasks(store, reader, canceler, releaser)
	if canceler.calls != 1 {
		t.Fatalf("expected cancel called once, got %d", canceler.calls)
	}
	if releaser.calls != 1 {
		t.Fatalf("expected release called once, got %d", releaser.calls)
	}
	if store.marked["T-1"] != "SUCCESS" {
		t.Fatalf("expected task marked success, got %s", store.marked["T-1"])
	}

	// Run again to ensure no duplicate release
	processTimeoutTasks(store, reader, canceler, releaser)
	if canceler.calls != 1 {
		t.Fatalf("expected cancel called once after repeat, got %d", canceler.calls)
	}
	if releaser.calls != 1 {
		t.Fatalf("expected release called once after repeat, got %d", releaser.calls)
	}
}

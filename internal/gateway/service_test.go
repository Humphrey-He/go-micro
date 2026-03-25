package gateway

import "testing"

func TestComputeViewStatus(t *testing.T) {
	tests := []struct {
		name         string
		orderStatus  string
		taskStatus   string
		taskType     string
		resvStatus   string
		wantView     string
		wantReason   string
	}{
		{name: "canceled", orderStatus: "CANCELED", taskStatus: "FAILED", taskType: "FULFILL", wantView: "CANCELED"},
		{name: "timeout", orderStatus: "CANCELED", taskStatus: "DEAD", taskType: "TIMEOUT_CANCEL", wantView: "TIMEOUT", wantReason: "timeout"},
		{name: "dead", orderStatus: "RESERVED", taskStatus: "DEAD", taskType: "FULFILL", wantView: "DEAD"},
		{name: "failed", orderStatus: "RESERVED", taskStatus: "FAILED", taskType: "FULFILL", wantView: "FAILED"},
		{name: "success", orderStatus: "SUCCESS", taskStatus: "SUCCESS", taskType: "FULFILL", wantView: "SUCCESS"},
		{name: "processing", orderStatus: "RESERVED", taskStatus: "RUNNING", taskType: "FULFILL", wantView: "PROCESSING"},
		{name: "pending", orderStatus: "RESERVED", taskStatus: "PENDING", taskType: "FULFILL", wantView: "PENDING"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotView, gotReason := computeViewStatus(tt.orderStatus, tt.taskStatus, tt.taskType, tt.resvStatus)
			if gotView != tt.wantView {
				t.Fatalf("view_status expected %s, got %s", tt.wantView, gotView)
			}
			if gotReason != tt.wantReason {
				t.Fatalf("cancel_reason expected %s, got %s", tt.wantReason, gotReason)
			}
		})
	}
}

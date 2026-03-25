package main

import "testing"

func TestRunSuccess(t *testing.T) {
	if code := run(func() error { return nil }); code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
}

func TestRunFail(t *testing.T) {
	if code := run(func() error { return errTest }); code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}
}

var errTest = errString("err")

type errString string

func (e errString) Error() string { return string(e) }

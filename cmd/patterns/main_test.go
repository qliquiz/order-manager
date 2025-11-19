package main

import (
	"errors"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	t.Run("Success eventually", func(t *testing.T) {
		attempts := 0
		op := func() error {
			attempts++
			if attempts < 3 {
				return errors.New("fail")
			}
			return nil
		}

		start := time.Now()
		err := Retry(op, 5, 10)

		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
		if time.Since(start) < 30*time.Millisecond {
			t.Log("Warning: execution was faster than expected delay, check backoff logic")
		}
	})

	t.Run("Fail after max retries", func(t *testing.T) {
		calls := 0
		op := func() error {
			calls++
			return errors.New("always fail")
		}

		err := Retry(op, 2, 1)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if calls != 3 {
			t.Errorf("expected 3 calls, got %d", calls)
		}
	})
}

func TestTimeout(t *testing.T) {
	t.Run("Operation succeeds in time", func(t *testing.T) {
		op := func() error {
			time.Sleep(10 * time.Millisecond)
			return nil
		}
		err := Timeout(op, 100)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("Operation times out", func(t *testing.T) {
		op := func() error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}
		err := Timeout(op, 10)
		if !errors.Is(err, ErrTimeout) {
			t.Errorf("expected timeout error, got %v", err)
		}
	})
}

func TestDLQ(t *testing.T) {
	messages := []string{"good1", "bad1", "good2", "bad2"}
	dlq := NewDeadLetterQueue()

	processor := func(msg string) error {
		if msg == "bad1" || msg == "bad2" {
			return errors.New("fail")
		}
		return nil
	}

	ProcessWithDLQ(messages, processor, dlq)

	deadMsgs := dlq.GetMessages()
	if len(deadMsgs) != 2 {
		t.Errorf("expected 2 messages in DLQ, got %d", len(deadMsgs))
	}

	expected := map[string]bool{"bad1": true, "bad2": true}
	for _, msg := range deadMsgs {
		if !expected[msg] {
			t.Errorf("unexpected message in DLQ: %s", msg)
		}
	}
}

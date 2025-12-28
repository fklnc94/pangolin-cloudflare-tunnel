package retry

import (
	"errors"
	"testing"
)

func TestDo_Success(t *testing.T) {
	callCount := 0
	operation := func() error {
		callCount++
		return nil
	}

	err := Do(3, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected operation to be called once, was called %d times", callCount)
	}
}

func TestDo_SuccessAfterRetries(t *testing.T) {
	callCount := 0
	operation := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	err := Do(3, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected operation to be called 3 times, was called %d times", callCount)
	}
}

func TestDo_FailureAfterMaxRetries(t *testing.T) {
	callCount := 0
	expectedErr := errors.New("persistent error")
	operation := func() error {
		callCount++
		return expectedErr
	}

	err := Do(3, operation)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if callCount != 3 {
		t.Errorf("Expected operation to be called 3 times, was called %d times", callCount)
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error to wrap original error")
	}
}

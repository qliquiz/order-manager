package main

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// Retry повторяет операцию maxRetries раз с экспоненциальной задержкой.
func Retry(operation func() error, maxRetries int, baseDelay int) error {
	var err error
	for i := 0; i <= maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		if i == maxRetries {
			break
		}

		delay := float64(baseDelay) * math.Pow(2, float64(i))
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}

var ErrTimeout = errors.New("operation timed out")

// Timeout выполняет операцию и возвращает ошибку, если она не уложилась в отведенное время.
func Timeout(operation func() error, timeoutMs int) error {
	done := make(chan error, 1)

	go func() {
		done <- operation()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
		return ErrTimeout
	}
}

// DeadLetterQueue - структура для хранения необработанных сообщений.
type DeadLetterQueue struct {
	mu       sync.Mutex
	messages []string
}

// NewDeadLetterQueue создает новую очередь.
func NewDeadLetterQueue() *DeadLetterQueue {
	return &DeadLetterQueue{
		messages: make([]string, 0),
	}
}

// Enqueue добавляет сообщение в очередь.
func (dlq *DeadLetterQueue) Enqueue(msg string) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	dlq.messages = append(dlq.messages, msg)
}

// GetMessages возвращает все сообщения из очереди.
func (dlq *DeadLetterQueue) GetMessages() []string {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	result := make([]string, len(dlq.messages))
	copy(result, dlq.messages)
	return result
}

// ProcessWithDLQ обрабатывает сообщения и отправляет неудачные в DLQ.
func ProcessWithDLQ(messages []string, processor func(string) error, dlq *DeadLetterQueue) {
	for _, msg := range messages {
		if err := processor(msg); err != nil {
			fmt.Printf("Failed to process '%s': %v. Sending to DLQ.\n", msg, err)
			dlq.Enqueue(msg)
		}
	}
}

func main() {
	// 1. Retry
	fmt.Println("--- Retry Demo ---")
	failCounter := 0
	operationRetry := func() error {
		if failCounter < 2 {
			failCounter++
			fmt.Println("Retry op: Operation failed...")
			return errors.New("temporary error")
		}
		fmt.Println("Retry op: Operation success!")
		return nil
	}
	_ = Retry(operationRetry, 3, 100)

	// 2. Timeout
	fmt.Println("\n--- Timeout Demo ---")
	operationTimeout := func() error {
		time.Sleep(500 * time.Millisecond)
		return nil
	}
	err := Timeout(operationTimeout, 200) // Таймаут 200мс, операция 500мс
	if err != nil {
		fmt.Println("Timeout result:", err)
	}

	// 3. DLQ
	fmt.Println("\n--- DLQ Demo ---")
	messages := []string{"msg1", "msg2", "msg3"}
	dlq := NewDeadLetterQueue()

	ProcessWithDLQ(messages, func(msg string) error {
		if msg == "msg2" {
			return errors.New("processing error")
		}
		fmt.Printf("Processed: %s\n", msg)
		return nil
	}, dlq)

	fmt.Println("Messages in DLQ:", dlq.GetMessages())
}

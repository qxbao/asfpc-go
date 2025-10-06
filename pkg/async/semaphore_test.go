package async

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestGetSemaphore tests semaphore creation
func TestGetSemaphore(t *testing.T) {
	t.Run("Valid limit", func(t *testing.T) {
		sem := GetSemaphore[int, int](5)
		if sem == nil {
			t.Fatal("Expected semaphore to be created")
		}
		if cap(sem.ch) != 5 {
			t.Errorf("Expected channel capacity of 5, got %d", cap(sem.ch))
		}
	})

	t.Run("Panic on zero limit", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for limit <= 0")
			}
		}()
		GetSemaphore[int, int](0)
	})

	t.Run("Panic on negative limit", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for negative limit")
			}
		}()
		GetSemaphore[int, int](-1)
	})
}

// TestAssign tests task assignment
func TestAssign(t *testing.T) {
	t.Run("Single task", func(t *testing.T) {
		sem := GetSemaphore[int, int](1)
		fn := func(x int) int { return x * 2 }
		sem.Assign(fn, 5)

		if len(sem.tasks) != 1 {
			t.Errorf("Expected 1 task, got %d", len(sem.tasks))
		}
	})

	t.Run("Multiple tasks", func(t *testing.T) {
		sem := GetSemaphore[int, int](3)
		fn := func(x int) int { return x * 2 }

		for i := 0; i < 10; i++ {
			sem.Assign(fn, i)
		}

		if len(sem.tasks) != 10 {
			t.Errorf("Expected 10 tasks, got %d", len(sem.tasks))
		}
	})

	t.Run("Cannot assign while running", func(t *testing.T) {
		sem := GetSemaphore[int, int](2)
		fn := func(x int) int {
			time.Sleep(100 * time.Millisecond) // Slow task to keep semaphore busy
			return x * 2
		}

		sem.Assign(fn, 1) // 1 * 2 = 2
		sem.Assign(fn, 2) // 2 * 2 = 4

		// Start execution in goroutine
		var firstRunResults []int
		done := make(chan struct{})
		go func() {
			firstRunResults, _ = sem.Run()
			close(done)
		}()

		time.Sleep(10 * time.Millisecond)

		assignDone := make(chan struct{})
		go func() {
			sem.Assign(fn, 3) // 3 * 2 = 6
			close(assignDone)
		}()

		<-done

		if len(firstRunResults) != 2 {
			t.Errorf("Expected 2 results from first run, got %d", len(firstRunResults))
		}

		<-assignDone

		sem.Assign(fn, 4) // 4 * 2 = 8

		results, _ := sem.Run()
		if len(results) != 2 {
			t.Errorf("Expected 2 results from second run (tasks 3 and 4), got %d", len(results))
		}
		expectedResults := []int{6, 8} // 3*2=6, 4*2=8
		for i, expected := range expectedResults {
			if i < len(results) && results[i] != expected {
				t.Errorf("Expected result[%d]=%d, got %d", i, expected, results[i])
			}
		}
	})
}

// TestRun tests basic execution
func TestRun(t *testing.T) {
	t.Run("Execute single task", func(t *testing.T) {
		sem := GetSemaphore[int, int](1)
		fn := func(x int) int { return x * 2 }
		sem.Assign(fn, 5)

		results, errs := sem.Run()

		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		if results[0] != 10 {
			t.Errorf("Expected result 10, got %d", results[0])
		}
		if errs[0] != nil {
			t.Errorf("Expected no error, got %v", errs[0])
		}
	})

	t.Run("Execute multiple tasks", func(t *testing.T) {
		sem := GetSemaphore[int, int](3)
		fn := func(x int) int { return x * 2 }

		expected := make(map[int]bool)
		for i := 0; i < 10; i++ {
			sem.Assign(fn, i)
			expected[i*2] = true
		}

		results, errs := sem.Run()

		if len(results) != 10 {
			t.Fatalf("Expected 10 results, got %d", len(results))
		}

		// Check all results are correct
		for i, result := range results {
			if !expected[result] {
				t.Errorf("Unexpected result %d at index %d", result, i)
			}
			if errs[i] != nil {
				t.Errorf("Expected no error at index %d, got %v", i, errs[i])
			}
		}
	})

	t.Run("Tasks cleared after run", func(t *testing.T) {
		sem := GetSemaphore[int, int](2)
		fn := func(x int) int { return x * 2 }

		sem.Assign(fn, 1)
		sem.Assign(fn, 2)
		sem.Run()

		if len(sem.tasks) != 0 {
			t.Errorf("Expected tasks to be cleared, got %d tasks", len(sem.tasks))
		}
	})

	t.Run("Can reuse semaphore", func(t *testing.T) {
		sem := GetSemaphore[int, int](2)
		fn := func(x int) int { return x * 2 }

		// First run
		sem.Assign(fn, 1)
		results1, _ := sem.Run()

		// Second run
		sem.Assign(fn, 2)
		results2, _ := sem.Run()

		if len(results1) != 1 || results1[0] != 2 {
			t.Errorf("First run failed: got %v", results1)
		}
		if len(results2) != 1 || results2[0] != 4 {
			t.Errorf("Second run failed: got %v", results2)
		}
	})
}

// TestConcurrencyLimit tests that semaphore respects concurrency limit
func TestConcurrencyLimit(t *testing.T) {
	limit := 3
	sem := GetSemaphore[int, int](limit)

	var concurrent atomic.Int32
	var maxConcurrent atomic.Int32

	fn := func(x int) int {
		current := concurrent.Add(1)
		defer concurrent.Add(-1)

		// Track maximum concurrent executions
		for {
			max := maxConcurrent.Load()
			if current <= max || maxConcurrent.CompareAndSwap(max, current) {
				break
			}
		}

		time.Sleep(50 * time.Millisecond) // Simulate work
		return x * 2
	}

	// Assign more tasks than the limit
	for i := 0; i < 10; i++ {
		sem.Assign(fn, i)
	}

	results, errs := sem.Run()

	if len(results) != 10 {
		t.Fatalf("Expected 10 results, got %d", len(results))
	}

	for i, err := range errs {
		if err != nil {
			t.Errorf("Task %d failed: %v", i, err)
		}
	}

	maxReached := maxConcurrent.Load()
	if maxReached > int32(limit) {
		t.Errorf("Concurrency limit exceeded: max %d, limit %d", maxReached, limit)
	}

	t.Logf("Max concurrent executions: %d (limit: %d)", maxReached, limit)
}

// TestErrorHandling tests panic recovery
func TestErrorHandling(t *testing.T) {
	t.Run("Panic with error", func(t *testing.T) {
		sem := GetSemaphore[int, int](2)

		sem.Assign(func(x int) int { return x * 2 }, 1)
		sem.Assign(func(x int) int {
			panic(errors.New("test error"))
		}, 2)
		sem.Assign(func(x int) int { return x * 3 }, 3)

		results, errs := sem.Run()

		if len(results) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(results))
		}

		// First task should succeed
		if errs[0] != nil {
			t.Errorf("Task 0 should not error: %v", errs[0])
		}
		if results[0] != 2 {
			t.Errorf("Expected result 2, got %d", results[0])
		}

		// Second task should panic with error
		if errs[1] == nil {
			t.Error("Task 1 should have error")
		} else if errs[1].Error() != "test error" {
			t.Errorf("Expected 'test error', got %v", errs[1])
		}

		// Third task should succeed
		if errs[2] != nil {
			t.Errorf("Task 2 should not error: %v", errs[2])
		}
		if results[2] != 9 {
			t.Errorf("Expected result 9, got %d", results[2])
		}
	})

	t.Run("Panic with string", func(t *testing.T) {
		sem := GetSemaphore[int, int](1)

		sem.Assign(func(x int) int {
			panic("string panic")
		}, 1)

		_, errs := sem.Run()

		if errs[0] == nil {
			t.Error("Expected error from panic")
		} else if errs[0].Error() != "string panic" {
			t.Errorf("Expected 'string panic', got %v", errs[0])
		}
	})

	t.Run("Panic with integer", func(t *testing.T) {
		sem := GetSemaphore[int, int](1)

		sem.Assign(func(x int) int {
			panic(42)
		}, 1)

		_, errs := sem.Run()

		if errs[0] == nil {
			t.Error("Expected error from panic")
		} else if errs[0].Error() != "42" {
			t.Errorf("Expected '42', got %v", errs[0])
		}
	})
}

// TestThreadSafety tests concurrent access
func TestThreadSafety(t *testing.T) {
	sem := GetSemaphore[int, int](5)
	fn := func(x int) int { return x * 2 }

	var wg sync.WaitGroup
	wg.Add(10)

	// Try to assign from multiple goroutines
	for i := 0; i < 10; i++ {
		go func(val int) {
			defer wg.Done()
			sem.Assign(fn, val)
		}(i)
	}

	wg.Wait()

	results, errs := sem.Run()

	if len(results) != 10 {
		t.Fatalf("Expected 10 results, got %d", len(results))
	}

	for i, err := range errs {
		if err != nil {
			t.Errorf("Task %d failed: %v", i, err)
		}
	}
}

// TestDifferentTypes tests semaphore with different type parameters
func TestDifferentTypes(t *testing.T) {
	t.Run("String to int", func(t *testing.T) {
		sem := GetSemaphore[string, int](2)
		fn := func(s string) int { return len(s) }

		sem.Assign(fn, "hello")
		sem.Assign(fn, "world!")

		results, errs := sem.Run()

		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}
		if results[0] != 5 || results[1] != 6 {
			t.Errorf("Expected [5, 6], got %v", results)
		}
		if errs[0] != nil || errs[1] != nil {
			t.Errorf("Expected no errors, got %v", errs)
		}
	})

	t.Run("Struct types", func(t *testing.T) {
		type Input struct {
			Name string
			Age  int
		}
		type Output struct {
			Message string
		}

		sem := GetSemaphore[Input, Output](2)
		fn := func(in Input) Output {
			return Output{Message: fmt.Sprintf("%s is %d years old", in.Name, in.Age)}
		}

		sem.Assign(fn, Input{Name: "Alice", Age: 30})
		sem.Assign(fn, Input{Name: "Bob", Age: 25})

		results, errs := sem.Run()

		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}
		if results[0].Message != "Alice is 30 years old" {
			t.Errorf("Unexpected result: %s", results[0].Message)
		}
		if results[1].Message != "Bob is 25 years old" {
			t.Errorf("Unexpected result: %s", results[1].Message)
		}
		if errs[0] != nil || errs[1] != nil {
			t.Errorf("Expected no errors, got %v", errs)
		}
	})
}

// TestEmptyRun tests running with no tasks
func TestEmptyRun(t *testing.T) {
	sem := GetSemaphore[int, int](5)

	results, errs := sem.Run()

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
	if len(errs) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(errs))
	}
}

// BenchmarkSemaphore benchmarks semaphore performance
func BenchmarkSemaphore(b *testing.B) {
	b.Run("Limit-1", func(b *testing.B) {
		benchmarkSemaphoreWithLimit(b, 1)
	})

	b.Run("Limit-5", func(b *testing.B) {
		benchmarkSemaphoreWithLimit(b, 5)
	})

	b.Run("Limit-10", func(b *testing.B) {
		benchmarkSemaphoreWithLimit(b, 10)
	})

	b.Run("Limit-100", func(b *testing.B) {
		benchmarkSemaphoreWithLimit(b, 100)
	})
}

func benchmarkSemaphoreWithLimit(b *testing.B, limit int) {
	fn := func(x int) int {
		// Simulate some work
		sum := 0
		for i := 0; i < 1000; i++ {
			sum += i
		}
		return x * 2
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sem := GetSemaphore[int, int](limit)
		for j := 0; j < 100; j++ {
			sem.Assign(fn, j)
		}
		sem.Run()
	}
}

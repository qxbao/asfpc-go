package async

import (
	"fmt"
	"sync"
)

type tsk[I any, O any] struct {
	fn  func(I) O
	arg I
}

type Semaphore[I any, O any] struct {
	ch      chan struct{}
	tasks   []tsk[I, O]
	mu      sync.Mutex
	started bool
}

func GetSemaphore[I any, O any](limit int) *Semaphore[I, O] {
	if limit <= 0 {
		panic("Semaphore limit must be greater than 0")
	}

	return &Semaphore[I, O]{
		ch:      make(chan struct{}, limit),
		tasks:   make([]tsk[I, O], 0),
		mu:      sync.Mutex{},
		started: false,
	}
}

func (s *Semaphore[I, O]) Assign(fn func(I) O, arg I) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return
	}
	s.tasks = append(s.tasks, tsk[I, O]{fn: fn, arg: arg})
}

func (s *Semaphore[I, O]) Run() ([]O, []error) {
	s.mu.Lock()
	s.started = true
	tasks := append([]tsk[I, O](nil), s.tasks...)

	defer func() {
		s.started = false
		s.mu.Unlock()
		s.tasks = make([]tsk[I, O], 0)
	}()

	errors := make([]error, len(tasks))
	results := make([]O, len(tasks))

	var wg sync.WaitGroup
	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t tsk[I, O]) {
			defer wg.Done()

			s.ch <- struct{}{}

			defer func() { <-s.ch }()
			defer func() {
				if r := recover(); r != nil {
					if err, ok := r.(error); ok {
						errors[idx] = err
					} else {
						errors[idx] = fmt.Errorf("%v", r)
					}
				}
			}()

			results[idx] = t.fn(t.arg)
		}(i, task)
	}
	wg.Wait()
	return results, errors
}

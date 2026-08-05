package health

import (
	"context"
	"sync"
	"time"
)

type Status string

const (
	StatusOK   Status = "ok"
	StatusFail Status = "fail"
)

type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

type Result struct {
	Status  Status `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

type Response struct {
	Status    Status            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]Result `json:"checks"`
}

type Aggregator struct {
	checkers []Checker
	timeout  time.Duration
}

func New(timeout time.Duration, checkers ...Checker) *Aggregator {
	return &Aggregator{checkers: checkers, timeout: timeout}
}

func (a *Aggregator) Run(ctx context.Context) Response {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results = make(map[string]Result, len(a.checkers))
		overall = StatusOK
	)

	for _, c := range a.checkers {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()
			start := time.Now()
			err := c.Check(ctx)
			r := Result{Latency: time.Since(start).String()}
			if err != nil {
				r.Status = StatusFail
				r.Error = err.Error()
			} else {
				r.Status = StatusOK
			}
			mu.Lock()
			results[c.Name()] = r
			if err != nil {
				overall = StatusFail
			}
			mu.Unlock()
		}(c)
	}
	wg.Wait()

	return Response{
		Status:    overall,
		Timestamp: time.Now(),
		Checks:    results,
	}
}

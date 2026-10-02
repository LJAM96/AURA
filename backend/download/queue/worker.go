package downloadqueue

import (
	"aura/logging"
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

const workerSafetyTickInterval = 30 * time.Second

var (
	wakeCh   = make(chan struct{}, 1)
	draining atomic.Bool
	workerID = fmt.Sprintf("worker-%d", time.Now().UnixNano())
)

// SignalNewJob wakes the worker loop to check for new pending jobs. Safe to
// call from any goroutine; bursts of calls coalesce into a single wake.
func SignalNewJob() {
	select {
	case wakeCh <- struct{}{}:
	default:
	}
}

// StartWorker starts the background goroutine that claims and processes
// Download Queue Jobs one at a time. It is woken by SignalNewJob (fired on
// enqueue) and by a periodic safety-net tick, so a job can never be picked up
// twice: claiming is an atomic DB operation regardless of what triggered it.
func StartWorker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(workerSafetyTickInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-wakeCh:
			case <-ticker.C:
			}
			drainQueue(ctx)
		}
	}()
}

func drainQueue(ctx context.Context) {
	if !draining.CompareAndSwap(false, true) {
		// Another drain is already in progress; it will pick up any jobs this signal was for.
		return
	}
	defer draining.Store(false)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		processedOne := processNextJob(ctx)
		if !processedOne {
			return
		}
	}
}

func processNextJob(ctx context.Context) (processed bool) {
	defer func() {
		if r := recover(); r != nil {
			logging.LOGGER.Error().Timestamp().Interface("recover", r).Msg("PANIC: in Download Queue worker")
		}
	}()

	return ProcessNextQueueJob(ctx)
}

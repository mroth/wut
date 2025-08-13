package wut

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

func NewRunnerWithExecutor(ctx context.Context, executor mockExecutor) *Runner {
	r := NewRunner(ctx, "")
	r.executor = executor
	return r
}

func TestRunner_Run(t *testing.T) {
	t.Run("simple success", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			r := NewRunnerWithExecutor(t.Context(), mockExecutor{exitcode: 0})

			runAssert(t, r, runnerExpectedResults{
				err:  nil,
				runs: 1,
			})
		})
	})

	t.Run("fails when exceeds max runs", func(t *testing.T) {
		t.Parallel()

		synctest.Test(t, func(t *testing.T) {
			r := NewRunnerWithExecutor(t.Context(), mockExecutor{exitcode: 1})
			r.MaxRuns = 3
			r.RetryDelay = 10 * time.Millisecond

			runAssert(t, r, runnerExpectedResults{
				err:          errMaxRunsCompleted,
				runs:         3,
				elapsedTotal: 30 * time.Millisecond,
			})
		})
	})

	t.Run("context deadline exceeded", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			// set overall runner context timeout for 45ms,
			// and have runner run a failing command every 10ms via retry delay.
			ctx, cancel := context.WithTimeout(t.Context(), 45*time.Millisecond)
			t.Cleanup(cancel)
			r := NewRunnerWithExecutor(ctx, mockExecutor{exitcode: 1})
			r.RetryDelay = 10 * time.Millisecond

			runAssert(t, r, runnerExpectedResults{
				err:          context.DeadlineExceeded,
				runs:         5,
				elapsedTotal: 45 * time.Millisecond,
			})
		})
	})

	t.Run("process timeout", func(t *testing.T) {
		// Run a command that sleeps for 100ms before success, but with a process timeout of 50ms.
		//
		// We should see MaxRuns failures occur over a total 150ms, as each process is canceled
		// due to a deadline exceeded error before it completes.
		var (
			executorSleepDuration = 100 * time.Millisecond
			runnerProcessTimeout  = 50 * time.Millisecond
			runnerMaxRuns         = 3
			wantErr               = errMaxRunsCompleted
			wantRunsCompleted     = uint(3)
			wantTotalElapsed      = time.Duration(runnerMaxRuns) * runnerProcessTimeout
		)

		synctest.Test(t, func(t *testing.T) {
			r := NewRunnerWithExecutor(t.Context(), mockExecutor{sleep: executorSleepDuration})
			r.ProcessTimeout = runnerProcessTimeout
			r.MaxRuns = uint(runnerMaxRuns)

			runAssert(t, r, runnerExpectedResults{
				err:          wantErr,
				runs:         wantRunsCompleted,
				elapsedTotal: wantTotalElapsed,
			})
		})
	})
}

// runAssert runs the given runner and asserts that it completes with the expected results.
func runAssert(t *testing.T, r *Runner, want runnerExpectedResults) {
	t.Helper()

	start := time.Now()
	err := r.Run()
	elapsed := time.Since(start)

	if !errors.Is(err, want.err) {
		t.Errorf("error: got %v, want %v", err, want.err)
	}

	if r.runsCompleted != want.runs {
		t.Errorf("runs completed: got %d, want %d", r.runsCompleted, want.runs)
	}

	if elapsed != want.elapsedTotal {
		t.Errorf("Expected duration %v, got %v", want.elapsedTotal, elapsed)
	}
}

type runnerExpectedResults struct {
	err          error
	runs         uint
	elapsedTotal time.Duration
}

package wut

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	testMode         = flag.Bool("testmode", false, "Enable test mode helper binary")
	testModeSleep    = flag.Duration("testmode-sleep", 0, "Sleep duration for test mode helper binary")
	testModeOutput   = flag.String("testmode-output", "", "Prints output for test mode helper binary")
	testModeExitcode = flag.Int("testmode-exitcode", 0, "Exit code for test mode helper binary")
)

// TestMain provides a test helper that can emulate different command behaviors
func TestMain(m *testing.M) {
	flag.Parse()
	if *testMode {
		emulateTestBinary()
		return
	}

	os.Exit(m.Run())
}

// emulateTestBinary implements the behavior for our test helper binary
func emulateTestBinary() {
	time.Sleep(*testModeSleep) // Simulate some processing time
	if *testModeOutput != "" {
		fmt.Println(*testModeOutput)
	}
	os.Exit(*testModeExitcode) // Exit with the specified code
}

type binaryActions struct {
	sleep    time.Duration
	output   string
	exitcode int
}

func (ba binaryActions) Args() []string {
	args := []string{"-testmode"}
	if ba.sleep > 0 {
		args = append(args, "-testmode-sleep", ba.sleep.String())
	}
	if ba.output != "" {
		args = append(args, "-testmode-output", ba.output)
	}
	if ba.exitcode != 0 {
		args = append(args, "-testmode-exitcode", fmt.Sprintf("%d", ba.exitcode))
	}
	return args
}

func TestRunner_Run(t *testing.T) {
	t.Run("simple success", func(t *testing.T) {
		t.Parallel()

		args := binaryActions{exitcode: 0}.Args()
		r := NewRunner(t.Context(), os.Args[0], args...)

		err := r.Run()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("fails when exceeds max runs", func(t *testing.T) {
		t.Parallel()

		args := binaryActions{exitcode: 1}.Args()
		r := NewRunner(t.Context(), os.Args[0], args...)
		r.MaxRuns = 3
		r.RetryDelay = 10 * time.Millisecond

		err := r.Run()
		if !errors.Is(err, errMaxRunsCompleted) {
			t.Errorf("Expected errMaxRunsCompleted, got %v", err)
		}
		if r.runsCompleted != 3 {
			t.Errorf("Expected 3 runs completed, got %d", r.runsCompleted)
		}
	})

	t.Run("context deadline exceeded", func(t *testing.T) {
		t.Parallel()

		// cancel context after 50ms
		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		t.Cleanup(cancel)

		// run a failing command every ~10ms
		args := binaryActions{exitcode: 1}.Args()
		r := NewRunner(ctx, os.Args[0], args...)
		r.RetryDelay = 10 * time.Millisecond

		start := time.Now()
		err := r.Run()
		elapsed := time.Since(start)
		t.Log("Elapsed time:", elapsed)
		t.Log("Runs completed:", r.runsCompleted)

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected context.DeadlineExceed, got %v", err)
		}

		if elapsed < 50*time.Millisecond {
			t.Errorf("Expected at least 50ms elapsed, got %v", elapsed)
		}

		if r.runsCompleted < 4 {
			t.Errorf("Expected at least 4 runs completed, got %d", r.runsCompleted)
		}
	})

}

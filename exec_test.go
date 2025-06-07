package wut

import (
	"context"
	"fmt"
	"time"
)

// A mockExecutor is a mock implementation of the executor interface for testing purposes.
type mockExecutor struct {
	sleep    time.Duration // 1. First, we sleep for the specified duration (simulating processing time)
	output   string        // 2. Then, we write output to the standard output stream
	exitcode int           // 3. Finally, we exit with the specified exit code
}

// verify mockExecutor implements the executor interface
var _ executor = mockExecutor{}

func (me mockExecutor) Run(ctx context.Context, opts CommandOpts, name string, args ...string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(me.sleep):
		if me.output != "" {
			_, err := opts.Stdout.Write([]byte(me.output))
			if err != nil {
				return err
			}
		}

		if me.exitcode != 0 {
			return fmt.Errorf("mock command failure with exit code %d", me.exitcode)
		}
		return nil
	}
}

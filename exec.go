package wut

import (
	"context"
	"os/exec"
)

// executor is used to abstract command execution for testing purposes.
type executor interface {
	Run(ctx context.Context, opts CommandOpts, name string, args ...string) error
}

// cmdExecutor is the default implementation of the executor interface.
// It uses the os/exec package to run commands in a subprocess.
type cmdExecutor struct{}

// verify cmdExecutor implements the executor interface
var _ executor = cmdExecutor{}

func (ce cmdExecutor) Run(ctx context.Context, opts CommandOpts, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = opts.Env
	cmd.Dir = opts.Dir
	cmd.Stdin = opts.Stdin
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr
	cmd.WaitDelay = opts.WaitDelay
	if opts.Cancel != nil {
		cmd.Cancel = opts.Cancel // not safe to set to nil
	}
	return cmd.Run()
}

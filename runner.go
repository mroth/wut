// Package wut provides a command runner that can execute an [exec.Cmd] repeatedly until success, with configurable retry and timeout capabilities.
package wut

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"
)

// Runner manages the repeated execution of commands with retry and timeout capabilities.
type Runner struct {
	name    string
	args    []string
	baseCtx context.Context

	// ProcessTimeout is the timeout duration for individual command run execution.
	// If a command execution does not complete within this duration, it will be cancelled.
	ProcessTimeout time.Duration

	// RetryDelay is the delay between retries of the command execution.
	RetryDelay time.Duration

	// MaxRuns is the maximum number of times the command will be executed before the Runner stops.
	// If MaxRuns is set to 0, there will be no cap on the number of times the command can be run,
	// prior to the Runner encountering another stop condition.
	MaxRuns uint

	// ContinueOnSuccess allows the Runner to continue executing commands even after a successful run.
	ContinueOnSuccess bool

	// CommandOptions are options for the underlying process command execution.
	CommandOptions CommandOpts

	runlock       sync.Mutex // locked when a command is running
	runsCompleted uint
	executor      executor
	logger        *slog.Logger
}

// CommandOpts provides options to configure the execution of [exec.Cmd] commands.
//
// These options are passed to the underlying command execution by the Runner
// and can be used to customize the environment, working directory, input/output
// streams, and other behavior.
type CommandOpts struct {
	Env       []string      // environment variables to set for Cmd execution, see https://pkg.go.dev/os/exec#Cmd.Env
	Dir       string        // working directory for Cmd execution, see https://pkg.go.dev/os/exec#Cmd.Dir
	Stdin     io.Reader     // standard input for Cmd execution, see https://pkg.go.dev/os/exec#Cmd.Stdin
	Stdout    io.Writer     // standard output for Cmd execution, see https://pkg.go.dev/os/exec#Cmd.Stdout
	Stderr    io.Writer     // standard error for Cmd execution, see https://pkg.go.dev/os/exec#Cmd.Stderr
	Cancel    func() error  // cancel function for Cmd processeses, see https://pkg.go.dev/os/exec#Cmd.Cancel
	WaitDelay time.Duration // wait delay for Cmd processeses, see https://pkg.go.dev/os/exec#Cmd.WaitDelay
}

var (
	errMaxRunsCompleted = errors.New("wut: maximum number of runs completed")
	// errRedundantStartCall = errors.New("wut: runner already started")
	// errRedundantWaitCall  = errors.New("wut: runner already waiting for completion")
)

// NewRunner creates a new Runner instance with the provided context, name, and command arguments.
//
// To set a timeout for the entire runner, provide a context with an appropriate timeout.
//
// Similarly, to stop execution of the runner prior to completion or failure, provide a context with a cancellation function.
func NewRunner(ctx context.Context, name string, arg ...string) *Runner {
	return &Runner{
		name:     name,
		args:     arg,
		baseCtx:  ctx,
		executor: cmdExecutor{},
		logger:   slog.New(slog.DiscardHandler),
	}
}

// SetLogger sets the logger for the Runner.
// If nil, it will use a discard logger.
func (r *Runner) SetLogger(logger *slog.Logger) {
	if logger != nil {
		r.logger = logger
	} else {
		r.logger = slog.New(slog.DiscardHandler)
	}
}

// Run starts the Runner and executes the command repeatedly until it succeeds or a stop condition is reached.
func (r *Runner) Run() error {
	r.logger.Info("Starting runner", "command", r.name, "args", r.args)
	for {
		select {
		case <-r.baseCtx.Done():
			r.logger.Warn("Runner stopped", "reason", context.Cause(r.baseCtx))
			return context.Cause(r.baseCtx)
		case <-time.After(r.nextExecDelay()):
			if r.MaxRuns > 0 && r.runsCompleted >= r.MaxRuns {
				r.logger.Warn("Runner stopped", "reason", errMaxRunsCompleted)
				return errMaxRunsCompleted
			}

			err := r.executeCommand()
			r.logger.Info("Command executed", "error", err)
			if err == nil && !r.ContinueOnSuccess {
				r.logger.Info("Completed successfully", "name", r.name, "attempts", r.runsCompleted)
				return nil
			}
		}
	}
}

func (r *Runner) nextExecDelay() time.Duration {
	r.runlock.Lock()
	defer r.runlock.Unlock()

	if r.runsCompleted == 0 {
		return 0 // no delay for the first run
	}
	return r.RetryDelay
}

func (r *Runner) executeCommand() error {
	r.runlock.Lock()
	defer r.runlock.Unlock()

	ctx := r.baseCtx
	if r.ProcessTimeout > 0 {
		pctx, cf := context.WithTimeout(r.baseCtx, r.ProcessTimeout)
		ctx = pctx
		defer cf()
	}

	defer func() {
		r.runsCompleted++
	}()

	return r.executor.Run(ctx, r.CommandOptions, r.name, r.args...)
}

// func (r *Runner) Stop() error
//
// NOTE: If we want to implement a Stop method, we need to handle the
// cancellation of the running command. The easiest way to do this is probably
// to wrap the base context in a cancellable context and call cancel on it,
// which will propagate to the command's context (use a CancelCauseFunc to track
// reason).  For now though, let's keep this out of the API to simplify it and
// the caller can handle cancellation themselves by providing a context with a
// cancellation function when creating the Runner.

// Keep the complexity of exposing internal state out of the API for now.  In
// the future, we may want to expose the state of the Runner via something like
// this.
//
// // RunnerState represents the state of the Runner.
// type RunnerState int

// const (
// 	RunnerStateIdle      RunnerState = iota // Runner is active but is not currently executing a command.
// 	RunnerStateRunning                      // Runner is active and is currently executing a command.
// 	RunnerStateCompleted                    // Runner has completed its success criteria and is no longer running.
// 	RunnerStateErrored                      // Runner met an exit/failure condition prior to success (for example, timed out).
// )

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/mroth/wut"
)

var (
	timeout           = flag.Duration("timeout", 0, "maximum time to wait for a successful execution")
	retryDelay        = flag.Duration("retry-delay", time.Second, "delay between retries")
	maxRuns           = flag.Uint("max-runs", 0, "maximum number of times to run the command (default unlimited)")
	continueOnSuccess = flag.Bool("continue", false, "continue running even after successful execution")
)

const (
	banner     = `wut - a command runner with retry and timeout capabilities`
	usageShort = `Usage: wut [OPTIONS] COMMAND [ARGS]...`
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, banner)
		fmt.Fprintln(os.Stderr, usageShort)
		fmt.Fprintln(os.Stderr, "\nOptions:")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(125)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if *timeout > 0 {
		tctx, cf := context.WithTimeoutCause(ctx, *timeout, errors.New("timeout exceeded"))
		ctx = tctx
		defer cf()
	}

	runner := wut.NewRunner(ctx, flag.Arg(0), flag.Args()[1:]...)
	runner.ContinueOnSuccess = *continueOnSuccess
	runner.MaxRuns = *maxRuns
	runner.RetryDelay = *retryDelay

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	runner.SetLogger(logger)

	err := runner.Run()
	if err != nil {
		logger.Error("Runner encountered an error", "error", err)
		os.Exit(1)
	}
}

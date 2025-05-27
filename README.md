# 💁 wut - _Wait Until/Timeout_

A handy combination of the `until` and `timeout` Linux utilities, with added support for retry
delay, and other nice things.

## Example

The following standard Linux utilies are quite handy:

 * `until`: Execute commands as long as a test does not succeed.
 * `timeout`: Run a command with a time limit.

A common pattern is to combine them as such, for example to wait for a server to successfully boot before performing another action:

```sh
timeout 1m bash -c "until curl --silent --fail 127.0.0.1:8080/health; do
	sleep 5
done"
```

_(Since `until` is a shell function, it is necessary to wrap it in a shell invocation as above to work with `timeout`.)_


With **`wut`**, this is as simple as:

    wut --retry-delay=5s --timeout=2m curl --silent --fail 127.0.0.1:8080/health

Critically, this also does not require access to a shell, handy for running in minimal containers.

## Usage

    wut - a command runner with retry and timeout capabilities
    Usage: wut [OPTIONS] COMMAND [ARGS]...

    Options:
    -continue
            continue running even after successful execution
    -max-runs uint
            maximum number of times to run the command (default unlimited)
    -retry-delay duration
            delay between retries (default 1s)
    -timeout duration
            maximum time to wait for a successful execution


## Installation

<!-- Uncomment once binary releases are enabled. -->
<!-- Download a binary from the [releases page][1] and place somewhere on your path. -->

If you have a Go toolchain installed, you can get the latest via:

    go install github.com/mroth/wut/cmd/wut@latest

[1]: https://github.com/mroth/wut/releases

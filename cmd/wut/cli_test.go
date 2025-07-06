package main

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e tests in short mode")
	}
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"wut":           main,
		"bintrue":       binTrue,
		"binfalse":      binFalse,
		"succeed-after": succeedAfterAttempts,
	})
}

func binTrue() {
	os.Exit(0)
}

func binFalse() {
	os.Exit(1)
}

func succeedAfterAttempts() {
	var (
		filename = flag.String("file", "attempts.dat", "data file to track attempts")
		fails    = flag.Int("fails", 5, "number of times to fail before succeeding")
	)
	flag.Parse()

	val := 0
	if data, err := os.ReadFile(*filename); err == nil {
		fmt.Sscanf(string(data), "%d", &val)
	}
	val++
	os.WriteFile(*filename, []byte(fmt.Sprint(val)), 0644)
	if val <= *fails {
		os.Exit(val)
	}
}

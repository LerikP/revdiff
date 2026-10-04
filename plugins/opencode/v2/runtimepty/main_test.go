//go:build darwin || linux

package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if os.Getenv("REVDIFF_PTY_TEST_HELPER") == "1" {
		os.Args = append(os.Args[:1], os.Args[2:]...)
		main()
		return
	}
	os.Exit(m.Run())
}

func TestTerminalIOAndExitCode(t *testing.T) {
	command := helperCommand(t, "/bin/sh", "-c", `
test -t 0 && test -t 1 && test -t 2 || exit 90
stty size < /dev/tty
IFS= read -r value
printf 'received:%s\nargument:%s\n' "$value" "$1"
printf 'stderr-line\n' >&2
exit 7
`, "fixture", "a b")
	input, err := command.StdinPipe()
	require.NoError(t, err)
	t.Cleanup(func() { _ = input.Close() })
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	require.NoError(t, command.Start())
	_, err = io.WriteString(input, "hello\n")
	require.NoError(t, err)
	var exited *exec.ExitError
	require.ErrorAs(t, command.Wait(), &exited)
	require.Equal(t, 7, exited.ExitCode(), output.String())
	require.Contains(t, output.String(), "35 120")
	require.Contains(t, output.String(), "received:hello")
	require.Contains(t, output.String(), "argument:a b")
	require.Contains(t, output.String(), "stderr-line")
}

func TestCancellationReapsTerminalChild(t *testing.T) {
	command := helperCommand(t, "/bin/sh", "-c", `trap '' TERM HUP; printf '%s\n' "$$"; while :; do sleep 1; done`)
	input, err := command.StdinPipe()
	require.NoError(t, err)
	t.Cleanup(func() { _ = input.Close() })
	output, err := command.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, command.Start())
	line, err := bufio.NewReader(output).ReadString('\n')
	require.NoError(t, err)
	pid, err := strconv.Atoi(strings.TrimSpace(line))
	require.NoError(t, err)
	reaped := false
	t.Cleanup(func() {
		if !reaped {
			_ = syscall.Kill(-pid, syscall.SIGKILL)
		}
	})
	require.NoError(t, command.Process.Signal(syscall.SIGTERM))
	var exited *exec.ExitError
	require.ErrorAs(t, command.Wait(), &exited)
	require.Equal(t, 137, exited.ExitCode())
	require.ErrorIs(t, syscall.Kill(pid, 0), syscall.ESRCH)
	reaped = true
}

func helperCommand(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()
	executable, err := os.Executable()
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, executable, append([]string{"--"}, args...)...) //nolint:gosec // executes this test binary with fixed fixture arguments
	command.Env = append(os.Environ(), "REVDIFF_PTY_TEST_HELPER=1")
	return command
}

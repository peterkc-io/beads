package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ErrSelfTestBinary prevents self-calls from re-executing a Go test binary.
var ErrSelfTestBinary = errors.New("running as test binary - cannot execute bd subcommands")

var selfExecutable = os.Executable

// SelfCommand builds a command for the running executable, never one from PATH.
// caller names the function whose self-call failed.
func SelfCommand(caller string, args ...string) (*exec.Cmd, error) {
	exe, err := selfExecutable()
	if err != nil {
		return nil, fmt.Errorf("%s: resolve executable: %w", caller, err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	base := filepath.Base(exe)
	if testing.Testing() || strings.HasSuffix(base, ".test") || strings.Contains(base, ".test.") {
		return nil, fmt.Errorf("%s: %w", caller, ErrSelfTestBinary)
	}
	return exec.Command(exe, args...), nil // #nosec G204 -- exe is the running executable
}

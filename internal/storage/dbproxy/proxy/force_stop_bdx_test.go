//go:build linux || darwin || windows

package proxy

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/dbproxy/pidfile"
	"github.com/steveyegge/beads/internal/testutil/bazeltest"
)

func TestBdxForceStopExecutableNames(t *testing.T) {
	for _, name := range []string{"bdx", "other"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			exe, err = filepath.EvalSymlinks(exe)
			if err != nil {
				t.Fatal(err)
			}
			base := name
			if runtime.GOOS == "windows" {
				base += ".exe"
			}
			path := filepath.Join(root, base)
			if err := os.Link(exe, path); err != nil {
				src, err := os.Open(exe)
				if err != nil {
					t.Fatal(err)
				}
				defer src.Close()
				dst, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o700)
				if err != nil {
					t.Fatal(err)
				}
				_, copyErr := io.Copy(dst, src)
				closeErr := dst.Close()
				if copyErr != nil || closeErr != nil {
					t.Fatalf("copy test executable: %v, %v", copyErr, closeErr)
				}
			}
			cmd := exec.Command(path, "-test.run=^TestBdxForceStopHelperProcess$")
			cmd.Env = append(bazeltest.ShardFreeEnv(os.Environ()), "BEADS_BDX_FORCE_STOP_HELPER=1")
			input, err := cmd.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			defer input.Close()
			output, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan struct{})
			go func() {
				_ = cmd.Wait()
				close(done)
			}()
			t.Cleanup(func() {
				_ = cmd.Process.Kill()
				<-done
			})
			ready := make([]byte, 1)
			if _, err := io.ReadFull(output, ready); err != nil || string(ready) != "!" {
				t.Fatalf("helper not ready: %q, %v", ready, err)
			}
			if err := pidfile.Write(root, PIDFileName, pidfile.PidFile{Pid: cmd.Process.Pid, Port: 3307}); err != nil {
				t.Fatal(err)
			}
			report, err := ForceStopUnverified(root)
			if name == "bdx" {
				if err != nil || !report.SignalSent || report.QuarantinedPath == "" {
					t.Fatalf("bdx must be accepted: report=%+v, err=%v", report, err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "executable basename") || report.SignalSent || report.QuarantinedPath != "" {
				t.Fatalf("other executable must be left alone: report=%+v, err=%v", report, err)
			}
			if _, err := os.Stat(pidfile.Path(root, PIDFileName)); err != nil {
				t.Fatalf("refusal changed the PID record: %v", err)
			}
		})
	}
}

func TestBdxForceStopHelperProcess(t *testing.T) {
	if os.Getenv("BEADS_BDX_FORCE_STOP_HELPER") != "1" {
		t.Skip("force-stop subprocess fixture")
	}
	fmt.Print("!")
	_, _ = io.Copy(io.Discard, os.Stdin)
}

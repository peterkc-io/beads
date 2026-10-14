package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/dbproxy/proxy"
)

func TestBdxForceStopResultExecutableVerified(t *testing.T) {
	for _, name := range []string{"bd", "bdx", "dolt", "other"} {
		t.Run(name, func(t *testing.T) {
			result := newForcedDoltStopResult(errors.New("unverified"), proxy.ForceStopReport{Executable: name}, nil)
			want := name != "other"
			if result.ExecutableVerified == nil || *result.ExecutableVerified != want {
				t.Fatalf("executable_verified for %q = %v, want %v", name, result.ExecutableVerified, want)
			}
		})
	}
}

func TestBdxDuplicateBinaries(t *testing.T) {
	var dirs, want []string
	for range 2 {
		dir := t.TempDir()
		dirs = append(dirs, dir)
		for _, name := range []string{"bd", "bdx"} {
			if runtime.GOOS == "windows" {
				name += ".exe"
			}
			path := filepath.Join(dir, name)
			if err := os.WriteFile(path, []byte("fixture"), 0o700); err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(name, "bdx") {
				want = append(want, path)
			}
		}
	}
	t.Setenv("PATH", strings.Join(dirs, string(os.PathListSeparator)))
	if got := findDuplicateBinaries(); !slices.Equal(got, want) {
		t.Fatalf("duplicate binaries = %q, want %q", got, want)
	}
}

func TestBdxTopLevelVersionSkipsStore(t *testing.T) {
	binary := buildBDForInitTests(t)
	cmd := exec.Command(binary, "version")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "BEADS_DIR="+cmd.Dir, "BD_DISABLE_METRICS=1", "BD_NAME=bdx")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bdx version must not open a store: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "version") {
		t.Fatalf("missing version output: %s", out)
	}
}

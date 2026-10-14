package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/config"
)

func TestBdxSelfCallsRefuseTestBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PATH trap is a POSIX shell script")
	}
	isolateBeadsDirForTest(t)
	root := t.TempDir()
	beadsDir := filepath.Join(root, ".beads")
	writeFileForWorktreeTest(t, filepath.Join(beadsDir, "metadata.json"), []byte("{}\n"))
	writeFileForWorktreeTest(t, filepath.Join(beadsDir, "issues.jsonl"), []byte("{}\n"))
	runGitForWorktreeTest(t, root, "init", "-q")
	runGitForWorktreeTest(t, root, "add", ".beads/metadata.json")
	t.Chdir(root)
	t.Setenv("BEADS_DIR", beadsDir)
	binDir := t.TempDir()
	trap := filepath.Join(binDir, "called")
	body := "#!/bin/sh\nprintf called > '" + strings.ReplaceAll(trap, "'", `'\''`) + "'\n"
	if err := os.WriteFile(filepath.Join(binDir, "bd"), []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	initConfigForTest(t)
	config.Set("export.auto", true)
	config.Set("export.path", "issues.jsonl")
	config.Set("import.auto", true)
	config.Set("sync.remote", "")
	config.Set("sync.git-remote", "")
	for _, tc := range []struct {
		name string
		run  func()
	}{
		{"exportJSONLForCommit", exportJSONLForCommit},
		{"importJSONLForSync", func() { importJSONLForSync("test") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := captureStderr(t, tc.run)
			if !strings.Contains(out, tc.name+":") || !strings.Contains(out, "test binary") {
				t.Errorf("want contextual test-binary refusal, got %q", out)
			}
		})
	}
	if err := closeIssueRunner("fixture-1"); err == nil || !strings.Contains(err.Error(), "closeIssueRunner:") || !strings.Contains(err.Error(), "test binary") {
		t.Errorf("closeIssueRunner must refuse test binary with context, got %v", err)
	}
	if _, err := os.Stat(trap); !os.IsNotExist(err) {
		t.Fatalf("a self-call ran bd from PATH: %v", err)
	}
}

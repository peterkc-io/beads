package doctor

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBdxPrimeSelfCallsRefuseTestBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PATH trap is a POSIX shell script")
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("Use bd prime"), 0o600); err != nil {
		t.Fatal(err)
	}
	binDir := t.TempDir()
	trap := filepath.Join(binDir, "called")
	body := "#!/bin/sh\nprintf called > '" + strings.ReplaceAll(trap, "'", `'\''`) + "'\nprintf context\n"
	if err := os.WriteFile(filepath.Join(binDir, "bd"), []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, tc := range []struct {
		name  string
		check func(string) DoctorCheck
	}{
		{"VerifyPrimeOutput", VerifyPrimeOutput},
		{"CheckDocumentationBdPrimeReference", CheckDocumentationBdPrimeReference},
	} {
		t.Run(tc.name, func(t *testing.T) {
			check := tc.check(root)
			if !strings.Contains(check.Detail, tc.name+":") || !strings.Contains(check.Detail, "test binary") {
				t.Errorf("want contextual test-binary refusal, got %+v", check)
			}
		})
	}
	if _, err := os.Stat(trap); !os.IsNotExist(err) {
		t.Fatalf("a self-call ran bd from PATH: %v", err)
	}
}

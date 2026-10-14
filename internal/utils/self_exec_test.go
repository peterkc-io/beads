package utils

import (
	"errors"
	"strings"
	"testing"
)

func TestSelfCommandRefusesTestBinary(t *testing.T) {
	cmd, err := SelfCommand("TestSelfCommandRefusesTestBinary", "prime")
	if cmd != nil || !errors.Is(err, ErrSelfTestBinary) || !strings.Contains(err.Error(), "TestSelfCommandRefusesTestBinary:") {
		t.Fatalf("want contextual test-binary refusal and no command, got %v, %v", cmd, err)
	}
}

func TestSelfCommandExecutableError(t *testing.T) {
	original := selfExecutable
	failure := errors.New("no executable")
	selfExecutable = func() (string, error) { return "", failure }
	t.Cleanup(func() { selfExecutable = original })
	cmd, err := SelfCommand("caller", "prime")
	if cmd != nil || !errors.Is(err, failure) || !strings.Contains(err.Error(), "caller: resolve executable:") {
		t.Fatalf("want wrapped executable error and no command, got %v, %v", cmd, err)
	}
}

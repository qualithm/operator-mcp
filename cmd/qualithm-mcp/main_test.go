package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	origVersion, origCommit := version, commit
	t.Cleanup(func() { version, commit = origVersion, origCommit })
	version, commit = "1.2.3", "abc1234"

	var out bytes.Buffer
	if err := run([]string{"--version"}, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "1.2.3 (abc1234)" {
		t.Errorf("--version printed %q, want %q", got, "1.2.3 (abc1234)")
	}
}

func TestRunHelpIncludesVersion(t *testing.T) {
	orig := version
	t.Cleanup(func() { version = orig })
	version = "1.2.3"

	var out bytes.Buffer
	if err := run([]string{"--help"}, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(out.String(), "qualithm-mcp 1.2.3") {
		t.Errorf("--help output %q does not name the version", out.String())
	}
}

func TestRunRejectsUnknownFlag(t *testing.T) {
	if err := run([]string{"--nope"}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected an error for an unknown flag")
	}
}

func TestRunRequiresToken(t *testing.T) {
	t.Setenv("QUALITHM_API_TOKEN", "")
	if err := run(nil, &bytes.Buffer{}); err == nil {
		t.Fatal("expected an error without an API token")
	}
}

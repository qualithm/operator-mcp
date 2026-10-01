package main

import "testing"

func TestVersionString(t *testing.T) {
	origVersion, origCommit := version, commit
	t.Cleanup(func() { version, commit = origVersion, origCommit })

	version, commit = "1.2.3", ""
	if got := versionString(); got != "1.2.3" {
		t.Errorf("versionString() = %q, want %q", got, "1.2.3")
	}

	version, commit = "1.2.3", "abc1234"
	if got := versionString(); got != "1.2.3 (abc1234)" {
		t.Errorf("versionString() = %q, want %q", got, "1.2.3 (abc1234)")
	}
}

func TestResolvedVersionPrefersStamp(t *testing.T) {
	orig := version
	t.Cleanup(func() { version = orig })

	version = "4.5.6"
	if got := resolvedVersion(); got != "4.5.6" {
		t.Errorf("resolvedVersion() = %q, want %q", got, "4.5.6")
	}
}

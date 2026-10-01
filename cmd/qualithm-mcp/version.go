package main

import (
	"runtime/debug"
	"strings"
)

// version and commit identify the build. Release builds stamp them via
//
//	-ldflags "-X main.version=1.2.3 -X main.commit=abc1234"
//
// (goreleaser, the Makefile, and dx's release workflow all do), and the
// release workflow fails if `--version` still reports "dev".
var (
	version = "dev"
	commit  = ""
)

// resolvedVersion returns the stamped version, falling back to the module
// version Go records in the binary. That covers `go install
// github.com/qualithm/<repo>/cmd/<binary>@v1.2.3`, which never sees ldflags.
func resolvedVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return strings.TrimPrefix(v, "v")
		}
	}
	return version
}

// versionString is what `--version` prints: the version, plus the commit when
// one was stamped.
func versionString() string {
	if commit == "" {
		return resolvedVersion()
	}
	return resolvedVersion() + " (" + commit + ")"
}

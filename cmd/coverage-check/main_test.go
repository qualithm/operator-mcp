package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	platformRoute = `package placeholder

import http from "../lib/http"

http.get(
  "/devices",
  async (c) => c.json({ data: [] }),
)

http.post(
  // Issues the resource immediately.
  "/devices/:deviceId/park",
  async (c) => c.json({ data: null }),
)
`
	serverTools = `package server

import "github.com/modelcontextprotocol/go-sdk/mcp"

func (s *Server) registerDevices(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_devices",
		Description: "List devices.",
	}, s.listDevices)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "park_device",
		Description: "Park a device.",
	}, s.parkDevice)
}
`
)

type fixture struct {
	platform string
	server   string
	ledger   string
}

// writeFixture builds a minimal platform checkout, tool registry, and ledger in
// t.TempDir(). routes maps "METHOD /path" to its ledger classification fields.
func writeFixture(t *testing.T, routes map[string]ledgerEntry) fixture {
	t.Helper()
	root := t.TempDir()

	platform := filepath.Join(root, "platform")
	if err := os.MkdirAll(filepath.Join(platform, "src", "management"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(platform, "src", "management", "routes.ts"), []byte(platformRoute), 0o644); err != nil {
		t.Fatal(err)
	}

	server := filepath.Join(root, "server")
	if err := os.MkdirAll(server, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(server, "devices.go"), []byte(serverTools), 0o644); err != nil {
		t.Fatal(err)
	}

	lg := ledger{}
	for key, e := range routes {
		parts := strings.SplitN(key, " ", 2)
		e.Method, e.Path = parts[0], parts[1]
		lg.Routes = append(lg.Routes, e)
	}
	raw, err := json.Marshal(lg)
	if err != nil {
		t.Fatal(err)
	}
	ledgerPath := filepath.Join(root, "coverage.json")
	if err := os.WriteFile(ledgerPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return fixture{platform: platform, server: server, ledger: ledgerPath}
}

func TestRunHappyPath(t *testing.T) {
	t.Parallel()
	fx := writeFixture(t, map[string]ledgerEntry{
		"GET /devices":                 {Tool: "list_devices"},
		"POST /devices/:deviceId/park": {Tool: "park_device"},
	})
	if err := run(fx.platform, fx.ledger, fx.server); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestRunUnclassifiedRoute(t *testing.T) {
	t.Parallel()
	fx := writeFixture(t, map[string]ledgerEntry{
		"GET /devices": {Tool: "list_devices"},
	})
	err := run(fx.platform, fx.ledger, fx.server)
	if err == nil || !strings.Contains(err.Error(), "drift") {
		t.Fatalf("expected drift error, got %v", err)
	}
}

func TestRunStaleEntry(t *testing.T) {
	t.Parallel()
	fx := writeFixture(t, map[string]ledgerEntry{
		"GET /devices":                 {Tool: "list_devices"},
		"POST /devices/:deviceId/park": {Tool: "park_device"},
		"DELETE /gone":                 {NoTool: "app-ui"},
	})
	if err := run(fx.platform, fx.ledger, fx.server); err == nil {
		t.Fatal("expected stale-entry drift, got nil")
	}
}

func TestRunUnregisteredTool(t *testing.T) {
	t.Parallel()
	fx := writeFixture(t, map[string]ledgerEntry{
		"GET /devices":                 {Tool: "list_devices"},
		"POST /devices/:deviceId/park": {Tool: "nonexistent_tool"},
	})
	if err := run(fx.platform, fx.ledger, fx.server); err == nil {
		t.Fatal("expected unregistered-tool drift, got nil")
	}
}

func TestRunUnclaimedTool(t *testing.T) {
	t.Parallel()
	// The fixture registers park_device, but no ledger entry claims it.
	fx := writeFixture(t, map[string]ledgerEntry{
		"GET /devices":                 {Tool: "list_devices"},
		"POST /devices/:deviceId/park": {NoTool: "app-ui"},
	})
	if err := run(fx.platform, fx.ledger, fx.server); err == nil {
		t.Fatal("expected unclaimed-tool drift, got nil")
	}
}

func TestRunSchemaViolation(t *testing.T) {
	t.Parallel()
	fx := writeFixture(t, map[string]ledgerEntry{
		"GET /devices":                 {Tool: "list_devices", NoTool: "app-ui"},
		"POST /devices/:deviceId/park": {Tool: "park_device"},
	})
	if err := run(fx.platform, fx.ledger, fx.server); err == nil {
		t.Fatal("expected schema drift, got nil")
	}
}

func TestRunMissingPlatform(t *testing.T) {
	t.Parallel()
	fx := writeFixture(t, map[string]ledgerEntry{
		"GET /devices":                 {Tool: "list_devices"},
		"POST /devices/:deviceId/park": {Tool: "park_device"},
	})
	if err := run(filepath.Join(fx.platform, "nope"), fx.ledger, fx.server); err == nil {
		t.Fatal("expected missing-platform error, got nil")
	}
}

func TestLoadLedgerRejectsEmpty(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "empty.json")
	if err := os.WriteFile(p, []byte(`{"routes": []}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadLedger(p); err == nil {
		t.Fatal("expected empty-ledger error, got nil")
	}
}

func TestScanToolsRejectsEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := scanTools(dir); err == nil {
		t.Fatal("expected no-tools error, got nil")
	}
}

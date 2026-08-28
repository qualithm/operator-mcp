package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestDecodeToolResult(t *testing.T) {
	ok := &mcp.CallToolResult{StructuredContent: map[string]any{
		"ok":   true,
		"data": map[string]any{"id": "dev_1"},
	}}
	var out struct {
		ID string `json:"id"`
	}
	if err := decodeToolResult("get_device", ok, &out); err != nil {
		t.Fatal(err)
	}
	if out.ID != "dev_1" {
		t.Fatalf("out = %+v", out)
	}

	// Failure envelopes surface the stable code.
	failed := &mcp.CallToolResult{
		IsError:           true,
		StructuredContent: map[string]any{"ok": false, "code": "not_found", "message": "no such device"},
	}
	err := decodeToolResult("get_device", failed, nil)
	if err == nil || !strings.Contains(err.Error(), "not_found") || !strings.Contains(err.Error(), "no such device") {
		t.Fatalf("err = %v", err)
	}

	// A failure without an envelope still errors.
	bare := &mcp.CallToolResult{IsError: true}
	if err := decodeToolResult("get_device", bare, nil); err == nil {
		t.Fatal("want error for codeless failure")
	}

	// Success with a nil out and no data is fine.
	plain := &mcp.CallToolResult{StructuredContent: map[string]any{"ok": true}}
	if err := decodeToolResult("delete_device", plain, nil); err != nil {
		t.Fatal(err)
	}
}

func TestClaim(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/provision/claim" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"deviceId":"dev_1","secret":"s3cr3t"}}`))
	}))
	defer srv.Close()

	p := &httpProvisioner{baseURL: srv.URL + "/", http: srv.Client()}
	cred, err := p.claim(context.Background(), "qmc_code", "agent-eval")
	if err != nil {
		t.Fatal(err)
	}
	if cred.DeviceID != "dev_1" || cred.Secret != "s3cr3t" {
		t.Fatalf("cred = %+v", cred)
	}
}

func TestClaimFailures(t *testing.T) {
	// Non-2xx carries the API message.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Invalid or expired claim code"}`))
	}))
	defer srv.Close()
	p := &httpProvisioner{baseURL: srv.URL, http: srv.Client()}
	_, err := p.claim(context.Background(), "qmc_bad", "agent-eval")
	if err == nil || !strings.Contains(err.Error(), "Invalid or expired claim code") {
		t.Fatalf("err = %v", err)
	}

	// A non-JSON body still errors.
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv2.Close()
	p = &httpProvisioner{baseURL: srv2.URL, http: srv2.Client()}
	if _, err := p.claim(context.Background(), "qmc_bad", "agent-eval"); err == nil {
		t.Fatal("want decode error")
	}

	// An unreachable endpoint errors instead of hanging.
	p = &httpProvisioner{baseURL: "http://127.0.0.1:1", http: &http.Client{Timeout: time.Second}}
	if _, err := p.claim(context.Background(), "qmc_bad", "agent-eval"); err == nil {
		t.Fatal("want transport error")
	}
}

func TestPublishTelemetryConnectFailure(t *testing.T) {
	// Nothing listens on the port: the connect fails fast and is reported.
	p := &mqttPublisher{host: "127.0.0.1", port: 1}
	err := p.publishTelemetry(context.Background(), claimResult{DeviceID: "dev_1", Secret: "s"}, "agent_eval", 1, 1)
	if err == nil {
		t.Fatal("want connect error")
	}
}

func TestEnvOr(t *testing.T) {
	t.Setenv("AGENT_EVAL_TEST_KEY", "set")
	if got := envOr("AGENT_EVAL_TEST_KEY", "fallback"); got != "set" {
		t.Fatalf("got %q", got)
	}
	if got := envOr("AGENT_EVAL_TEST_UNSET", "fallback"); got != "fallback" {
		t.Fatalf("got %q", got)
	}
}

func TestRunMainRequiresConfig(t *testing.T) {
	// With no server binary or token the eval refuses to run and exits 2 so the
	// workflow treats it as unconfigured rather than failed.
	os.Args = []string{"agent-eval"}
	if code := runMain(); code != exitCodeConfig {
		t.Fatalf("code = %d", code)
	}

	// Flag parse errors exit the same way.
	os.Args = []string{"agent-eval", "-nope"}
	if code := runMain(); code != exitCodeConfig {
		t.Fatalf("code = %d", code)
	}
}

// TestConnectMCPServer exercises the subprocess transport for real: it builds
// qualithm-mcp from this module, spawns it, and round-trips a tool call.
func TestConnectMCPServer(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not on PATH")
	}
	bin := filepath.Join(t.TempDir(), "qualithm-mcp")
	build := exec.Command(goBin, "build", "-o", bin, "../qualithm-mcp")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build server: %v\n%s", err, out)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	tools, cleanup, err := connectMCPServer(ctx, bin, []string{"QUALITHM_API_TOKEN=qmt_selector.verifier"})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	// The server is up: a tool call with no reachable API fails cleanly through
	// the envelope, proving the transport both ways.
	err = tools.callTool(ctx, "list_spaces", map[string]any{}, nil)
	if err == nil {
		t.Fatal("want an API-unreachable error")
	}
}

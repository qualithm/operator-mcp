package main

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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
		if r.URL.Path != "/v1/provision/claim" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"deviceId":"dev_1","secret":"s3cr3t","broker":{"host":"gw.example","port":8883}}}`))
	}))
	defer srv.Close()

	p := &httpProvisioner{baseURL: srv.URL + "/", http: srv.Client()}
	cred, err := p.claim(context.Background(), "qmc_code", "agent-eval")
	if err != nil {
		t.Fatal(err)
	}
	if cred.DeviceID != "dev_1" || cred.Secret != "s3cr3t" || cred.Broker != (endpoint{Host: "gw.example", Port: 8883}) {
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
	p := &mqttPublisher{}
	err := p.publishTelemetry(context.Background(), claimResult{DeviceID: "dev_1", Secret: "s"}, endpoint{Host: "127.0.0.1", Port: 1}, "agent_eval", 1, 1)
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
	err = tools.callTool(ctx, "list_scopes", map[string]any{}, nil)
	if err == nil {
		t.Fatal("want an API-unreachable error")
	}
}

func TestPublishReadingDeclaresCapabilitiesFirst(t *testing.T) {
	var topics []string
	var payloads []map[string]any
	err := publishReading(func(topic string, payload []byte) error {
		topics = append(topics, topic)
		var m map[string]any
		if err := json.Unmarshal(payload, &m); err != nil {
			t.Fatalf("%s payload: %v", topic, err)
		}
		payloads = append(payloads, m)
		return nil
	}, "agent_eval", 42, 1767225600000)
	if err != nil {
		t.Fatal(err)
	}
	if len(topics) != 2 || topics[0] != "capabilities" || topics[1] != "telemetry" {
		t.Fatalf("topics = %v", topics)
	}
	wantManifest := map[string]any{"capabilities": []any{map[string]any{
		"key": "agent_eval", "type": "sensor", "config": map[string]any{"metric": "agent_eval"},
	}}}
	if !reflect.DeepEqual(payloads[0], wantManifest) {
		t.Fatalf("manifest = %v", payloads[0])
	}
	wantReading := map[string]any{"ts": float64(1767225600000), "metrics": map[string]any{"agent_eval": float64(42)}}
	if !reflect.DeepEqual(payloads[1], wantReading) {
		t.Fatalf("reading = %v", payloads[1])
	}
}

func TestPublishReadingNamesTheFailedPublish(t *testing.T) {
	refuse := func(on string) qos1Publish {
		return func(topic string, _ []byte) error {
			if topic == on {
				return errors.New("mqtt publish timed out")
			}
			return nil
		}
	}
	err := publishReading(refuse("capabilities"), "agent_eval", 1, 1)
	if err == nil || err.Error() != "declare capabilities: mqtt publish timed out" {
		t.Fatalf("err = %v", err)
	}
	err = publishReading(refuse("telemetry"), "agent_eval", 1, 1)
	if err == nil || err.Error() != "publish telemetry: mqtt publish timed out" {
		t.Fatalf("err = %v", err)
	}
	// A value JSON can't carry fails before the reading is published.
	err = publishReading(refuse("telemetry"), "agent_eval", math.NaN(), 1)
	if err == nil || !strings.HasPrefix(err.Error(), "encode telemetry:") {
		t.Fatalf("err = %v", err)
	}
}

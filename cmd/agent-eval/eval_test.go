package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// toolFunc answers one MCP tool call.
type toolFunc func(args map[string]any) (any, error)

// respond answers a tool call with a fixed payload.
func respond(v any) toolFunc {
	return func(map[string]any) (any, error) { return v, nil }
}

// failWith answers a tool call with a fixed error.
func failWith(err error) toolFunc {
	return func(map[string]any) (any, error) { return nil, err }
}

// fakeTools answers MCP tool calls from a per-tool map.
type fakeTools struct {
	responses map[string]toolFunc
	calls     []string
}

func (f *fakeTools) callTool(_ context.Context, name string, args map[string]any, out any) error {
	f.calls = append(f.calls, name)
	fn, ok := f.responses[name]
	if !ok {
		return fmt.Errorf("unexpected tool call %s", name)
	}
	data, err := fn(args)
	if err != nil {
		return err
	}
	if out != nil && data != nil {
		raw, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, out); err != nil {
			return err
		}
	}
	return nil
}

type fakeProvisioner struct {
	result claimResult
	err    error
}

func (f fakeProvisioner) claim(context.Context, string, string) (claimResult, error) {
	return f.result, f.err
}

type fakePublisher struct {
	err error
}

func (f fakePublisher) publishTelemetry(context.Context, claimResult, string, float64, int64) error {
	return f.err
}

// recordingPublisher captures the value the eval publishes so the fake
// telemetry read can answer with it, as the real store would.
type recordingPublisher struct {
	sink *float64
}

func (p *recordingPublisher) publishTelemetry(_ context.Context, _ claimResult, _ string, value float64, _ int64) error {
	*p.sink = value
	return nil
}

var spacesWithOne = respond(map[string]any{"items": []map[string]any{{"id": "spc_1"}}})

var enrollmentOK = respond(map[string]any{"enrollment": map[string]any{"id": "enr_1"}, "code": "qmc_code"})

var deleteOK = respond(nil)

func TestRunHappyPath(t *testing.T) {
	var publishedValue float64
	tools := &fakeTools{responses: map[string]toolFunc{
		"list_spaces":       spacesWithOne,
		"create_enrollment": enrollmentOK,
		// Read-back answers with the value the publisher recorded, as the real
		// telemetry store would once the reading lands.
		"get_telemetry": func(map[string]any) (any, error) {
			return []map[string]any{{"ts": 1, "value": publishedValue}}, nil
		},
		"delete_device": deleteOK,
		"revoke_enrollment": func(map[string]any) (any, error) {
			return nil, errors.New("should not be called on the happy path")
		},
	}}
	report := Run(context.Background(), tools,
		fakeProvisioner{result: claimResult{DeviceID: "dev_1", Secret: "s"}},
		&recordingPublisher{sink: &publishedValue},
		Config{Zone: "sandbox"},
	)
	if !report.OK {
		t.Fatalf("report = %+v", report)
	}
	wantSteps := []string{"list_spaces", "create_enrollment", "claim", "connect_publish", "verify_telemetry"}
	if len(report.Steps) != len(wantSteps) {
		t.Fatalf("steps = %v", report.Steps)
	}
	for i, name := range wantSteps {
		if report.Steps[i].Name != name || !report.Steps[i].OK {
			t.Errorf("step %d = %+v", i, report.Steps[i])
		}
	}
	// The broker-host gap is always friction on the happy path.
	if len(report.Friction) != 1 || report.Friction[0].Step != "connect" {
		t.Fatalf("friction = %+v", report.Friction)
	}
}

func TestRunCreatesSpaceWhenNone(t *testing.T) {
	var publishedValue float64
	tools := &fakeTools{responses: map[string]toolFunc{
		"list_spaces": respond(map[string]any{"items": []any{}}),
		"create_space": func(args map[string]any) (any, error) {
			if args["zone"] != "sandbox" {
				return nil, fmt.Errorf("wrong zone %v", args["zone"])
			}
			return map[string]any{"id": "spc_new"}, nil
		},
		"create_enrollment": func(args map[string]any) (any, error) {
			if args["spaceId"] != "spc_new" {
				return nil, fmt.Errorf("wrong space %v", args["spaceId"])
			}
			return map[string]any{"enrollment": map[string]any{"id": "enr_1"}, "code": "qmc_code"}, nil
		},
		"get_telemetry": func(map[string]any) (any, error) {
			return []map[string]any{{"ts": 1, "value": publishedValue}}, nil
		},
		"delete_device": deleteOK,
	}}
	report := Run(context.Background(), tools,
		fakeProvisioner{result: claimResult{DeviceID: "dev_1", Secret: "s"}},
		&recordingPublisher{sink: &publishedValue},
		Config{Zone: "sandbox"},
	)
	if !report.OK {
		t.Fatalf("report = %+v", report)
	}
	// Both the zone and the broker host had to come from configuration.
	steps := make([]string, 0, len(report.Friction))
	for _, f := range report.Friction {
		steps = append(steps, f.Step)
	}
	if fmt.Sprint(steps) != "[create_space connect]" {
		t.Fatalf("friction steps = %v", steps)
	}
}

func TestRunClaimFailureStopsAndRevokes(t *testing.T) {
	tools := &fakeTools{responses: map[string]toolFunc{
		"list_spaces":       spacesWithOne,
		"create_enrollment": enrollmentOK,
		"revoke_enrollment": respond(nil),
	}}
	report := Run(context.Background(), tools,
		fakeProvisioner{err: errors.New("Invalid or expired claim code")},
		fakePublisher{},
		Config{Zone: "sandbox"},
	)
	if report.OK {
		t.Fatal("want failure")
	}
	if got := report.Steps[len(report.Steps)-1].Name; got != "claim" {
		t.Fatalf("last step = %q", got)
	}
	// The unused enrollment is revoked; no device exists to delete.
	if tools.calls[len(tools.calls)-1] != "revoke_enrollment" {
		t.Fatalf("calls = %v", tools.calls)
	}
}

func TestRunPublishFailureDeletesDevice(t *testing.T) {
	tools := &fakeTools{responses: map[string]toolFunc{
		"list_spaces":       spacesWithOne,
		"create_enrollment": enrollmentOK,
		"delete_device":     deleteOK,
	}}
	report := Run(context.Background(), tools,
		fakeProvisioner{result: claimResult{DeviceID: "dev_1", Secret: "s"}},
		fakePublisher{err: errors.New("mqtt connect: timeout")},
		Config{Zone: "sandbox"},
	)
	if report.OK {
		t.Fatal("want failure")
	}
	if got := report.Steps[len(report.Steps)-1].Name; got != "connect_publish" {
		t.Fatalf("last step = %q", got)
	}
	if tools.calls[len(tools.calls)-1] != "delete_device" {
		t.Fatalf("calls = %v", tools.calls)
	}
}

func TestRunVerifyTimesOut(t *testing.T) {
	tools := &fakeTools{responses: map[string]toolFunc{
		"list_spaces":       spacesWithOne,
		"create_enrollment": enrollmentOK,
		// The point never carries the eval's value: read-back never confirms.
		"get_telemetry": respond([]map[string]any{{"ts": 1, "value": -1}}),
		"delete_device": deleteOK,
	}}
	report := Run(context.Background(), tools,
		fakeProvisioner{result: claimResult{DeviceID: "dev_1", Secret: "s"}},
		fakePublisher{},
		Config{Zone: "sandbox", VerifyTimeout: 30 * time.Millisecond, PollInterval: 5 * time.Millisecond},
	)
	if report.OK {
		t.Fatal("want failure")
	}
	last := report.Steps[len(report.Steps)-1]
	if last.Name != "verify_telemetry" || !strings.Contains(last.Detail, "no telemetry point") {
		t.Fatalf("last step = %+v", last)
	}
}

func TestRunToolFailureStopsEarly(t *testing.T) {
	tools := &fakeTools{responses: map[string]toolFunc{
		"list_spaces": failWith(errors.New("list_spaces: auth: bad token")),
	}}
	report := Run(context.Background(), tools,
		fakeProvisioner{result: claimResult{DeviceID: "dev_1", Secret: "s"}},
		fakePublisher{},
		Config{Zone: "sandbox"},
	)
	if report.OK {
		t.Fatal("want failure")
	}
	if len(report.Steps) != 1 || report.Steps[0].Name != "list_spaces" {
		t.Fatalf("steps = %+v", report.Steps)
	}
}

func TestReportRender(t *testing.T) {
	report := Report{
		OK:        false,
		StartedAt: time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC),
		Steps: []Step{
			{Name: "claim", OK: true, DurationMS: 10},
			{Name: "connect_publish", OK: false, DurationMS: 20, Detail: "mqtt connect: timeout"},
		},
		Friction: []Friction{{Step: "connect", Missing: "broker host"}},
	}
	out := report.Render()
	for _, want := range []string{"FAIL", "[ok] claim", "[FAIL] connect_publish", "connect: broker host"} {
		if !strings.Contains(out, want) {
			t.Errorf("render missing %q:\n%s", want, out)
		}
	}
}

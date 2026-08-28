// Command agent-eval is the agent-native end-to-end eval: it drives the
// documented agent path against a live environment — provision a device
// through the MCP tools, claim it, connect, publish telemetry, and read it
// back — with no human input, and records every step where the surfaces did
// not provide what the harness needed (the friction log). It exits non-zero
// when any step fails; the weekly workflow turns that into a deduped issue.
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Step is one eval step's outcome.
type Step struct {
	Name       string `json:"name"`
	OK         bool   `json:"ok"`
	DurationMS int64  `json:"durationMs"`
	Detail     string `json:"detail,omitempty"`
}

// Friction records one place the harness had to supply information the
// product surfaces did not provide.
type Friction struct {
	Step    string `json:"step"`
	Missing string `json:"missing"`
}

// Report is the eval's structured result, written as JSON and rendered into
// the workflow summary or failure issue.
type Report struct {
	OK        bool       `json:"ok"`
	StartedAt time.Time  `json:"startedAt"`
	Steps     []Step     `json:"steps"`
	Friction  []Friction `json:"friction"`
}

// toolCaller invokes one MCP tool and decodes the uniform result envelope's
// data field into out (nil for none). It returns an error carrying the
// envelope's stable code when the tool reports failure.
type toolCaller interface {
	callTool(ctx context.Context, name string, args map[string]any, out any) error
}

// claimResult is the provisioning API's claim response: the device identity
// and its first credential secret.
type claimResult struct {
	DeviceID string `json:"deviceId"`
	Secret   string `json:"secret"`
}

// provisioner exchanges a claim code for a device credential.
type provisioner interface {
	claim(ctx context.Context, code, name string) (claimResult, error)
}

// publisher connects a device and publishes one telemetry reading.
type publisher interface {
	publishTelemetry(ctx context.Context, cred claimResult, metric string, value float64, ts int64) error
}

// Config carries everything the eval needs that the product surfaces do not
// provide. Each field the harness had to be given is a friction entry.
type Config struct {
	// Zone creates the eval space in this device zone when the team has none.
	Zone string
	// VerifyTimeout bounds the telemetry read-back poll.
	VerifyTimeout time.Duration
	// PollInterval is the delay between read-back attempts.
	PollInterval time.Duration
}

type evalRunner struct {
	tools     toolCaller
	provision provisioner
	publish   publisher
	cfg       Config
	report    *Report
}

// Run executes the full claim → connect → telemetry round-trip. The report is
// always returned, including on failure.
func Run(ctx context.Context, tools toolCaller, provision provisioner, publish publisher, cfg Config) Report {
	r := &evalRunner{
		tools:     tools,
		provision: provision,
		publish:   publish,
		cfg:       cfg,
		report:    &Report{StartedAt: time.Now().UTC()},
	}
	if cfg.VerifyTimeout <= 0 {
		cfg.VerifyTimeout = 60 * time.Second
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 3 * time.Second
	}
	r.cfg = cfg
	r.run(ctx)
	r.report.OK = true
	for _, s := range r.report.Steps {
		if !s.OK {
			r.report.OK = false
			break
		}
	}
	return *r.report
}

// step records one step's outcome and fails the eval on error.
func (r *evalRunner) step(name string, fn func() (string, error)) error {
	start := time.Now()
	detail, err := fn()
	s := Step{Name: name, OK: err == nil, DurationMS: time.Since(start).Milliseconds(), Detail: detail}
	r.report.Steps = append(r.report.Steps, s)
	if err != nil {
		s.Detail = err.Error()
		r.report.Steps[len(r.report.Steps)-1] = s
	}
	return err
}

// friction records a surface gap the harness had to compensate for.
func (r *evalRunner) friction(step, missing string) {
	r.report.Friction = append(r.report.Friction, Friction{Step: step, Missing: missing})
}

func (r *evalRunner) run(ctx context.Context) {
	// The eval metric's value doubles as the run's fingerprint: read-back
	// matches on it, so a stale point from an earlier run never satisfies the
	// verify step.
	metric := "agent_eval"
	value := float64(time.Now().UnixMilli()%1000000) / 100
	nowMS := time.Now().UnixMilli()

	var spaceID string
	needsSpace := false
	if err := r.step("list_spaces", func() (string, error) {
		var page struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
		}
		if err := r.tools.callTool(ctx, "list_spaces", map[string]any{}, &page); err != nil {
			return "", err
		}
		if len(page.Items) == 0 {
			// An empty team is not a failure; the eval creates the space.
			needsSpace = true
			return "no spaces yet", nil
		}
		spaceID = page.Items[0].ID
		return fmt.Sprintf("space %s", spaceID), nil
	}); err != nil {
		return
	}
	if needsSpace {
		// The tool surface never enumerates the zones create_space needs; the
		// zone comes from eval configuration.
		r.friction("create_space", "no tool lists the device zones; the eval's zone comes from configuration")
		if err := r.step("create_space", func() (string, error) {
			var space struct {
				ID string `json:"id"`
			}
			if err := r.tools.callTool(ctx, "create_space", map[string]any{"zone": r.cfg.Zone}, &space); err != nil {
				return "", err
			}
			spaceID = space.ID
			return fmt.Sprintf("space %s", spaceID), nil
		}); err != nil {
			return
		}
	}

	var code string
	var enrollmentID string
	if err := r.step("create_enrollment", func() (string, error) {
		var out struct {
			Enrollment struct {
				ID string `json:"id"`
			} `json:"enrollment"`
			Code string `json:"code"`
		}
		if err := r.tools.callTool(ctx, "create_enrollment", map[string]any{
			"spaceId":          spaceID,
			"label":            "agent-eval",
			"expiresInMinutes": 15,
		}, &out); err != nil {
			return "", err
		}
		if out.Code == "" {
			return "", errors.New("create_enrollment returned no claim code")
		}
		code = out.Code
		enrollmentID = out.Enrollment.ID
		return fmt.Sprintf("enrollment %s", enrollmentID), nil
	}); err != nil {
		return
	}

	var cred claimResult
	if err := r.step("claim", func() (string, error) {
		var err error
		cred, err = r.provision.claim(ctx, code, "agent-eval")
		if err != nil {
			return "", err
		}
		if cred.DeviceID == "" || cred.Secret == "" {
			return "", errors.New("claim response is missing the device id or credential secret")
		}
		return fmt.Sprintf("device %s", cred.DeviceID), nil
	}); err != nil {
		r.revokeEnrollment(ctx, enrollmentID)
		return
	}

	// The claim response identifies the device but not where it connects.
	r.friction("connect", "claim response omits the broker host; the eval's broker comes from configuration")

	if err := r.step("connect_publish", func() (string, error) {
		if err := r.publish.publishTelemetry(ctx, cred, metric, value, nowMS); err != nil {
			return "", err
		}
		return fmt.Sprintf("published %s=%v", metric, value), nil
	}); err != nil {
		r.deleteDevice(ctx, cred.DeviceID)
		return
	}

	if err := r.step("verify_telemetry", func() (string, error) {
		deadline := time.Now().Add(r.cfg.VerifyTimeout)
		for {
			var out []struct {
				TS    int64   `json:"ts"`
				Value float64 `json:"value"`
			}
			err := r.tools.callTool(ctx, "get_telemetry", map[string]any{
				"deviceId": cred.DeviceID,
				"metric":   metric,
				"from":     nowMS - 60_000,
				"to":       time.Now().UnixMilli() + 60_000,
				"bucket":   60,
			}, &out)
			if err == nil {
				for _, p := range out {
					if p.Value == value {
						return fmt.Sprintf("read back %s=%v", metric, value), nil
					}
				}
			}
			if time.Now().After(deadline) {
				if err != nil {
					return "", fmt.Errorf("get_telemetry kept failing: %w", err)
				}
				return "", fmt.Errorf("no telemetry point with %s=%v within %s", metric, value, r.cfg.VerifyTimeout)
			}
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(r.cfg.PollInterval):
			}
		}
	}); err != nil {
		r.deleteDevice(ctx, cred.DeviceID)
		return
	}

	r.deleteDevice(ctx, cred.DeviceID)
}

// deleteDevice removes the eval device. Cleanup is best-effort: a failure is
// friction, not an eval failure — but it is recorded, because an agent that
// cannot clean up after itself leaves the team dirty.
func (r *evalRunner) deleteDevice(ctx context.Context, deviceID string) {
	if err := r.tools.callTool(ctx, "delete_device", map[string]any{"deviceId": deviceID}, nil); err != nil {
		r.friction("cleanup", fmt.Sprintf("delete_device failed: %v", err))
	}
}

// revokeEnrollment withdraws the claim code when the eval never claimed it.
func (r *evalRunner) revokeEnrollment(ctx context.Context, enrollmentID string) {
	if err := r.tools.callTool(ctx, "revoke_enrollment", map[string]any{"enrollmentId": enrollmentID}, nil); err != nil {
		r.friction("cleanup", fmt.Sprintf("revoke_enrollment failed: %v", err))
	}
}

// Render formats the report as the human-readable summary used for the
// workflow log and the failure issue body.
func (r Report) Render() string {
	var b []byte
	status := "PASS"
	if !r.OK {
		status = "FAIL"
	}
	b = append(b, fmt.Sprintf("agent-eval %s (%s)\n\nSteps:\n", status, r.StartedAt.Format(time.RFC3339))...)
	for _, s := range r.Steps {
		mark := "ok"
		if !s.OK {
			mark = "FAIL"
		}
		line := fmt.Sprintf("- [%s] %s (%dms)", mark, s.Name, s.DurationMS)
		if s.Detail != "" {
			line += ": " + s.Detail
		}
		b = append(b, line+"\n"...)
	}
	b = append(b, "\nFriction log (steps the agent had to supply information the surfaces did not provide):\n"...)
	if len(r.Friction) == 0 {
		b = append(b, "- none\n"...)
	}
	for _, f := range r.Friction {
		b = append(b, fmt.Sprintf("- %s: %s\n", f.Step, f.Missing)...)
	}
	return string(b)
}

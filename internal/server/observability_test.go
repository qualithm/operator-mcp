package server

import (
	"strings"
	"testing"
)

func TestObservabilityTools(t *testing.T) {
	rec := &record{}
	s := testServer(t, 200, envelope(`{"items":[{"ts":1000,"value":2.5}]}`), rec)
	if _, out, err := s.getTelemetry(ctx(), nil, GetTelemetryInput{DeviceID: "dev_1", Metric: "temp", From: "2026-10-01T00:00:00Z", To: "2026-10-01T01:00:00Z", Bucket: 60, Agg: "max", FillLOCF: true}); err != nil || !out.OK {
		t.Fatalf("getTelemetry: %v %+v", err, out)
	}
	if _, out, _ := s.getTelemetry(ctx(), nil, GetTelemetryInput{DeviceID: "dev_1", Metric: "temp", From: "yesterday", To: "2026-10-01T01:00:00Z"}); out.OK || !strings.Contains(out.Message, "from must be an RFC 3339 time") {
		t.Fatalf("bad from: %+v", out)
	}
	if rec.path != "/v1/teams/team_1/devices/dev_1/telemetry" {
		t.Fatalf("path %q", rec.path)
	}
	for _, want := range []string{"metric=temp", "from=2026-10-01T00%3A00%3A00Z", "bucket=60", "agg=max", "fill=locf"} {
		if !strings.Contains(rec.query, want) {
			t.Fatalf("query %q missing %q", rec.query, want)
		}
	}

	// The SSE body is served as-is; the read returns after the first event.
	sse := "event: device.connected\ndata: {\"deviceId\":\"dev_1\"}\n\n"
	s = testServer(t, 200, sse, rec)
	_, out, err := s.streamEvents(ctx(), nil, StreamEventsInput{Limit: 1, TimeoutMS: 1000})
	if err != nil || !out.OK {
		t.Fatalf("streamEvents: %v %+v", err, out)
	}

	s = testServer(t, 200, envelope(`{"deviceTotal":3,"deviceMax":10,"spaceTotal":2}`), rec)
	if _, out, err := s.getUsage(ctx(), nil, GetUsageInput{}); err != nil || !out.OK {
		t.Fatalf("getUsage: %v %+v", err, out)
	}
	if rec.path != "/v1/teams/team_1/usage" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"current":1,"items":[{"id":"aud_1","action":"device.create"}],"last":1}`), rec)
	if _, out, err := s.getAuditLog(ctx(), nil, GetAuditLogInput{}); err != nil || !out.OK {
		t.Fatalf("getAuditLog: %v %+v", err, out)
	}
	if rec.path != "/v1/teams/team_1/audit-events" {
		t.Fatalf("path %q", rec.path)
	}
}

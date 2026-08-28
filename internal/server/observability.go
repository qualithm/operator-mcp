package server

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// GetTelemetryInput reads bucketed telemetry points for a device metric.
// From and To are epoch milliseconds.
type GetTelemetryInput struct {
	DeviceID string `json:"deviceId" jsonschema:"id of the device to read telemetry for"`
	Metric   string `json:"metric" jsonschema:"metric name to read"`
	From     int64  `json:"from" jsonschema:"range start, epoch milliseconds"`
	To       int64  `json:"to" jsonschema:"range end, epoch milliseconds"`
	Bucket   int    `json:"bucket,omitempty" jsonschema:"bucket width in seconds; 0 uses the server default"`
	Agg      string `json:"agg,omitempty" jsonschema:"bucket aggregate: avg, min, max, sum, count; default avg"`
	FillLOCF bool   `json:"fillLocf,omitempty" jsonschema:"carry the last observation forward across empty buckets"`
}

// StreamEventsInput reads a bounded batch from the team's live event stream.
// The stream never terminates on its own, so the read always ends: after
// limit events or after timeoutMs, whichever comes first.
type StreamEventsInput struct {
	Limit     int `json:"limit,omitempty" jsonschema:"maximum number of events to read; 0 reads a single event"`
	TimeoutMS int `json:"timeoutMs,omitempty" jsonschema:"give up after this many milliseconds; 0 uses a 5s default"`
}

// GetUsageInput reads the team's current usage against its plan limits.
type GetUsageInput struct{}

// GetAuditLogInput lists audit-trail events for the token's team.
type GetAuditLogInput struct {
	Page  int `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit int `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

func (s *Server) registerObservability(srv *mcp.Server) {
	addTool(s, srv, &mcp.Tool{
		Name:        "get_telemetry",
		Description: "Read bucketed telemetry points for a device metric over a time range.",
	}, s.getTelemetry, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "stream_events",
		Description: "Read a bounded batch from the team's live event stream; returns after limit events or the timeout.",
	}, s.streamEvents, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "get_usage",
		Description: "Read the team's current usage against its plan limits.",
	}, s.getUsage, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "get_audit_log",
		Description: "List audit-trail events for the token's team.",
	}, s.getAuditLog, false)
}

func (s *Server) getTelemetry(ctx context.Context, _ *mcp.CallToolRequest, in GetTelemetryInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	pts, err := c.GetTelemetry(ctx, operator.GetTelemetryInput{
		DeviceID: in.DeviceID,
		Metric:   in.Metric,
		From:     in.From,
		To:       in.To,
		Bucket:   in.Bucket,
		Agg:      in.Agg,
		FillLOCF: in.FillLOCF,
	})
	if err != nil {
		return fail(err)
	}
	return ok(pts)
}

func (s *Server) streamEvents(ctx context.Context, _ *mcp.CallToolRequest, in StreamEventsInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	timeout := time.Duration(in.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	events, err := c.ReadEvents(ctx, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(events)
}

func (s *Server) getUsage(ctx context.Context, _ *mcp.CallToolRequest, _ GetUsageInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	usage, err := c.GetUsage(ctx)
	if err != nil {
		return fail(err)
	}
	return ok(usage)
}

func (s *Server) getAuditLog(ctx context.Context, _ *mcp.CallToolRequest, in GetAuditLogInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.GetAuditLog(ctx, in.Page, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

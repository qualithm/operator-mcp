// Package server implements the Qualithm operator MCP server: the platform
// provisioning surface (authorities, enrollments, credentials, devices, API
// tokens) exposed as agent-native MCP tools.
//
// Every tool is a thin mapping over the shared operator client that also backs
// the qualithm CLI, so the human and agent surfaces never diverge. Mutating
// tools accept a dry_run flag and report the planned action without applying
// it; failures carry a stable code mirroring the CLI's exit-code contract.
package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// Config configures a [Server].
type Config struct {
	// Token is the member API token (Bearer) every tool authenticates with.
	Token string
	// BaseURL overrides the management API base URL. Empty uses the default.
	BaseURL string
}

// Server builds MCP tools backed by the operator client.
type Server struct {
	// newClient builds an operator client with the given dry-run setting. main
	// wires the real constructor from a token; tests inject a fake transport.
	newClient func(dryRun bool) (*operator.Client, error)
}

// New returns a [Server] that builds real operator clients from cfg. It fails
// fast when the token is missing or malformed.
func New(cfg Config) (*Server, error) {
	newClient := func(dryRun bool) (*operator.Client, error) {
		opts := []operator.Option{operator.WithDryRun(dryRun)}
		if cfg.BaseURL != "" {
			opts = append(opts, operator.WithBaseURL(cfg.BaseURL))
		}
		return operator.New(cfg.Token, opts...)
	}
	if _, err := newClient(false); err != nil {
		return nil, err
	}
	return &Server{newClient: newClient}, nil
}

// newServerWith builds a Server from a client factory. Used by tests to inject
// a fake transport.
func newServerWith(newClient func(dryRun bool) (*operator.Client, error)) *Server {
	return &Server{newClient: newClient}
}

// MCPServer builds the MCP server with every provisioning tool registered.
func (s *Server) MCPServer(version string) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "qualithm-operator",
		Title:   "Qualithm Operator",
		Version: version,
	}, nil)
	s.registerAuthorities(srv)
	s.registerEnrollments(srv)
	s.registerCredentials(srv)
	s.registerDevices(srv)
	s.registerTokens(srv)
	return srv
}

// Run serves the provisioning tools over stdio until ctx is cancelled.
func (s *Server) Run(ctx context.Context, version string) error {
	return s.MCPServer(version).Run(ctx, &mcp.StdioTransport{})
}

// ActionOut describes a mutation the server planned but did not send.
type ActionOut struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// Result is the uniform envelope every tool returns. It mirrors the API's own
// response envelope and the CLI's --json output so agents branch on the same
// shape regardless of tool.
type Result struct {
	// OK reports whether the call succeeded (a dry-run counts as success).
	OK bool `json:"ok"`
	// Code classifies a failure: auth, not_found, conflict, rate_limited, api,
	// or error. Empty on success.
	Code string `json:"code,omitempty"`
	// Message is a human-readable error message. Empty on success.
	Message string `json:"message,omitempty"`
	// DryRun is true when a mutation was planned but not applied.
	DryRun bool `json:"dryRun,omitempty"`
	// Action is the planned mutation, set only for dry-run results.
	Action *ActionOut `json:"action,omitempty"`
	// Data is the resource payload returned by the API on success.
	Data any `json:"data,omitempty"`
}

// Stable failure classifications, mirroring the CLI exit-code contract.
const (
	codeError       = "error"        // transport or unexpected error
	codeAuth        = "auth"         // 401 / 403
	codeNotFound    = "not_found"    // 404
	codeConflict    = "conflict"     // 409
	codeRateLimited = "rate_limited" // 429
	codeAPI         = "api"          // other non-2xx response
)

func codeForStatus(status int) string {
	switch status {
	case 401, 403:
		return codeAuth
	case 404:
		return codeNotFound
	case 409:
		return codeConflict
	case 429:
		return codeRateLimited
	default:
		return codeAPI
	}
}

// classify maps a client error to a stable failure code.
func classify(err error) string {
	var ce *operator.ClientError
	if errors.As(err, &ce) {
		return codeForStatus(ce.StatusCode)
	}
	return codeError
}

// ok wraps a successful payload.
func ok(data any) (*mcp.CallToolResult, Result, error) {
	return nil, Result{OK: true, Data: data}, nil
}

// fail converts any error into a failed tool result carrying a stable code.
// Dry-run "errors" from the client are reported as successful planned actions.
func fail(err error) (*mcp.CallToolResult, Result, error) {
	var dre *operator.DryRunError
	if errors.As(err, &dre) {
		r := Result{OK: true, DryRun: true, Action: &ActionOut{
			Method: dre.Action.Method,
			Path:   dre.Action.Path,
		}}
		return nil, r, nil
	}
	r := Result{OK: false, Code: classify(err), Message: err.Error()}
	res := &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("%s: %s", r.Code, err.Error())}},
	}
	return res, r, nil
}

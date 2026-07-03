package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

const tok = "qmt_selector.verifier"

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// record captures the last request a tool issued.
type record struct {
	method string
	path   string
	query  string
	body   string
}

// testServer returns a Server whose client answers every request with the given
// status and body, recording the request into rec when non-nil.
func testServer(t *testing.T, status int, body string, rec *record) *Server {
	t.Helper()
	return newServerWith(func(dryRun bool) (*operator.Client, error) {
		rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if rec != nil {
				rec.method = req.Method
				rec.path = req.URL.Path
				rec.query = req.URL.RawQuery
				if req.Body != nil {
					b, _ := io.ReadAll(req.Body)
					rec.body = string(b)
				}
			}
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		})
		return operator.New(tok,
			operator.WithDryRun(dryRun),
			operator.WithHTTPClient(&http.Client{Transport: rt}),
		)
	})
}

func envelope(data string) string { return `{"data":` + data + `}` }

func ctx() context.Context { return context.Background() }

func TestNewRequiresToken(t *testing.T) {
	if _, err := New(Config{Token: ""}); err == nil {
		t.Fatal("expected error for empty token")
	}
	if _, err := New(Config{Token: "nope"}); err == nil {
		t.Fatal("expected error for token without prefix")
	}
	if _, err := New(Config{Token: tok, BaseURL: "https://example.test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMCPServerRegistersTools(t *testing.T) {
	s := testServer(t, 200, envelope("{}"), nil)
	srv := s.MCPServer("test")
	if srv == nil {
		t.Fatal("expected non-nil MCP server")
	}
}

func TestCodeForStatus(t *testing.T) {
	cases := map[int]string{
		401: codeAuth, 403: codeAuth,
		404: codeNotFound, 409: codeConflict, 429: codeRateLimited,
		500: codeAPI, 418: codeAPI,
	}
	for status, want := range cases {
		if got := codeForStatus(status); got != want {
			t.Errorf("codeForStatus(%d) = %q, want %q", status, got, want)
		}
	}
}

// --- list tools (GET, no dry-run) ---

func TestListAuthorities(t *testing.T) {
	rec := &record{}
	s := testServer(t, 200, envelope(`{"current":1,"last":1,"items":[{"id":"auth_1"}]}`), rec)
	res, out, err := s.listAuthorities(ctx(), nil, ListAuthoritiesInput{Page: 2, Limit: 5})
	if err != nil || res != nil {
		t.Fatalf("res=%v err=%v", res, err)
	}
	if !out.OK {
		t.Fatal("want OK")
	}
	if rec.method != http.MethodGet || rec.path != "/authorities" {
		t.Fatalf("unexpected request %s %s", rec.method, rec.path)
	}
	if !strings.Contains(rec.query, "page=2") || !strings.Contains(rec.query, "limit=5") {
		t.Fatalf("missing page params: %q", rec.query)
	}
}

func TestListEnrollmentsDevicesTokens(t *testing.T) {
	s := testServer(t, 200, envelope(`{"current":1,"last":1,"items":[]}`), nil)
	if _, out, _ := s.listEnrollments(ctx(), nil, ListEnrollmentsInput{}); !out.OK {
		t.Error("enrollments not ok")
	}
	if _, out, _ := s.listDevices(ctx(), nil, ListDevicesInput{}); !out.OK {
		t.Error("devices not ok")
	}
	if _, out, _ := s.listSpaceDevices(ctx(), nil, ListSpaceDevicesInput{SpaceID: "sp_1"}); !out.OK {
		t.Error("space devices not ok")
	}
	if _, out, _ := s.listAPITokens(ctx(), nil, ListAPITokensInput{}); !out.OK {
		t.Error("tokens not ok")
	}
}

func TestListCredentialsAndGetDevice(t *testing.T) {
	rec := &record{}
	s := testServer(t, 200, envelope(`[{"id":"cred_1"}]`), rec)
	if _, out, _ := s.listCredentials(ctx(), nil, ListCredentialsInput{DeviceID: "dev_1"}); !out.OK {
		t.Error("credentials not ok")
	}
	if !strings.Contains(rec.path, "dev_1") {
		t.Fatalf("device id not in path: %s", rec.path)
	}

	s = testServer(t, 200, envelope(`{"id":"dev_1","name":"edge"}`), nil)
	if _, out, _ := s.getDevice(ctx(), nil, GetDeviceInput{DeviceID: "dev_1"}); !out.OK {
		t.Error("get device not ok")
	}
}

// --- mutations (applied) ---

func TestCreateAuthorityApplied(t *testing.T) {
	rec := &record{}
	s := testServer(t, 201, envelope(`{"id":"auth_1","name":"ca"}`), rec)
	res, out, err := s.createAuthority(ctx(), nil, CreateAuthorityInput{Name: "ca", Kind: "platform"})
	if err != nil || res != nil || !out.OK {
		t.Fatalf("res=%v out=%+v err=%v", res, out, err)
	}
	if rec.method != http.MethodPost || rec.path != "/authorities" {
		t.Fatalf("unexpected %s %s", rec.method, rec.path)
	}
	if !strings.Contains(rec.body, `"name":"ca"`) {
		t.Fatalf("body missing name: %s", rec.body)
	}
}

func TestMutationsApplied(t *testing.T) {
	s := testServer(t, 200, envelope(`{"id":"x"}`), nil)
	checks := []struct {
		name string
		fn   func() Result
	}{
		{"create_enrollment", func() Result {
			_, o, _ := s.createEnrollment(ctx(), nil, CreateEnrollmentInput{SpaceID: "sp_1"})
			return o
		}},
		{"revoke_enrollment", func() Result {
			_, o, _ := s.revokeEnrollment(ctx(), nil, RevokeEnrollmentInput{EnrollmentID: "en_1"})
			return o
		}},
		{"revoke_authority", func() Result {
			_, o, _ := s.revokeAuthority(ctx(), nil, RevokeAuthorityInput{AuthorityID: "auth_1"})
			return o
		}},
		{"mint_credential", func() Result {
			_, o, _ := s.mintCredential(ctx(), nil, MintCredentialInput{DeviceID: "dev_1"})
			return o
		}},
		{"issue_cert", func() Result {
			_, o, _ := s.issueCert(ctx(), nil, IssueCertInput{DeviceID: "dev_1", CSRPEM: "pem"})
			return o
		}},
		{"rotate_credential", func() Result {
			_, o, _ := s.rotateCredential(ctx(), nil, RotateCredentialInput{DeviceID: "dev_1", CredentialID: "cred_1", Revoke: true})
			return o
		}},
		{"revoke_credential", func() Result {
			_, o, _ := s.revokeCredential(ctx(), nil, RevokeCredentialInput{DeviceID: "dev_1", CredentialID: "cred_1"})
			return o
		}},
		{"create_device", func() Result {
			_, o, _ := s.createDevice(ctx(), nil, CreateDeviceInput{SpaceID: "sp_1"})
			return o
		}},
		{"update_device", func() Result {
			_, o, _ := s.updateDevice(ctx(), nil, UpdateDeviceInput{DeviceID: "dev_1", Name: "n"})
			return o
		}},
		{"delete_device", func() Result {
			_, o, _ := s.deleteDevice(ctx(), nil, DeleteDeviceInput{DeviceID: "dev_1"})
			return o
		}},
		{"create_api_token", func() Result {
			_, o, _ := s.createAPIToken(ctx(), nil, CreateAPITokenInput{Name: "t"})
			return o
		}},
		{"revoke_api_token", func() Result {
			_, o, _ := s.revokeAPIToken(ctx(), nil, RevokeAPITokenInput{TokenID: "tok_1"})
			return o
		}},
	}
	for _, c := range checks {
		if out := c.fn(); !out.OK {
			t.Errorf("%s: want OK, got %+v", c.name, out)
		}
	}
}

// --- dry-run ---

func TestDryRunReportsPlannedAction(t *testing.T) {
	rec := &record{}
	s := testServer(t, 200, envelope(`{}`), rec)
	_, out, err := s.createDevice(ctx(), nil, CreateDeviceInput{SpaceID: "sp_1", DryRun: true})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !out.OK || !out.DryRun || out.Action == nil {
		t.Fatalf("want dry-run success with action, got %+v", out)
	}
	if out.Action.Method != http.MethodPost {
		t.Fatalf("want POST action, got %s", out.Action.Method)
	}
	if rec.method != "" {
		t.Fatalf("dry-run must not send a request, sent %s", rec.method)
	}
}

// --- error classification ---

func TestErrorClassification(t *testing.T) {
	cases := map[int]string{403: codeAuth, 404: codeNotFound, 409: codeConflict, 429: codeRateLimited, 500: codeAPI}
	for status, want := range cases {
		s := testServer(t, status, `{"message":"boom"}`, nil)
		res, out, err := s.getDevice(ctx(), nil, GetDeviceInput{DeviceID: "dev_1"})
		if err != nil {
			t.Fatalf("status %d: unexpected go error %v", status, err)
		}
		if out.OK || out.Code != want {
			t.Fatalf("status %d: want code %q, got %+v", status, want, out)
		}
		if res == nil || !res.IsError {
			t.Fatalf("status %d: want IsError result", status)
		}
	}
}

func TestOutputMarshalsCleanly(t *testing.T) {
	s := testServer(t, 200, envelope(`{"id":"dev_1"}`), nil)
	_, out, _ := s.getDevice(ctx(), nil, GetDeviceInput{DeviceID: "dev_1"})
	if _, err := json.Marshal(out); err != nil {
		t.Fatalf("result did not marshal: %v", err)
	}
}

// handlerCall names one handler invocation so the branch-coverage tests can run
// every tool against a shared server.
type handlerCall struct {
	name string
	fn   func() (*mcp.CallToolResult, Result, error)
}

func allHandlers(s *Server) []handlerCall {
	return []handlerCall{
		{"list_authorities", func() (*mcp.CallToolResult, Result, error) {
			return s.listAuthorities(ctx(), nil, ListAuthoritiesInput{})
		}},
		{"create_authority", func() (*mcp.CallToolResult, Result, error) {
			return s.createAuthority(ctx(), nil, CreateAuthorityInput{Name: "ca", Kind: "platform"})
		}},
		{"revoke_authority", func() (*mcp.CallToolResult, Result, error) {
			return s.revokeAuthority(ctx(), nil, RevokeAuthorityInput{AuthorityID: "auth_1"})
		}},
		{"list_enrollments", func() (*mcp.CallToolResult, Result, error) {
			return s.listEnrollments(ctx(), nil, ListEnrollmentsInput{})
		}},
		{"create_enrollment", func() (*mcp.CallToolResult, Result, error) {
			return s.createEnrollment(ctx(), nil, CreateEnrollmentInput{SpaceID: "sp_1"})
		}},
		{"revoke_enrollment", func() (*mcp.CallToolResult, Result, error) {
			return s.revokeEnrollment(ctx(), nil, RevokeEnrollmentInput{EnrollmentID: "en_1"})
		}},
		{"list_credentials", func() (*mcp.CallToolResult, Result, error) {
			return s.listCredentials(ctx(), nil, ListCredentialsInput{DeviceID: "dev_1"})
		}},
		{"mint_credential", func() (*mcp.CallToolResult, Result, error) {
			return s.mintCredential(ctx(), nil, MintCredentialInput{DeviceID: "dev_1"})
		}},
		{"issue_cert", func() (*mcp.CallToolResult, Result, error) {
			return s.issueCert(ctx(), nil, IssueCertInput{DeviceID: "dev_1", CSRPEM: "pem"})
		}},
		{"rotate_credential", func() (*mcp.CallToolResult, Result, error) {
			return s.rotateCredential(ctx(), nil, RotateCredentialInput{DeviceID: "dev_1", CredentialID: "cred_1"})
		}},
		{"revoke_credential", func() (*mcp.CallToolResult, Result, error) {
			return s.revokeCredential(ctx(), nil, RevokeCredentialInput{DeviceID: "dev_1", CredentialID: "cred_1"})
		}},
		{"list_devices", func() (*mcp.CallToolResult, Result, error) {
			return s.listDevices(ctx(), nil, ListDevicesInput{})
		}},
		{"list_space_devices", func() (*mcp.CallToolResult, Result, error) {
			return s.listSpaceDevices(ctx(), nil, ListSpaceDevicesInput{SpaceID: "sp_1"})
		}},
		{"get_device", func() (*mcp.CallToolResult, Result, error) {
			return s.getDevice(ctx(), nil, GetDeviceInput{DeviceID: "dev_1"})
		}},
		{"create_device", func() (*mcp.CallToolResult, Result, error) {
			return s.createDevice(ctx(), nil, CreateDeviceInput{SpaceID: "sp_1"})
		}},
		{"update_device", func() (*mcp.CallToolResult, Result, error) {
			return s.updateDevice(ctx(), nil, UpdateDeviceInput{DeviceID: "dev_1", Name: "n"})
		}},
		{"delete_device", func() (*mcp.CallToolResult, Result, error) {
			return s.deleteDevice(ctx(), nil, DeleteDeviceInput{DeviceID: "dev_1"})
		}},
		{"list_api_tokens", func() (*mcp.CallToolResult, Result, error) {
			return s.listAPITokens(ctx(), nil, ListAPITokensInput{})
		}},
		{"create_api_token", func() (*mcp.CallToolResult, Result, error) {
			return s.createAPIToken(ctx(), nil, CreateAPITokenInput{Name: "t"})
		}},
		{"revoke_api_token", func() (*mcp.CallToolResult, Result, error) {
			return s.revokeAPIToken(ctx(), nil, RevokeAPITokenInput{TokenID: "tok_1"})
		}},
	}
}

// TestFactoryErrorFailsEveryTool covers the client-construction failure branch
// in every handler.
func TestFactoryErrorFailsEveryTool(t *testing.T) {
	boom := errors.New("client unavailable")
	s := newServerWith(func(bool) (*operator.Client, error) { return nil, boom })
	for _, h := range allHandlers(s) {
		res, out, err := h.fn()
		if err != nil {
			t.Errorf("%s: unexpected go error %v", h.name, err)
		}
		if out.OK || out.Code != codeError {
			t.Errorf("%s: want code %q, got %+v", h.name, codeError, out)
		}
		if res == nil || !res.IsError {
			t.Errorf("%s: want IsError result", h.name)
		}
	}
}

// TestAPIErrorFailsEveryTool covers the API-call failure branch in every
// handler.
func TestAPIErrorFailsEveryTool(t *testing.T) {
	s := testServer(t, 500, `{"message":"boom"}`, nil)
	for _, h := range allHandlers(s) {
		res, out, err := h.fn()
		if err != nil {
			t.Errorf("%s: unexpected go error %v", h.name, err)
		}
		if out.OK || out.Code != codeAPI {
			t.Errorf("%s: want code %q, got %+v", h.name, codeAPI, out)
		}
		if res == nil || !res.IsError {
			t.Errorf("%s: want IsError result", h.name)
		}
	}
}

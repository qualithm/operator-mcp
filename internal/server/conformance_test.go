package server

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registry returns a Server whose MCPServer has run, so its registered
// metadata reflects the real tool wiring.
func registry(t *testing.T) *Server {
	t.Helper()
	s := testServer(t, 200, envelope("{}"), nil)
	s.MCPServer("test")
	return s
}

// findMeta returns the registration metadata for name from s, which must have
// run MCPServer. The metadata's invoke closure is bound to s, so calls exercise
// s's client factory.
func findMeta(t *testing.T, s *Server, name string) toolMeta {
	t.Helper()
	for _, meta := range s.registered {
		if meta.name == name {
			return meta
		}
	}
	t.Fatalf("tool %s not registered", name)
	return toolMeta{}
}

// fieldByJSONTag returns the struct field index carrying the given json tag
// name, or -1 when absent.
func fieldByJSONTag(typ reflect.Type, name string) int {
	for i := 0; i < typ.NumField(); i++ {
		if strings.Split(typ.Field(i).Tag.Get("json"), ",")[0] == name {
			return i
		}
	}
	return -1
}

// zeroInput builds a zero value of the tool's input type, setting its dryRun
// flag when asked.
func zeroInput(t *testing.T, meta toolMeta, dryRun bool) any {
	t.Helper()
	v := reflect.New(meta.input).Elem()
	if dryRun {
		i := fieldByJSONTag(meta.input, "dryRun")
		if i < 0 {
			t.Fatalf("tool %s: no dryRun field to set", meta.name)
		}
		v.Field(i).SetBool(true)
	}
	return v.Interface()
}

// TestRegistryMatchesExposedTools walks the live MCP server over an in-memory
// transport and asserts every exposed tool was registered through addTool
// (and so carries contract metadata), with no duplicates and nothing recorded
// but unexposed.
func TestRegistryMatchesExposedTools(t *testing.T) {
	s := registry(t)

	cTransport, sTransport := mcp.NewInMemoryTransports()
	ss, err := s.MCPServer("test").Connect(ctx(), sTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	defer func() { _ = ss.Close() }()
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "conformance-test"}, nil).Connect(ctx(), cTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() { _ = cs.Close() }()

	exposed := map[string]bool{}
	params := &mcp.ListToolsParams{}
	for {
		res, err := cs.ListTools(ctx(), params)
		if err != nil {
			t.Fatalf("list tools: %v", err)
		}
		for _, tool := range res.Tools {
			if exposed[tool.Name] {
				t.Errorf("tool %s exposed twice", tool.Name)
			}
			exposed[tool.Name] = true
		}
		if res.NextCursor == "" {
			break
		}
		params.Cursor = res.NextCursor
	}

	recorded := map[string]bool{}
	for _, meta := range s.registered {
		if recorded[meta.name] {
			t.Errorf("tool %s registered twice", meta.name)
		}
		recorded[meta.name] = true
		if !exposed[meta.name] {
			t.Errorf("tool %s registered through addTool but not exposed by the server", meta.name)
		}
	}
	for name := range exposed {
		if !recorded[name] {
			t.Errorf("tool %s exposed by the server but not registered through addTool; the conformance contract cannot see it", name)
		}
	}
}

// TestMutatingToolsHonourDryRun asserts every tool marked mutating accepts a
// dryRun input flag and that setting it plans the mutation without any HTTP
// request leaving the client.
func TestMutatingToolsHonourDryRun(t *testing.T) {
	for _, meta := range registry(t).registered {
		if !meta.mutating {
			continue
		}
		t.Run(meta.name, func(t *testing.T) {
			if fieldByJSONTag(meta.input, "dryRun") < 0 {
				t.Fatalf("mutating tool input %s lacks a `dryRun` json field", meta.input)
			}
			rec := &record{}
			// A 500 transport doubles as the tripwire: if the dry-run
			// short-circuit ever breaks, the request is sent and the result
			// comes back as an api failure instead of a planned action.
			s := testServer(t, 500, `{"message":"must not be sent"}`, rec)
			s.MCPServer("test")
			bound := findMeta(t, s, meta.name)
			res, out, err := bound.invoke(ctx(), zeroInput(t, bound, true))
			if err != nil {
				t.Fatalf("invoke: %v", err)
			}
			if res != nil {
				t.Errorf("dry-run returned a non-nil CallToolResult; want success via the output value")
			}
			if !out.OK || !out.DryRun || out.Action == nil {
				t.Errorf("dry-run result = %+v; want OK+DryRun with a planned Action", out)
			}
			if out.Action != nil && (out.Action.Method == "" || out.Action.Method == http.MethodGet) {
				t.Errorf("mutating tool planned a non-mutating action: %+v", out.Action)
			}
			if rec.method != "" {
				t.Errorf("dry-run still sent %s %s", rec.method, rec.path)
			}
		})
	}
}

// TestToolsReturnStableErrorCodes drives every tool against each classified
// HTTP failure and asserts the uniform envelope: OK false, the stable code for
// the status, the upstream message, and an IsError CallToolResult.
func TestToolsReturnStableErrorCodes(t *testing.T) {
	statuses := map[int]string{
		401: codeAuth,
		403: codeAuth,
		404: codeNotFound,
		409: codeConflict,
		429: codeRateLimited,
		500: codeAPI,
	}
	for _, meta := range registry(t).registered {
		t.Run(meta.name, func(t *testing.T) {
			for status, wantCode := range statuses {
				s := testServer(t, status, `{"message":"boom"}`, nil)
				s.MCPServer("test")
				bound := findMeta(t, s, meta.name)
				res, out, err := bound.invoke(ctx(), zeroInput(t, bound, false))
				if err != nil {
					t.Fatalf("status %d: invoke: %v", status, err)
				}
				if out.OK || out.Code != wantCode {
					t.Errorf("status %d: result = %+v; want OK=false code=%q", status, out, wantCode)
				}
				if !strings.Contains(out.Message, "boom") {
					t.Errorf("status %d: message = %q; want the upstream message carried through", status, out.Message)
				}
				if res == nil || !res.IsError {
					t.Errorf("status %d: want an IsError CallToolResult", status)
				}
			}
		})
	}
}

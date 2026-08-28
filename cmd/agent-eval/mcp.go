package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// mcpToolsClient calls the qualithm-mcp server over stdio, the same transport
// an agent uses. toolResult mirrors the server's uniform result envelope.
type mcpToolsClient struct {
	session *mcp.ClientSession
}

// toolEnvelope is the server's Result shape as seen in structured content.
type toolEnvelope struct {
	OK      bool            `json:"ok"`
	Code    string          `json:"code,omitempty"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// connectMCPServer spawns the qualithm-mcp binary and connects an MCP client
// to it over stdio. The token and API URL reach the server via its
// environment.
func connectMCPServer(ctx context.Context, serverBin string, env []string) (*mcpToolsClient, func(), error) {
	// The harness execs exactly the server binary its operator passed; an
	// absolute path to a regular file is all it accepts.
	if !filepath.IsAbs(serverBin) {
		return nil, nil, fmt.Errorf("server binary path must be absolute, got %q", serverBin)
	}
	info, err := os.Stat(serverBin)
	if err != nil || !info.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("server binary %q is not a regular file", serverBin)
	}
	cmd := exec.Command(serverBin) // #nosec G702 -- path validated above; spawning the configured server is the harness's purpose
	cmd.Env = env
	transport := &mcp.CommandTransport{Command: cmd}
	client := mcp.NewClient(&mcp.Implementation{Name: "agent-eval"}, nil)
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to %s: %w", serverBin, err)
	}
	cleanup := func() {
		_ = session.Close()
		_ = cmd.Wait()
	}
	return &mcpToolsClient{session: session}, cleanup, nil
}

// callTool implements toolCaller.
func (c *mcpToolsClient) callTool(ctx context.Context, name string, args map[string]any, out any) error {
	res, err := c.session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return decodeToolResult(name, res, out)
}

// decodeToolResult unpacks one tool response: the uniform envelope's data
// into out, or an error carrying the envelope's stable code.
func decodeToolResult(name string, res *mcp.CallToolResult, out any) error {
	var env toolEnvelope
	if res.StructuredContent != nil {
		raw, err := json.Marshal(res.StructuredContent)
		if err != nil {
			return fmt.Errorf("%s: encode structured content: %w", name, err)
		}
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("%s: decode result envelope: %w", name, err)
		}
	}
	if res.IsError || !env.OK {
		if env.Code != "" {
			return fmt.Errorf("%s: %s: %s", name, env.Code, env.Message)
		}
		return fmt.Errorf("%s: tool reported failure without an envelope code", name)
	}
	if out != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return fmt.Errorf("%s: decode data: %w", name, err)
		}
	}
	return nil
}

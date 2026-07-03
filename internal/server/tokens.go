package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// ListAPITokensInput lists member API tokens.
type ListAPITokensInput struct {
	Page  int `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit int `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// CreateAPITokenInput mints a member API token.
type CreateAPITokenInput struct {
	Name          string `json:"name,omitempty" jsonschema:"human-readable token name"`
	ExpiresInDays int    `json:"expiresInDays,omitempty" jsonschema:"token lifetime in days; 0 uses the server default of 90"`
	DryRun        bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// RevokeAPITokenInput revokes a member API token.
type RevokeAPITokenInput struct {
	TokenID string `json:"tokenId" jsonschema:"id of the API token to revoke"`
	DryRun  bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

func (s *Server) registerTokens(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_api_tokens",
		Description: "List member API tokens for the token's team (metadata only; secrets are never listed).",
	}, s.listAPITokens)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_api_token",
		Description: "Mint a member API token. The plaintext secret is returned once.",
	}, s.createAPIToken)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "revoke_api_token",
		Description: "Revoke a member API token by id.",
	}, s.revokeAPIToken)
}

func (s *Server) listAPITokens(ctx context.Context, _ *mcp.CallToolRequest, in ListAPITokensInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListAPITokens(ctx, in.Page, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) createAPIToken(ctx context.Context, _ *mcp.CallToolRequest, in CreateAPITokenInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	t, err := c.CreateAPIToken(ctx, operator.CreateAPITokenInput{
		Name:          in.Name,
		ExpiresInDays: in.ExpiresInDays,
	})
	if err != nil {
		return fail(err)
	}
	return ok(t)
}

func (s *Server) revokeAPIToken(ctx context.Context, _ *mcp.CallToolRequest, in RevokeAPITokenInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.RevokeAPIToken(ctx, in.TokenID); err != nil {
		return fail(err)
	}
	return ok(nil)
}

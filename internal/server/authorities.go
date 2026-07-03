package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// ListAuthoritiesInput lists certificate authorities.
type ListAuthoritiesInput struct {
	Page  int `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit int `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// CreateAuthorityInput creates a certificate authority.
type CreateAuthorityInput struct {
	Name           string `json:"name" jsonschema:"human-readable authority name"`
	Kind           string `json:"kind" jsonschema:"authority kind: platform or byo"`
	CertificatePEM string `json:"certificatePem,omitempty" jsonschema:"PEM certificate, required when kind is byo"`
	DryRun         bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// RevokeAuthorityInput revokes a certificate authority.
type RevokeAuthorityInput struct {
	AuthorityID string `json:"authorityId" jsonschema:"id of the authority to revoke"`
	DryRun      bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

func (s *Server) registerAuthorities(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_authorities",
		Description: "List device certificate authorities for the token's team.",
	}, s.listAuthorities)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_authority",
		Description: "Create a device certificate authority (platform-generated or BYO).",
	}, s.createAuthority)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "revoke_authority",
		Description: "Revoke a device certificate authority by id.",
	}, s.revokeAuthority)
}

func (s *Server) listAuthorities(ctx context.Context, _ *mcp.CallToolRequest, in ListAuthoritiesInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListAuthorities(ctx, in.Page, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) createAuthority(ctx context.Context, _ *mcp.CallToolRequest, in CreateAuthorityInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	a, err := c.CreateAuthority(ctx, operator.CreateAuthorityInput{
		Name:           in.Name,
		Kind:           in.Kind,
		CertificatePEM: in.CertificatePEM,
	})
	if err != nil {
		return fail(err)
	}
	return ok(a)
}

func (s *Server) revokeAuthority(ctx context.Context, _ *mcp.CallToolRequest, in RevokeAuthorityInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.RevokeAuthority(ctx, in.AuthorityID); err != nil {
		return fail(err)
	}
	return ok(nil)
}

package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// ListCredentialsInput lists a device's credentials.
type ListCredentialsInput struct {
	DeviceID string `json:"deviceId" jsonschema:"id of the device whose credentials to list"`
}

// MintCredentialInput mints a token credential for a device.
type MintCredentialInput struct {
	DeviceID  string `json:"deviceId" jsonschema:"id of the device to mint a credential for"`
	Label     string `json:"label,omitempty" jsonschema:"human-readable credential label"`
	ExpiresAt string `json:"expiresAt,omitempty" jsonschema:"optional ISO 8601 expiry timestamp in the future"`
	DryRun    bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// IssueCertInput issues an mTLS certificate credential for a device.
type IssueCertInput struct {
	DeviceID      string `json:"deviceId" jsonschema:"id of the device to issue a certificate for"`
	CSRPEM        string `json:"csrPem" jsonschema:"PEM-encoded certificate signing request"`
	Label         string `json:"label,omitempty" jsonschema:"human-readable credential label"`
	ExpiresInDays int    `json:"expiresInDays,omitempty" jsonschema:"certificate lifetime in days; 0 uses the server default of 30"`
	DryRun        bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// RotateCredentialInput rotates a device credential, optionally revoking the old one.
type RotateCredentialInput struct {
	DeviceID     string `json:"deviceId" jsonschema:"id of the device that owns the credential"`
	CredentialID string `json:"credentialId" jsonschema:"id of the credential to rotate"`
	Label        string `json:"label,omitempty" jsonschema:"human-readable label for the new credential"`
	ExpiresAt    string `json:"expiresAt,omitempty" jsonschema:"optional ISO 8601 expiry timestamp in the future"`
	Revoke       bool   `json:"revoke,omitempty" jsonschema:"revoke the old credential immediately after rotating"`
	DryRun       bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// RevokeCredentialInput revokes a device credential.
type RevokeCredentialInput struct {
	DeviceID     string `json:"deviceId" jsonschema:"id of the device that owns the credential"`
	CredentialID string `json:"credentialId" jsonschema:"id of the credential to revoke"`
	DryRun       bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

func (s *Server) registerCredentials(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_credentials",
		Description: "List the credentials (tokens and certificates) for a device.",
	}, s.listCredentials)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "mint_credential",
		Description: "Mint a token credential for a device. The plaintext secret is returned once.",
	}, s.mintCredential)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "issue_cert",
		Description: "Issue an mTLS certificate credential for a device from a CSR. The certificate is returned once.",
	}, s.issueCert)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "rotate_credential",
		Description: "Rotate a device credential, optionally revoking the old one. The new secret is returned once.",
	}, s.rotateCredential)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "revoke_credential",
		Description: "Revoke a device credential by id.",
	}, s.revokeCredential)
}

func (s *Server) listCredentials(ctx context.Context, _ *mcp.CallToolRequest, in ListCredentialsInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	creds, err := c.ListCredentials(ctx, in.DeviceID)
	if err != nil {
		return fail(err)
	}
	return ok(creds)
}

func (s *Server) mintCredential(ctx context.Context, _ *mcp.CallToolRequest, in MintCredentialInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	cred, err := c.MintCredential(ctx, in.DeviceID, operator.MintCredentialInput{
		Label:     in.Label,
		ExpiresAt: in.ExpiresAt,
	})
	if err != nil {
		return fail(err)
	}
	return ok(cred)
}

func (s *Server) issueCert(ctx context.Context, _ *mcp.CallToolRequest, in IssueCertInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	cred, err := c.IssueCert(ctx, in.DeviceID, operator.IssueCertInput{
		CSRPEM:        in.CSRPEM,
		Label:         in.Label,
		ExpiresInDays: in.ExpiresInDays,
	})
	if err != nil {
		return fail(err)
	}
	return ok(cred)
}

func (s *Server) rotateCredential(ctx context.Context, _ *mcp.CallToolRequest, in RotateCredentialInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	cred, err := c.RotateCredential(ctx, in.DeviceID, in.CredentialID, operator.MintCredentialInput{
		Label:     in.Label,
		ExpiresAt: in.ExpiresAt,
	}, in.Revoke)
	if err != nil {
		return fail(err)
	}
	return ok(cred)
}

func (s *Server) revokeCredential(ctx context.Context, _ *mcp.CallToolRequest, in RevokeCredentialInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.RevokeCredential(ctx, in.DeviceID, in.CredentialID); err != nil {
		return fail(err)
	}
	return ok(nil)
}

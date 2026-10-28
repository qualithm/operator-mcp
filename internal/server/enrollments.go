package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// ListEnrollmentsInput lists enrollment claim codes.
type ListEnrollmentsInput struct {
	Cursor string `json:"cursor,omitempty" jsonschema:"the nextCursor from the previous page; empty for the first page"`
	Limit  int    `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// CreateEnrollmentInput creates a one-time enrollment claim code.
type CreateEnrollmentInput struct {
	SpaceID          string `json:"spaceId" jsonschema:"id of the space the device will enroll into"`
	Label            string `json:"label,omitempty" jsonschema:"human-readable label for the enrollment"`
	ExpiresInMinutes int    `json:"expiresInMinutes,omitempty" jsonschema:"lifetime in minutes; 0 uses the server default of 1440"`
	DryRun           bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// RevokeEnrollmentInput revokes an enrollment claim code.
type RevokeEnrollmentInput struct {
	EnrollmentID string `json:"enrollmentId" jsonschema:"id of the enrollment to revoke"`
	DryRun       bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

func (s *Server) registerEnrollments(srv *mcp.Server) {
	addTool(s, srv, &mcp.Tool{
		Name:        "list_enrollments",
		Description: "List one-time device enrollment claim codes for the token's team. Pass the result's nextCursor back as cursor for the next page; it is null on the last page.",
	}, s.listEnrollments, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "create_enrollment",
		Description: "Create a one-time device enrollment claim code. The plaintext code is returned once.",
	}, s.createEnrollment, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "revoke_enrollment",
		Description: "Revoke a device enrollment claim code by id.",
	}, s.revokeEnrollment, true)
}

func (s *Server) listEnrollments(ctx context.Context, _ *mcp.CallToolRequest, in ListEnrollmentsInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListEnrollments(ctx, in.Cursor, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) createEnrollment(ctx context.Context, _ *mcp.CallToolRequest, in CreateEnrollmentInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	e, err := c.CreateEnrollment(ctx, operator.CreateEnrollmentInput{
		SpaceID:          in.SpaceID,
		Label:            in.Label,
		ExpiresInMinutes: in.ExpiresInMinutes,
	})
	if err != nil {
		return fail(err)
	}
	return ok(e)
}

func (s *Server) revokeEnrollment(ctx context.Context, _ *mcp.CallToolRequest, in RevokeEnrollmentInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.RevokeEnrollment(ctx, in.EnrollmentID); err != nil {
		return fail(err)
	}
	return ok(nil)
}

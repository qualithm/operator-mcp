package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// ListSpacesInput lists spaces for the token's team.
type ListSpacesInput struct {
	Page  int `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit int `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// GetSpaceInput fetches a single space.
type GetSpaceInput struct {
	SpaceID string `json:"spaceId" jsonschema:"id of the space to fetch"`
}

// CreateSpaceInput creates a space in a device zone. The platform assigns the
// initial name; rename with update_space.
type CreateSpaceInput struct {
	Zone   string `json:"zone" jsonschema:"device zone to create the space in"`
	DryRun bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// UpdateSpaceInput renames a space.
type UpdateSpaceInput struct {
	SpaceID string `json:"spaceId" jsonschema:"id of the space to rename"`
	Name    string `json:"name" jsonschema:"new space name"`
	DryRun  bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// DeleteSpaceInput deletes a space, cascading deletion to its devices.
type DeleteSpaceInput struct {
	SpaceID string `json:"spaceId" jsonschema:"id of the space to delete"`
	DryRun  bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

func (s *Server) registerSpaces(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_spaces",
		Description: "List spaces for the token's team.",
	}, s.listSpaces)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_space",
		Description: "Fetch a single space by id.",
	}, s.getSpace)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_space",
		Description: "Create a space in a device zone.",
	}, s.createSpace)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_space",
		Description: "Rename a space.",
	}, s.updateSpace)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "delete_space",
		Description: "Delete a space by id, cascading to its devices.",
	}, s.deleteSpace)
}

func (s *Server) listSpaces(ctx context.Context, _ *mcp.CallToolRequest, in ListSpacesInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListSpaces(ctx, in.Page, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) getSpace(ctx context.Context, _ *mcp.CallToolRequest, in GetSpaceInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	space, err := c.GetSpace(ctx, in.SpaceID)
	if err != nil {
		return fail(err)
	}
	return ok(space)
}

func (s *Server) createSpace(ctx context.Context, _ *mcp.CallToolRequest, in CreateSpaceInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	space, err := c.CreateSpace(ctx, operator.CreateSpaceInput{Zone: in.Zone})
	if err != nil {
		return fail(err)
	}
	return ok(space)
}

func (s *Server) updateSpace(ctx context.Context, _ *mcp.CallToolRequest, in UpdateSpaceInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.UpdateSpace(ctx, in.SpaceID, in.Name); err != nil {
		return fail(err)
	}
	return ok(nil)
}

func (s *Server) deleteSpace(ctx context.Context, _ *mcp.CallToolRequest, in DeleteSpaceInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.DeleteSpace(ctx, in.SpaceID); err != nil {
		return fail(err)
	}
	return ok(nil)
}

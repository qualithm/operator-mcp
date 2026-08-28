package server

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// ListDashboardsInput lists dashboards for the token's team.
type ListDashboardsInput struct {
	Page  int `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit int `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// CreateDashboardInput creates a dashboard. Omit spaceId for a team-wide
// dashboard; payload is the ordered widget array (empty starts blank).
type CreateDashboardInput struct {
	Name    string          `json:"name" jsonschema:"dashboard name"`
	SpaceID string          `json:"spaceId,omitempty" jsonschema:"id of the space to scope the dashboard to; omit for team-wide"`
	Payload json.RawMessage `json:"payload,omitempty" jsonschema:"ordered widget array; omit for an empty dashboard"`
	DryRun  bool            `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// GetDashboardInput fetches a single dashboard.
type GetDashboardInput struct {
	DashboardID string `json:"dashboardId" jsonschema:"id of the dashboard to fetch"`
}

// UpdateDashboardInput updates a dashboard's name, space, or widgets. Unset
// fields stay unchanged.
type UpdateDashboardInput struct {
	DashboardID string          `json:"dashboardId" jsonschema:"id of the dashboard to update"`
	Name        string          `json:"name,omitempty" jsonschema:"new dashboard name"`
	SpaceID     string          `json:"spaceId,omitempty" jsonschema:"new space id to scope the dashboard to"`
	Payload     json.RawMessage `json:"payload,omitempty" jsonschema:"new ordered widget array"`
	DryRun      bool            `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// DeleteDashboardInput deletes a dashboard.
type DeleteDashboardInput struct {
	DashboardID string `json:"dashboardId" jsonschema:"id of the dashboard to delete"`
	DryRun      bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

func (s *Server) registerDashboards(srv *mcp.Server) {
	addTool(s, srv, &mcp.Tool{
		Name:        "list_dashboards",
		Description: "List dashboards for the token's team.",
	}, s.listDashboards, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "create_dashboard",
		Description: "Create a dashboard, optionally scoped to a space.",
	}, s.createDashboard, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "get_dashboard",
		Description: "Fetch a single dashboard by id, including its widgets.",
	}, s.getDashboard, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "update_dashboard",
		Description: "Update a dashboard's name, space, or widget layout.",
	}, s.updateDashboard, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "delete_dashboard",
		Description: "Delete a dashboard by id.",
	}, s.deleteDashboard, true)
}

func (s *Server) listDashboards(ctx context.Context, _ *mcp.CallToolRequest, in ListDashboardsInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListDashboards(ctx, in.Page, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) createDashboard(ctx context.Context, _ *mcp.CallToolRequest, in CreateDashboardInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	var spaceID *string
	if in.SpaceID != "" {
		spaceID = &in.SpaceID
	}
	d, err := c.CreateDashboard(ctx, operator.CreateDashboardInput{
		Name:    in.Name,
		SpaceID: spaceID,
		Payload: in.Payload,
	})
	if err != nil {
		return fail(err)
	}
	return ok(d)
}

func (s *Server) getDashboard(ctx context.Context, _ *mcp.CallToolRequest, in GetDashboardInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	d, err := c.GetDashboard(ctx, in.DashboardID)
	if err != nil {
		return fail(err)
	}
	return ok(d)
}

func (s *Server) updateDashboard(ctx context.Context, _ *mcp.CallToolRequest, in UpdateDashboardInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	var spaceID *string
	if in.SpaceID != "" {
		spaceID = &in.SpaceID
	}
	if err := c.UpdateDashboard(ctx, in.DashboardID, operator.UpdateDashboardInput{
		Name:    in.Name,
		SpaceID: spaceID,
		Payload: in.Payload,
	}); err != nil {
		return fail(err)
	}
	return ok(nil)
}

func (s *Server) deleteDashboard(ctx context.Context, _ *mcp.CallToolRequest, in DeleteDashboardInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.DeleteDashboard(ctx, in.DashboardID); err != nil {
		return fail(err)
	}
	return ok(nil)
}

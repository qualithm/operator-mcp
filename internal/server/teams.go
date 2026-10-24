package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// ListTeamsInput lists the teams the token's member belongs to.
type ListTeamsInput struct {
	Page  int `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit int `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// CreateTeamInput creates a team. The platform assigns the initial name and
// the caller becomes its owner.
type CreateTeamInput struct {
	DryRun bool `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// GetTeamInput fetches a team: by default the one this server acts on.
type GetTeamInput struct {
	TeamID string `json:"teamId,omitempty" jsonschema:"id of the team to fetch; empty means the team this server acts on"`
}

// UpdateTeamInput renames a team.
type UpdateTeamInput struct {
	TeamID string `json:"teamId" jsonschema:"id of the team to rename"`
	Name   string `json:"name" jsonschema:"new team name"`
	DryRun bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// DeleteTeamInput deletes a team.
type DeleteTeamInput struct {
	TeamID string `json:"teamId" jsonschema:"id of the team to delete"`
	DryRun bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// GetTeamDeviceStateInput fetches the latest state of every device in a team.
type GetTeamDeviceStateInput struct {
	TeamID string `json:"teamId" jsonschema:"id of the team to read device state for"`
}

func (s *Server) registerTeams(srv *mcp.Server) {
	addTool(s, srv, &mcp.Tool{
		Name:        "list_teams",
		Description: "List the teams the token's member belongs to.",
	}, s.listTeams, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "create_team",
		Description: "Create a team with the caller as owner. The platform assigns the initial name; rename it with update_team.",
	}, s.createTeam, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "get_team",
		Description: "Fetch a team with your role and the permissions you hold in it, so you can check before acting. Without an id, it's the team this server acts on.",
	}, s.getTeam, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "update_team",
		Description: "Rename a team.",
	}, s.updateTeam, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "delete_team",
		Description: "Delete a team by id.",
	}, s.deleteTeam, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "get_team_device_state",
		Description: "Read the latest online/metrics snapshot of every device in a team.",
	}, s.getTeamDeviceState, false)
}

func (s *Server) listTeams(ctx context.Context, _ *mcp.CallToolRequest, in ListTeamsInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListTeams(ctx, in.Page, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) createTeam(ctx context.Context, _ *mcp.CallToolRequest, in CreateTeamInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	t, err := c.CreateTeam(ctx)
	if err != nil {
		return fail(err)
	}
	return ok(t)
}

func (s *Server) getTeam(ctx context.Context, _ *mcp.CallToolRequest, in GetTeamInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	teamID := in.TeamID
	if teamID == "" {
		if teamID, err = c.TeamID(ctx); err != nil {
			return fail(err)
		}
	}
	t, err := c.GetTeam(ctx, teamID)
	if err != nil {
		return fail(err)
	}
	return ok(t)
}

func (s *Server) updateTeam(ctx context.Context, _ *mcp.CallToolRequest, in UpdateTeamInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.UpdateTeam(ctx, in.TeamID, operator.UpdateTeamInput{Name: in.Name}); err != nil {
		return fail(err)
	}
	return ok(nil)
}

func (s *Server) deleteTeam(ctx context.Context, _ *mcp.CallToolRequest, in DeleteTeamInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.DeleteTeam(ctx, in.TeamID); err != nil {
		return fail(err)
	}
	return ok(nil)
}

func (s *Server) getTeamDeviceState(ctx context.Context, _ *mcp.CallToolRequest, in GetTeamDeviceStateInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	snaps, err := c.GetTeamDeviceState(ctx, in.TeamID)
	if err != nil {
		return fail(err)
	}
	return ok(snaps)
}

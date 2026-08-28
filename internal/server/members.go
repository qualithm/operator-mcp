package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// ListTeamMembersInput lists active members of a team.
type ListTeamMembersInput struct {
	TeamID string `json:"teamId" jsonschema:"id of the team to list members for"`
	Page   int    `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit  int    `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// AddTeamMemberInput activates the caller's own invitation to a team.
type AddTeamMemberInput struct {
	TeamID   string `json:"teamId" jsonschema:"id of the team the invite belongs to"`
	InviteID string `json:"inviteId" jsonschema:"id of the invite being accepted"`
	DryRun   bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// GetTeamMemberInput fetches a single team member.
type GetTeamMemberInput struct {
	TeamID   string `json:"teamId" jsonschema:"id of the team"`
	MemberID string `json:"memberId" jsonschema:"id of the member to fetch"`
}

// UpdateTeamMemberInput changes a member's role.
type UpdateTeamMemberInput struct {
	TeamID   string `json:"teamId" jsonschema:"id of the team"`
	MemberID string `json:"memberId" jsonschema:"id of the member to update"`
	Role     string `json:"role" jsonschema:"new role: owner, manager, or guest"`
	DryRun   bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// RemoveTeamMemberInput removes a member from a team.
type RemoveTeamMemberInput struct {
	TeamID   string `json:"teamId" jsonschema:"id of the team"`
	MemberID string `json:"memberId" jsonschema:"id of the member to remove"`
	DryRun   bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// ListTeamInvitesInput lists pending invitations to a team.
type ListTeamInvitesInput struct {
	TeamID string `json:"teamId" jsonschema:"id of the team to list invites for"`
	Page   int    `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit  int    `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// CreateTeamInviteInput invites an account to a team by email.
type CreateTeamInviteInput struct {
	TeamID string `json:"teamId" jsonschema:"id of the team to invite into"`
	Email  string `json:"email" jsonschema:"email address of the account to invite"`
	Role   string `json:"role" jsonschema:"role the invite grants: owner, manager, or guest"`
	DryRun bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// RevokeTeamInviteInput revokes a pending invitation.
type RevokeTeamInviteInput struct {
	TeamID   string `json:"teamId" jsonschema:"id of the team"`
	InviteID string `json:"inviteId" jsonschema:"id of the invite to revoke"`
	DryRun   bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

func (s *Server) registerMembers(srv *mcp.Server) {
	addTool(s, srv, &mcp.Tool{
		Name:        "list_team_members",
		Description: "List the active members of a team.",
	}, s.listTeamMembers, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "add_team_member",
		Description: "Accept the caller's invitation to a team, activating their membership.",
	}, s.addTeamMember, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "get_team_member",
		Description: "Fetch a single team member by id.",
	}, s.getTeamMember, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "update_team_member",
		Description: "Change a team member's role.",
	}, s.updateTeamMember, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "remove_team_member",
		Description: "Remove a member from a team.",
	}, s.removeTeamMember, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "list_team_invites",
		Description: "List pending invitations to a team.",
	}, s.listTeamInvites, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "create_team_invite",
		Description: "Invite an account to a team by email.",
	}, s.createTeamInvite, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "revoke_team_invite",
		Description: "Revoke a pending team invitation.",
	}, s.revokeTeamInvite, true)
}

func (s *Server) listTeamMembers(ctx context.Context, _ *mcp.CallToolRequest, in ListTeamMembersInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListTeamMembers(ctx, in.TeamID, in.Page, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) addTeamMember(ctx context.Context, _ *mcp.CallToolRequest, in AddTeamMemberInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	m, err := c.AddTeamMember(ctx, in.TeamID, operator.AddTeamMemberInput{InviteID: in.InviteID})
	if err != nil {
		return fail(err)
	}
	return ok(m)
}

func (s *Server) getTeamMember(ctx context.Context, _ *mcp.CallToolRequest, in GetTeamMemberInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	m, err := c.GetTeamMember(ctx, in.TeamID, in.MemberID)
	if err != nil {
		return fail(err)
	}
	return ok(m)
}

func (s *Server) updateTeamMember(ctx context.Context, _ *mcp.CallToolRequest, in UpdateTeamMemberInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.UpdateTeamMember(ctx, in.TeamID, in.MemberID, operator.UpdateTeamMemberInput{Role: in.Role}); err != nil {
		return fail(err)
	}
	return ok(nil)
}

func (s *Server) removeTeamMember(ctx context.Context, _ *mcp.CallToolRequest, in RemoveTeamMemberInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.RemoveTeamMember(ctx, in.TeamID, in.MemberID); err != nil {
		return fail(err)
	}
	return ok(nil)
}

func (s *Server) listTeamInvites(ctx context.Context, _ *mcp.CallToolRequest, in ListTeamInvitesInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListTeamInvites(ctx, in.TeamID, in.Page, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) createTeamInvite(ctx context.Context, _ *mcp.CallToolRequest, in CreateTeamInviteInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	m, err := c.CreateTeamInvite(ctx, in.TeamID, operator.CreateTeamInviteInput{Email: in.Email, Role: in.Role})
	if err != nil {
		return fail(err)
	}
	return ok(m)
}

func (s *Server) revokeTeamInvite(ctx context.Context, _ *mcp.CallToolRequest, in RevokeTeamInviteInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.RevokeTeamInvite(ctx, in.TeamID, in.InviteID); err != nil {
		return fail(err)
	}
	return ok(nil)
}

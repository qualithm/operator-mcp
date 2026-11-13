package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// ListTeamMembersInput lists active members of a team.
type ListTeamMembersInput struct {
	TeamID string `json:"teamId" jsonschema:"id of the team to list members for"`
	Cursor string `json:"cursor,omitempty" jsonschema:"the nextCursor from the previous page; empty for the first page"`
	Limit  int    `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// AddTeamMemberInput accepts an invite addressed to the caller's email.
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
	Role     string `json:"role" jsonschema:"new role: owner, manager, or viewer"`
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
	Cursor string `json:"cursor,omitempty" jsonschema:"the nextCursor from the previous page; empty for the first page"`
	Limit  int    `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// CreateTeamInviteInput invites email addresses to a team.
type CreateTeamInviteInput struct {
	TeamID string   `json:"teamId" jsonschema:"id of the team to invite into"`
	Emails []string `json:"emails" jsonschema:"addresses to invite, with or without an account; up to 20"`
	Role   string   `json:"role" jsonschema:"role every invite grants: owner, manager, or viewer"`
	DryRun bool     `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// ResendTeamInviteInput mails a pending invite again with a new link.
type ResendTeamInviteInput struct {
	TeamID   string `json:"teamId" jsonschema:"id of the team"`
	InviteID string `json:"inviteId" jsonschema:"id of the invite to resend"`
	DryRun   bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// RevokeTeamInviteInput withdraws a pending invite.
type RevokeTeamInviteInput struct {
	TeamID   string `json:"teamId" jsonschema:"id of the team"`
	InviteID string `json:"inviteId" jsonschema:"id of the invite to withdraw"`
	DryRun   bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// LookupInviteInput reads the invite a link token names.
type LookupInviteInput struct {
	Token string `json:"token" jsonschema:"the token from the invite link, the part after /invites/"`
}

// AcceptInviteInput accepts the invite a link token names.
type AcceptInviteInput struct {
	Token  string `json:"token" jsonschema:"the token from the invite link, the part after /invites/"`
	DryRun bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

func (s *Server) registerMembers(srv *mcp.Server) {
	addTool(s, srv, &mcp.Tool{
		Name:        "list_team_members",
		Description: "List the active members of a team. Pass the result's nextCursor back as cursor for the next page; it is null on the last page.",
	}, s.listTeamMembers, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "add_team_member",
		Description: "Accept an invite by id. Only the account whose email the invite names can accept it.",
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
		Description: "List pending invites to a team. Pass the result's nextCursor back as cursor for the next page; it is null on the last page.",
	}, s.listTeamInvites, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "create_team_invite",
		Description: "Invite email addresses to a team with one role. Each address gets a result: sent, already_member, or already_invited (resend that one instead). Each link expires after 7 days, and the team can send a limited number of invite emails per hour.",
	}, s.createTeamInvite, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "resend_team_invite",
		Description: "Mail a pending invite again with a new link that expires after 7 days; the old link stops working.",
	}, s.resendTeamInvite, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "revoke_team_invite",
		Description: "Withdraw a pending team invite; its link stops working.",
	}, s.revokeTeamInvite, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "lookup_invite",
		Description: "Read the invite an invite link names: the team, role, inviter and expiry, and forCaller, whether the signed-in account may accept it.",
	}, s.lookupInvite, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "accept_invite",
		Description: "Accept the invite an invite link names, joining its team. Only the account whose email the invite names can accept it.",
	}, s.acceptInvite, true)
}

func (s *Server) listTeamMembers(ctx context.Context, _ *mcp.CallToolRequest, in ListTeamMembersInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListTeamMembers(ctx, in.TeamID, in.Cursor, in.Limit)
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
	page, err := c.ListTeamInvites(ctx, in.TeamID, in.Cursor, in.Limit)
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
	res, err := c.CreateTeamInvite(ctx, in.TeamID, operator.CreateTeamInviteInput{Emails: in.Emails, Role: in.Role})
	if err != nil {
		return fail(err)
	}
	return ok(res)
}

func (s *Server) resendTeamInvite(ctx context.Context, _ *mcp.CallToolRequest, in ResendTeamInviteInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	inv, err := c.ResendTeamInvite(ctx, in.TeamID, in.InviteID)
	if err != nil {
		return fail(err)
	}
	return ok(inv)
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

func (s *Server) lookupInvite(ctx context.Context, _ *mcp.CallToolRequest, in LookupInviteInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	link, err := c.LookupInvite(ctx, in.Token)
	if err != nil {
		return fail(err)
	}
	return ok(link)
}

func (s *Server) acceptInvite(ctx context.Context, _ *mcp.CallToolRequest, in AcceptInviteInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	m, err := c.AcceptInvite(ctx, in.Token)
	if err != nil {
		return fail(err)
	}
	return ok(m)
}

package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// GetWorkspaceInput reads the caller's current workspace context.
type GetWorkspaceInput struct{}

// GetAccountInput reads the caller's account.
type GetAccountInput struct{}

// ListCapabilitiesInput lists device capabilities across the team.
type ListCapabilitiesInput struct {
	Page  int    `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit int    `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
	Type  string `json:"type,omitempty" jsonschema:"filter by capability type: onoff, range, enum, trigger, or sensor"`
	Tag   string `json:"tag,omitempty" jsonschema:"filter by capability tag"`
	Key   string `json:"key,omitempty" jsonschema:"filter by capability key"`
}

// ListRolesInput lists the caller's team role assignments.
type ListRolesInput struct {
	Page  int `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit int `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// ListSessionsInput lists the caller's authenticated sessions.
type ListSessionsInput struct {
	Page           int  `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit          int  `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
	ThisDeviceOnly bool `json:"thisDeviceOnly,omitempty" jsonschema:"list only the session the token belongs to"`
}

// GetSessionInput fetches a single session.
type GetSessionInput struct {
	SessionID string `json:"sessionId" jsonschema:"id of the session to fetch"`
}

// GetCommunicationPreferencesInput reads the caller's notification
// preferences.
type GetCommunicationPreferencesInput struct{}

// ListZoneSpacesInput lists spaces in a device zone.
type ListZoneSpacesInput struct {
	Zone  string `json:"zone" jsonschema:"device zone to list spaces in"`
	Page  int    `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit int    `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

func (s *Server) registerWorkspace(srv *mcp.Server) {
	addTool(s, srv, &mcp.Tool{
		Name:        "get_workspace",
		Description: "Read the caller's current workspace: team, membership, and role context.",
	}, s.getWorkspace, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "get_account",
		Description: "Read the caller's account.",
	}, s.getAccount, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "list_capabilities",
		Description: "List device capabilities across the team, optionally filtered by type, tag, or key.",
	}, s.listCapabilities, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "list_roles",
		Description: "List the caller's team role assignments.",
	}, s.listRoles, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "list_sessions",
		Description: "List the caller's authenticated sessions.",
	}, s.listSessions, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "get_session",
		Description: "Fetch a single session by id.",
	}, s.getSession, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "get_communication_preferences",
		Description: "Read the caller's email/push notification preferences.",
	}, s.getCommunicationPreferences, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "list_zone_spaces",
		Description: "List spaces in a device zone.",
	}, s.listZoneSpaces, false)
}

func (s *Server) getWorkspace(ctx context.Context, _ *mcp.CallToolRequest, _ GetWorkspaceInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	ws, err := c.GetWorkspace(ctx)
	if err != nil {
		return fail(err)
	}
	return ok(ws)
}

func (s *Server) getAccount(ctx context.Context, _ *mcp.CallToolRequest, _ GetAccountInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	acct, err := c.GetAccount(ctx)
	if err != nil {
		return fail(err)
	}
	return ok(acct)
}

func (s *Server) listCapabilities(ctx context.Context, _ *mcp.CallToolRequest, in ListCapabilitiesInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListCapabilities(ctx, operator.ListCapabilitiesInput{
		Page:  in.Page,
		Limit: in.Limit,
		Type:  in.Type,
		Tag:   in.Tag,
		Key:   in.Key,
	})
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) listRoles(ctx context.Context, _ *mcp.CallToolRequest, in ListRolesInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListRoles(ctx, in.Page, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) listSessions(ctx context.Context, _ *mcp.CallToolRequest, in ListSessionsInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListSessions(ctx, in.Page, in.Limit, in.ThisDeviceOnly)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) getSession(ctx context.Context, _ *mcp.CallToolRequest, in GetSessionInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	sess, err := c.GetSession(ctx, in.SessionID)
	if err != nil {
		return fail(err)
	}
	return ok(sess)
}

func (s *Server) getCommunicationPreferences(ctx context.Context, _ *mcp.CallToolRequest, _ GetCommunicationPreferencesInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	prefs, err := c.GetCommunicationPreferences(ctx)
	if err != nil {
		return fail(err)
	}
	return ok(prefs)
}

func (s *Server) listZoneSpaces(ctx context.Context, _ *mcp.CallToolRequest, in ListZoneSpacesInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListZoneSpaces(ctx, in.Zone, in.Page, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

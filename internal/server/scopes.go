package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// ListScopesInput lists the team's scopes, optionally in one zone.
type ListScopesInput struct {
	Zone   string `json:"zone,omitempty" jsonschema:"only the scopes in this device zone, e.g. de-fra-a"`
	Cursor string `json:"cursor,omitempty" jsonschema:"the nextCursor from the previous page; empty for the first page"`
	Limit  int    `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// GetScopeInput fetches a single scope.
type GetScopeInput struct {
	ScopeID string `json:"scopeId" jsonschema:"id of the scope to fetch"`
}

// CreateScopeInput creates a scope: top-level in a device zone, or under a
// parent scope, whose zone it inherits. The platform assigns the initial name;
// rename with update_scope.
//
// A zone unavailable in the current environment returns the `auth` failure
// with message `Zone rejects creation in this environment` — treat it as a
// signaling condition and pick an open zone, not as a fault.
type CreateScopeInput struct {
	Zone     string `json:"zone,omitempty" jsonschema:"device zone for a top-level scope; omit when parentId is set"`
	ParentID string `json:"parentId,omitempty" jsonschema:"id of the scope to nest the new scope under; omit for a top-level scope"`
	DryRun   bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// UpdateScopeInput renames a scope.
type UpdateScopeInput struct {
	ScopeID string `json:"scopeId" jsonschema:"id of the scope to rename"`
	Name    string `json:"name" jsonschema:"new scope name"`
	DryRun  bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// DeleteScopeInput deletes an empty scope.
type DeleteScopeInput struct {
	ScopeID string `json:"scopeId" jsonschema:"id of the scope to delete"`
	DryRun  bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

func (s *Server) registerScopes(srv *mcp.Server) {
	addTool(s, srv, &mcp.Tool{
		Name:        "list_scopes",
		Description: "List the team's scopes, optionally only those in one device zone. Pass the result's nextCursor back as cursor for the next page; it is null on the last page.",
	}, s.listScopes, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "get_scope",
		Description: "Fetch a single scope by id.",
	}, s.getScope, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "create_scope",
		Description: "Create a scope: top-level in a device zone, or nested under a parent scope.",
	}, s.createScope, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "update_scope",
		Description: "Rename a scope.",
	}, s.updateScope, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "delete_scope",
		Description: "Delete an empty scope by id. The platform refuses it with a conflict while the scope still holds devices or scopes.",
	}, s.deleteScope, true)
}

func (s *Server) listScopes(ctx context.Context, _ *mcp.CallToolRequest, in ListScopesInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	var page operator.Page[operator.Scope]
	if in.Zone != "" {
		page, err = c.ListZoneScopes(ctx, in.Zone, in.Cursor, in.Limit)
	} else {
		page, err = c.ListScopes(ctx, in.Cursor, in.Limit)
	}
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) getScope(ctx context.Context, _ *mcp.CallToolRequest, in GetScopeInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	scope, err := c.GetScope(ctx, in.ScopeID)
	if err != nil {
		return fail(err)
	}
	return ok(scope)
}

func (s *Server) createScope(ctx context.Context, _ *mcp.CallToolRequest, in CreateScopeInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	scope, err := c.CreateScope(ctx, operator.CreateScopeInput{Zone: in.Zone, ParentID: in.ParentID})
	if err != nil {
		return fail(err)
	}
	return ok(scope)
}

func (s *Server) updateScope(ctx context.Context, _ *mcp.CallToolRequest, in UpdateScopeInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.UpdateScope(ctx, in.ScopeID, in.Name); err != nil {
		return fail(err)
	}
	return ok(nil)
}

func (s *Server) deleteScope(ctx context.Context, _ *mcp.CallToolRequest, in DeleteScopeInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.DeleteScope(ctx, in.ScopeID); err != nil {
		return fail(err)
	}
	return ok(nil)
}

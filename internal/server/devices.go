package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// ListDevicesInput lists the team's devices, optionally only those directly in one scope.
type ListDevicesInput struct {
	ScopeID string `json:"scopeId,omitempty" jsonschema:"only the devices directly in this scope, not in the scopes under it"`
	Cursor  string `json:"cursor,omitempty" jsonschema:"the nextCursor from the previous page; empty for the first page"`
	Limit   int    `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// GetDeviceInput fetches a single device.
type GetDeviceInput struct {
	DeviceID string `json:"deviceId" jsonschema:"id of the device to fetch"`
}

// ListScopeDevicesInput lists the devices in a scope and every scope under it.
type ListScopeDevicesInput struct {
	ScopeID string `json:"scopeId" jsonschema:"id of the scope whose subtree to list"`
	Cursor  string `json:"cursor,omitempty" jsonschema:"the nextCursor from the previous page; empty for the first page"`
	Limit   int    `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// CreateDeviceInput creates a device in a scope.
type CreateDeviceInput struct {
	ScopeID string `json:"scopeId" jsonschema:"id of the scope to create the device in"`
	DryRun  bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// UpdateDeviceInput updates a device. At least one of name, scopeId or parked must be set.
type UpdateDeviceInput struct {
	DeviceID string `json:"deviceId" jsonschema:"id of the device to update"`
	Name     string `json:"name,omitempty" jsonschema:"new device name"`
	ScopeID  string `json:"scopeId,omitempty" jsonschema:"id of the scope to move the device into"`
	Parked   *bool  `json:"parked,omitempty" jsonschema:"true to park the device: it stops connecting and counting against the active-device limit, while its credentials and history survive; false to return it to active"`
	DryRun   bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// DeleteDeviceInput deletes a device.
type DeleteDeviceInput struct {
	DeviceID string `json:"deviceId" jsonschema:"id of the device to delete"`
	DryRun   bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

func (s *Server) registerDevices(srv *mcp.Server) {
	addTool(s, srv, &mcp.Tool{
		Name:        "list_devices",
		Description: "List the team's devices, optionally only those directly in one scope; list_scope_devices includes the scopes under it. Pass the result's nextCursor back as cursor for the next page; it is null on the last page.",
	}, s.listDevices, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "list_scope_devices",
		Description: "List the devices in a scope and every scope under it. Pass the result's nextCursor back as cursor for the next page; it is null on the last page.",
	}, s.listScopeDevices, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "get_device",
		Description: "Fetch a single device by id.",
	}, s.getDevice, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "create_device",
		Description: "Create a device in a scope.",
	}, s.createDevice, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "update_device",
		Description: "Update a device's name or scope, or park or unpark it.",
	}, s.updateDevice, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "delete_device",
		Description: "Delete a device by id.",
	}, s.deleteDevice, true)
}

func (s *Server) listDevices(ctx context.Context, _ *mcp.CallToolRequest, in ListDevicesInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	var page operator.Page[operator.Device]
	if in.ScopeID != "" {
		page, err = c.ListScopeDevices(ctx, in.ScopeID, in.Cursor, in.Limit)
	} else {
		page, err = c.ListDevices(ctx, in.Cursor, in.Limit)
	}
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) listScopeDevices(ctx context.Context, _ *mcp.CallToolRequest, in ListScopeDevicesInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListScopeSubtreeDevices(ctx, in.ScopeID, in.Cursor, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) getDevice(ctx context.Context, _ *mcp.CallToolRequest, in GetDeviceInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	d, err := c.GetDevice(ctx, in.DeviceID)
	if err != nil {
		return fail(err)
	}
	return ok(d)
}

func (s *Server) createDevice(ctx context.Context, _ *mcp.CallToolRequest, in CreateDeviceInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	d, err := c.CreateDevice(ctx, operator.CreateDeviceInput{ScopeID: in.ScopeID})
	if err != nil {
		return fail(err)
	}
	return ok(d)
}

func (s *Server) updateDevice(ctx context.Context, _ *mcp.CallToolRequest, in UpdateDeviceInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.UpdateDevice(ctx, in.DeviceID, operator.UpdateDeviceInput{
		Name:    in.Name,
		ScopeID: in.ScopeID,
		Parked:  in.Parked,
	}); err != nil {
		return fail(err)
	}
	return ok(nil)
}

func (s *Server) deleteDevice(ctx context.Context, _ *mcp.CallToolRequest, in DeleteDeviceInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.DeleteDevice(ctx, in.DeviceID); err != nil {
		return fail(err)
	}
	return ok(nil)
}

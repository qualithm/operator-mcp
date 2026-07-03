package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// ListDevicesInput lists devices for the token's team.
type ListDevicesInput struct {
	Page  int `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit int `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// ListSpaceDevicesInput lists devices scoped to a space.
type ListSpaceDevicesInput struct {
	SpaceID string `json:"spaceId" jsonschema:"id of the space to list devices for"`
	Page    int    `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit   int    `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// GetDeviceInput fetches a single device.
type GetDeviceInput struct {
	DeviceID string `json:"deviceId" jsonschema:"id of the device to fetch"`
}

// CreateDeviceInput creates a device in a space.
type CreateDeviceInput struct {
	SpaceID string `json:"spaceId" jsonschema:"id of the space to create the device in"`
	DryRun  bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// UpdateDeviceInput updates a device. At least one of name or spaceId must be set.
type UpdateDeviceInput struct {
	DeviceID string `json:"deviceId" jsonschema:"id of the device to update"`
	Name     string `json:"name,omitempty" jsonschema:"new device name"`
	SpaceID  string `json:"spaceId,omitempty" jsonschema:"new space id to move the device into"`
	DryRun   bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// DeleteDeviceInput deletes a device.
type DeleteDeviceInput struct {
	DeviceID string `json:"deviceId" jsonschema:"id of the device to delete"`
	DryRun   bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

func (s *Server) registerDevices(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_devices",
		Description: "List devices for the token's team.",
	}, s.listDevices)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_space_devices",
		Description: "List devices scoped to a single space.",
	}, s.listSpaceDevices)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_device",
		Description: "Fetch a single device by id.",
	}, s.getDevice)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_device",
		Description: "Create a device in a space.",
	}, s.createDevice)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_device",
		Description: "Update a device's name or space.",
	}, s.updateDevice)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "delete_device",
		Description: "Delete a device by id.",
	}, s.deleteDevice)
}

func (s *Server) listDevices(ctx context.Context, _ *mcp.CallToolRequest, in ListDevicesInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListDevices(ctx, in.Page, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) listSpaceDevices(ctx context.Context, _ *mcp.CallToolRequest, in ListSpaceDevicesInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListSpaceDevices(ctx, in.SpaceID, in.Page, in.Limit)
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
	d, err := c.CreateDevice(ctx, operator.CreateDeviceInput{SpaceID: in.SpaceID})
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
		SpaceID: in.SpaceID,
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

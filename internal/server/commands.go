package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// ListDeviceCommandsInput lists command deliveries for a device.
type ListDeviceCommandsInput struct {
	DeviceID string `json:"deviceId" jsonschema:"id of the device to list commands for"`
	Page     int    `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit    int    `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// SendDeviceCommandInput queues a command for a device. dedupKey is
// caller-owned: a retried send with the same key does not command the device
// twice.
type SendDeviceCommandInput struct {
	DeviceID   string `json:"deviceId" jsonschema:"id of the device to command"`
	Capability string `json:"capability" jsonschema:"key of the declared capability to invoke"`
	Value      any    `json:"value,omitempty" jsonschema:"command value; required for every commandable capability type except trigger"`
	DedupKey   string `json:"dedupKey" jsonschema:"caller-owned idempotency key for the command"`
	DryRun     bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// GetDeviceCapabilitiesInput lists a device's declared capabilities.
type GetDeviceCapabilitiesInput struct {
	DeviceID string `json:"deviceId" jsonschema:"id of the device to list capabilities for"`
}

// ParkDeviceInput parks a device.
type ParkDeviceInput struct {
	DeviceID string `json:"deviceId" jsonschema:"id of the device to park"`
	DryRun   bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// UnparkDeviceInput returns a parked device to active.
type UnparkDeviceInput struct {
	DeviceID string `json:"deviceId" jsonschema:"id of the device to unpark"`
	DryRun   bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

func (s *Server) registerCommands(srv *mcp.Server) {
	addTool(s, srv, &mcp.Tool{
		Name:        "list_device_commands",
		Description: "List command deliveries for a device (pending, sent, failed).",
	}, s.listDeviceCommands, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "send_device_command",
		Description: "Queue a command for a device, idempotently via dedupKey.",
	}, s.sendDeviceCommand, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "get_device_capabilities",
		Description: "List the capabilities a device declared in its connect-time manifest.",
	}, s.getDeviceCapabilities, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "park_device",
		Description: "Park a device: it stops counting against the active-device limit while its credentials and history survive. Idempotent.",
	}, s.parkDevice, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "unpark_device",
		Description: "Return a parked device to active. Idempotent.",
	}, s.unparkDevice, true)
}

func (s *Server) listDeviceCommands(ctx context.Context, _ *mcp.CallToolRequest, in ListDeviceCommandsInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListDeviceCommands(ctx, in.DeviceID, in.Page, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) sendDeviceCommand(ctx context.Context, _ *mcp.CallToolRequest, in SendDeviceCommandInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	q, err := c.SendDeviceCommand(ctx, in.DeviceID, operator.SendDeviceCommandInput{
		Capability: in.Capability,
		Value:      in.Value,
		DedupKey:   in.DedupKey,
	})
	if err != nil {
		return fail(err)
	}
	return ok(q)
}

func (s *Server) getDeviceCapabilities(ctx context.Context, _ *mcp.CallToolRequest, in GetDeviceCapabilitiesInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	caps, err := c.GetDeviceCapabilities(ctx, in.DeviceID)
	if err != nil {
		return fail(err)
	}
	return ok(caps)
}

func (s *Server) parkDevice(ctx context.Context, _ *mcp.CallToolRequest, in ParkDeviceInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.ParkDevice(ctx, in.DeviceID); err != nil {
		return fail(err)
	}
	return ok(nil)
}

func (s *Server) unparkDevice(ctx context.Context, _ *mcp.CallToolRequest, in UnparkDeviceInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.UnparkDevice(ctx, in.DeviceID); err != nil {
		return fail(err)
	}
	return ok(nil)
}

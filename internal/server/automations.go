package server

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	operator "github.com/qualithm/operator-go"
)

// ListAutomationsInput lists automations for the token's team.
type ListAutomationsInput struct {
	Page  int `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit int `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// CreateAutomationInput creates an automation from a full payload. Omit
// spaceId for a team-wide automation.
type CreateAutomationInput struct {
	Name    string          `json:"name" jsonschema:"automation name"`
	SpaceID string          `json:"spaceId,omitempty" jsonschema:"id of the space to scope the automation to; omit for team-wide"`
	Payload json.RawMessage `json:"payload" jsonschema:"automation document: trigger plus ordered condition/action chain"`
	DryRun  bool            `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// GetAutomationInput fetches a single automation.
type GetAutomationInput struct {
	AutomationID string `json:"automationId" jsonschema:"id of the automation to fetch"`
}

// UpdateAutomationInput updates an automation. template and payload are
// mutually exclusive: setting template re-expands the chain server-side.
type UpdateAutomationInput struct {
	AutomationID string          `json:"automationId" jsonschema:"id of the automation to update"`
	Name         string          `json:"name,omitempty" jsonschema:"new automation name"`
	SpaceID      string          `json:"spaceId,omitempty" jsonschema:"new space id to scope the automation to"`
	Payload      json.RawMessage `json:"payload,omitempty" jsonschema:"new trigger+chain document"`
	Template     string          `json:"template,omitempty" jsonschema:"template id to re-expand the chain from"`
	Params       json.RawMessage `json:"params,omitempty" jsonschema:"template slot values, required with template"`
	DryRun       bool            `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// DeleteAutomationInput deletes an automation.
type DeleteAutomationInput struct {
	AutomationID string `json:"automationId" jsonschema:"id of the automation to delete"`
	DryRun       bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// EnableAutomationInput enables an automation.
type EnableAutomationInput struct {
	AutomationID string `json:"automationId" jsonschema:"id of the automation to enable"`
	DryRun       bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// DisableAutomationInput disables an automation.
type DisableAutomationInput struct {
	AutomationID string `json:"automationId" jsonschema:"id of the automation to disable"`
	DryRun       bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// RunAutomationInput queues a manual run of an automation.
type RunAutomationInput struct {
	AutomationID string `json:"automationId" jsonschema:"id of the automation to run"`
	DryRun       bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// ListAutomationRunsInput lists execution summaries for an automation.
type ListAutomationRunsInput struct {
	AutomationID string `json:"automationId" jsonschema:"id of the automation to list runs for"`
	Page         int    `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit        int    `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// CreateAutomationFromTemplateInput creates an automation by expanding a
// catalogue template server-side.
type CreateAutomationFromTemplateInput struct {
	Name     string          `json:"name" jsonschema:"automation name"`
	Template string          `json:"template" jsonschema:"id of the catalogue template to expand"`
	Params   json.RawMessage `json:"params,omitempty" jsonschema:"template slot values"`
	DryRun   bool            `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// ListAutomationTemplatesInput lists the template catalogue for a device.
type ListAutomationTemplatesInput struct {
	DeviceID string `json:"deviceId" jsonschema:"id of the device whose declared capabilities filter the catalogue"`
}

// TriggerAutomationInput fires an automation's webhook trigger. secret is the
// trigger secret issued by create_automation_trigger_secret, not the member
// API token.
type TriggerAutomationInput struct {
	Secret  string          `json:"secret" jsonschema:"webhook trigger secret for the automation"`
	Context json.RawMessage `json:"context,omitempty" jsonschema:"optional context object passed to the chain"`
	DryRun  bool            `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

// CreateAutomationTriggerSecretInput issues a webhook trigger secret for an
// automation.
type CreateAutomationTriggerSecretInput struct {
	AutomationID string `json:"automationId" jsonschema:"id of the automation to issue a trigger secret for"`
	DryRun       bool   `json:"dryRun,omitempty" jsonschema:"plan the mutation without applying it"`
}

func (s *Server) registerAutomations(srv *mcp.Server) {
	addTool(s, srv, &mcp.Tool{
		Name:        "list_automations",
		Description: "List automations for the token's team.",
	}, s.listAutomations, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "create_automation",
		Description: "Create an automation from a full trigger+chain payload.",
	}, s.createAutomation, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "get_automation",
		Description: "Fetch a single automation by id, including its chain.",
	}, s.getAutomation, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "update_automation",
		Description: "Update an automation's name, space, chain, or template params.",
	}, s.updateAutomation, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "delete_automation",
		Description: "Delete an automation by id.",
	}, s.deleteAutomation, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "enable_automation",
		Description: "Enable an automation so its trigger fires.",
	}, s.enableAutomation, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "disable_automation",
		Description: "Disable an automation without deleting it.",
	}, s.disableAutomation, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "run_automation",
		Description: "Queue a manual run of an automation.",
	}, s.runAutomation, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "list_automation_runs",
		Description: "List execution summaries for an automation.",
	}, s.listAutomationRuns, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "create_automation_from_template",
		Description: "Create an automation by expanding a catalogue template with slot values.",
	}, s.createAutomationFromTemplate, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "list_automation_templates",
		Description: "List the automation template catalogue filtered to what a device declared it can do.",
	}, s.listAutomationTemplates, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "trigger_automation",
		Description: "Fire an automation's webhook trigger, authenticated by its trigger secret.",
	}, s.triggerAutomation, true)
	addTool(s, srv, &mcp.Tool{
		Name:        "create_automation_trigger_secret",
		Description: "Issue a webhook trigger secret for an automation. The URL and secret are returned once.",
	}, s.createAutomationTriggerSecret, true)
}

func (s *Server) listAutomations(ctx context.Context, _ *mcp.CallToolRequest, in ListAutomationsInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListAutomations(ctx, in.Page, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) createAutomation(ctx context.Context, _ *mcp.CallToolRequest, in CreateAutomationInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	var spaceID *string
	if in.SpaceID != "" {
		spaceID = &in.SpaceID
	}
	var payload operator.AutomationPayload
	if len(in.Payload) > 0 {
		if err := json.Unmarshal(in.Payload, &payload); err != nil {
			return fail(err)
		}
	}
	a, err := c.CreateAutomation(ctx, operator.CreateAutomationInput{
		Name:    in.Name,
		SpaceID: spaceID,
		Payload: payload,
	})
	if err != nil {
		return fail(err)
	}
	return ok(a)
}

func (s *Server) getAutomation(ctx context.Context, _ *mcp.CallToolRequest, in GetAutomationInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	a, err := c.GetAutomation(ctx, in.AutomationID)
	if err != nil {
		return fail(err)
	}
	return ok(a)
}

func (s *Server) updateAutomation(ctx context.Context, _ *mcp.CallToolRequest, in UpdateAutomationInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	var spaceID *string
	if in.SpaceID != "" {
		spaceID = &in.SpaceID
	}
	upd := operator.UpdateAutomationInput{
		Name:     in.Name,
		SpaceID:  spaceID,
		Template: in.Template,
	}
	if len(in.Payload) > 0 {
		var payload operator.AutomationPayload
		if err := json.Unmarshal(in.Payload, &payload); err != nil {
			return fail(err)
		}
		upd.Payload = &payload
	}
	if len(in.Params) > 0 {
		var params map[string]any
		if err := json.Unmarshal(in.Params, &params); err != nil {
			return fail(err)
		}
		upd.Params = params
	}
	if err := c.UpdateAutomation(ctx, in.AutomationID, upd); err != nil {
		return fail(err)
	}
	return ok(nil)
}

func (s *Server) deleteAutomation(ctx context.Context, _ *mcp.CallToolRequest, in DeleteAutomationInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	if err := c.DeleteAutomation(ctx, in.AutomationID); err != nil {
		return fail(err)
	}
	return ok(nil)
}

func (s *Server) enableAutomation(ctx context.Context, _ *mcp.CallToolRequest, in EnableAutomationInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	a, err := c.EnableAutomation(ctx, in.AutomationID)
	if err != nil {
		return fail(err)
	}
	return ok(a)
}

func (s *Server) disableAutomation(ctx context.Context, _ *mcp.CallToolRequest, in DisableAutomationInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	a, err := c.DisableAutomation(ctx, in.AutomationID)
	if err != nil {
		return fail(err)
	}
	return ok(a)
}

func (s *Server) runAutomation(ctx context.Context, _ *mcp.CallToolRequest, in RunAutomationInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	run, err := c.RunAutomation(ctx, in.AutomationID)
	if err != nil {
		return fail(err)
	}
	return ok(run)
}

func (s *Server) listAutomationRuns(ctx context.Context, _ *mcp.CallToolRequest, in ListAutomationRunsInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListAutomationRuns(ctx, in.AutomationID, in.Page, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) createAutomationFromTemplate(ctx context.Context, _ *mcp.CallToolRequest, in CreateAutomationFromTemplateInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	var params map[string]any
	if len(in.Params) > 0 {
		if err := json.Unmarshal(in.Params, &params); err != nil {
			return fail(err)
		}
	}
	a, err := c.CreateAutomationFromTemplate(ctx, operator.CreateAutomationFromTemplateInput{
		Name:     in.Name,
		Template: in.Template,
		Params:   params,
	})
	if err != nil {
		return fail(err)
	}
	return ok(a)
}

func (s *Server) listAutomationTemplates(ctx context.Context, _ *mcp.CallToolRequest, in ListAutomationTemplatesInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	tmpls, err := c.ListAutomationTemplates(ctx, in.DeviceID)
	if err != nil {
		return fail(err)
	}
	return ok(tmpls)
}

func (s *Server) triggerAutomation(ctx context.Context, _ *mcp.CallToolRequest, in TriggerAutomationInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	var triggerContext map[string]any
	if len(in.Context) > 0 {
		if err := json.Unmarshal(in.Context, &triggerContext); err != nil {
			return fail(err)
		}
	}
	accepted, err := c.TriggerAutomation(ctx, in.Secret, triggerContext)
	if err != nil {
		return fail(err)
	}
	return ok(accepted)
}

func (s *Server) createAutomationTriggerSecret(ctx context.Context, _ *mcp.CallToolRequest, in CreateAutomationTriggerSecretInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(in.DryRun)
	if err != nil {
		return fail(err)
	}
	issued, err := c.CreateAutomationTriggerSecret(ctx, in.AutomationID)
	if err != nil {
		return fail(err)
	}
	return ok(issued)
}

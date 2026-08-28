package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetBillingSummaryInput reads the team's billing state.
type GetBillingSummaryInput struct{}

// ListInvoicesInput lists the team's invoices.
type ListInvoicesInput struct {
	Page  int `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	Limit int `json:"limit,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// PreviewTierChangeInput prices a tier change without applying it.
type PreviewTierChangeInput struct {
	Tier string `json:"tier" jsonschema:"target paid tier to price"`
}

// Billing tools are read-only by decision: money-moving operations (tier
// changes, add-ons, checkout and portal sessions) stay human-only.
func (s *Server) registerBilling(srv *mcp.Server) {
	addTool(s, srv, &mcp.Tool{
		Name:        "get_billing_summary",
		Description: "Read the team's billing state for the current month: tier, limits, usage per dimension, and add-ons.",
	}, s.getBillingSummary, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "list_invoices",
		Description: "List the team's invoices.",
	}, s.listInvoices, false)
	addTool(s, srv, &mcp.Tool{
		Name:        "preview_tier_change",
		Description: "Price a move to a paid tier without applying it. Read-only.",
	}, s.previewTierChange, false)
}

func (s *Server) getBillingSummary(ctx context.Context, _ *mcp.CallToolRequest, _ GetBillingSummaryInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	sum, err := c.GetBillingSummary(ctx)
	if err != nil {
		return fail(err)
	}
	return ok(sum)
}

func (s *Server) listInvoices(ctx context.Context, _ *mcp.CallToolRequest, in ListInvoicesInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	page, err := c.ListInvoices(ctx, in.Page, in.Limit)
	if err != nil {
		return fail(err)
	}
	return ok(page)
}

func (s *Server) previewTierChange(ctx context.Context, _ *mcp.CallToolRequest, in PreviewTierChangeInput) (*mcp.CallToolResult, Result, error) {
	c, err := s.newClient(false)
	if err != nil {
		return fail(err)
	}
	preview, err := c.PreviewTierChange(ctx, in.Tier)
	if err != nil {
		return fail(err)
	}
	return ok(preview)
}

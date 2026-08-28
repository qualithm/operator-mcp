package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestBillingTools(t *testing.T) {
	rec := &record{}
	s := testServer(t, 200, envelope(`{"tier":"pro","monthlyTotal":2900,"month":"2026-08","usage":{"devices":{"used":3,"limit":50}}}`), rec)
	if _, out, err := s.getBillingSummary(ctx(), nil, GetBillingSummaryInput{}); err != nil || !out.OK {
		t.Fatalf("getBillingSummary: %v %+v", err, out)
	}
	if rec.path != "/billing/summary" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"current":1,"items":[{"id":"inv_1","status":"paid"}],"last":1}`), rec)
	if _, out, err := s.listInvoices(ctx(), nil, ListInvoicesInput{}); err != nil || !out.OK {
		t.Fatalf("listInvoices: %v %+v", err, out)
	}
	if rec.path != "/invoices" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"from":"starter","to":"pro","amountDue":1200,"currency":"usd"}`), rec)
	if _, out, err := s.previewTierChange(ctx(), nil, PreviewTierChangeInput{Tier: "pro"}); err != nil || !out.OK {
		t.Fatalf("previewTierChange: %v %+v", err, out)
	}
	if rec.method != http.MethodPost || rec.path != "/billing/tier/preview" || !strings.Contains(rec.body, `"tier":"pro"`) {
		t.Fatalf("request %s %s body=%q", rec.method, rec.path, rec.body)
	}
}

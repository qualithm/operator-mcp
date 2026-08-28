package server

import (
	"strings"
	"testing"
)

func TestWorkspaceTools(t *testing.T) {
	rec := &record{}
	s := testServer(t, 200, envelope(`{"accountId":"acc_1","teamId":"team_1","memberRole":"owner"}`), rec)
	if _, out, err := s.getWorkspace(ctx(), nil, GetWorkspaceInput{}); err != nil || !out.OK {
		t.Fatalf("getWorkspace: %v %+v", err, out)
	}
	if rec.path != "/workspace" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"id":"acc_1","email":"a@b.c"}`), rec)
	if _, out, err := s.getAccount(ctx(), nil, GetAccountInput{}); err != nil || !out.OK {
		t.Fatalf("getAccount: %v %+v", err, out)
	}
	if rec.path != "/account" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"current":1,"items":[{"id":"cap_1","key":"power"}],"last":1}`), rec)
	if _, out, err := s.listCapabilities(ctx(), nil, ListCapabilitiesInput{Type: "onoff", Tag: "lamp", Key: "power"}); err != nil || !out.OK {
		t.Fatalf("listCapabilities: %v %+v", err, out)
	}
	if rec.path != "/capabilities" {
		t.Fatalf("path %q", rec.path)
	}
	for _, want := range []string{"type=onoff", "tag=lamp", "key=power"} {
		if !strings.Contains(rec.query, want) {
			t.Fatalf("query %q missing %q", rec.query, want)
		}
	}

	s = testServer(t, 200, envelope(`{"current":1,"items":[{"id":"mem_1","role":"owner"}],"last":1}`), rec)
	if _, out, err := s.listRoles(ctx(), nil, ListRolesInput{}); err != nil || !out.OK {
		t.Fatalf("listRoles: %v %+v", err, out)
	}
	if rec.path != "/roles" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"current":1,"items":[{"id":"ses_1","thisDevice":true}],"last":1}`), rec)
	if _, out, err := s.listSessions(ctx(), nil, ListSessionsInput{ThisDeviceOnly: true}); err != nil || !out.OK {
		t.Fatalf("listSessions: %v %+v", err, out)
	}
	if rec.path != "/sessions" || !strings.Contains(rec.query, "this_device=true") {
		t.Fatalf("path %q query %q", rec.path, rec.query)
	}

	s = testServer(t, 200, envelope(`{"id":"ses_1","userAgent":"agent"}`), rec)
	if _, out, err := s.getSession(ctx(), nil, GetSessionInput{SessionID: "ses_1"}); err != nil || !out.OK {
		t.Fatalf("getSession: %v %+v", err, out)
	}
	if rec.path != "/sessions/ses_1" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"productUpdates":true}`), rec)
	if _, out, err := s.getCommunicationPreferences(ctx(), nil, GetCommunicationPreferencesInput{}); err != nil || !out.OK {
		t.Fatalf("getCommunicationPreferences: %v %+v", err, out)
	}
	if rec.path != "/account/communication-preferences" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"current":1,"items":[{"id":"spc_1","zone":"us"}],"last":1}`), rec)
	if _, out, err := s.listZoneSpaces(ctx(), nil, ListZoneSpacesInput{Zone: "us"}); err != nil || !out.OK {
		t.Fatalf("listZoneSpaces: %v %+v", err, out)
	}
	if rec.path != "/zones/us/spaces" {
		t.Fatalf("path %q", rec.path)
	}
}

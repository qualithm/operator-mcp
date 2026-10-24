package server

import (
	"strings"
	"testing"

	operator "github.com/qualithm/operator-go"
)

func TestWorkspaceTools(t *testing.T) {
	rec := &record{}
	s := testServer(t, 200, envelope(`{"id":"team_1","memberRole":"viewer","permissions":["device.read"]}`), rec)
	// Without an id, get_team reads the team the server acts on.
	_, out, err := s.getTeam(ctx(), nil, GetTeamInput{})
	if err != nil || !out.OK {
		t.Fatalf("getTeam: %v %+v", err, out)
	}
	if rec.path != "/v1/teams/team_1" {
		t.Fatalf("path %q", rec.path)
	}
	// An agent reads what it may do before acting.
	if team, ok := out.Data.(operator.Team); !ok || len(team.Permissions) != 1 || team.Permissions[0] != "device.read" {
		t.Fatalf("team permissions: %+v", out.Data)
	}

	s = testServer(t, 200, envelope(`{"id":"acc_1","email":"a@b.c"}`), rec)
	if _, out, err := s.getAccount(ctx(), nil, GetAccountInput{}); err != nil || !out.OK {
		t.Fatalf("getAccount: %v %+v", err, out)
	}
	if rec.path != "/v1/account" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"current":1,"items":[{"id":"cap_1","key":"power"}],"last":1}`), rec)
	if _, out, err := s.listCapabilities(ctx(), nil, ListCapabilitiesInput{Type: "onoff", Tag: "lamp", Key: "power"}); err != nil || !out.OK {
		t.Fatalf("listCapabilities: %v %+v", err, out)
	}
	if rec.path != "/v1/teams/team_1/capabilities" {
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
	if rec.path != "/v1/account/memberships" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"current":1,"items":[{"id":"ses_1","thisDevice":true}],"last":1}`), rec)
	if _, out, err := s.listSessions(ctx(), nil, ListSessionsInput{ThisDeviceOnly: true}); err != nil || !out.OK {
		t.Fatalf("listSessions: %v %+v", err, out)
	}
	if rec.path != "/v1/account/sessions" || !strings.Contains(rec.query, "this_device=true") {
		t.Fatalf("path %q query %q", rec.path, rec.query)
	}

	s = testServer(t, 200, envelope(`{"id":"ses_1","userAgent":"agent"}`), rec)
	if _, out, err := s.getSession(ctx(), nil, GetSessionInput{SessionID: "ses_1"}); err != nil || !out.OK {
		t.Fatalf("getSession: %v %+v", err, out)
	}
	if rec.path != "/v1/account/sessions/ses_1" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"productUpdates":true}`), rec)
	if _, out, err := s.getCommunicationPreferences(ctx(), nil, GetCommunicationPreferencesInput{}); err != nil || !out.OK {
		t.Fatalf("getCommunicationPreferences: %v %+v", err, out)
	}
	if rec.path != "/v1/account/communication-preferences" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"api":"api.example","zones":[{"id":"sg-sin-a","name":"Singapore","open":true}]}`), rec)
	if _, out, err := s.listZones(ctx(), nil, ListZonesInput{}); err != nil || !out.OK {
		t.Fatalf("listZones: %v %+v", err, out)
	}
	if rec.path != "/v1/zones" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"current":1,"items":[{"id":"spc_1","zone":"us"}],"last":1}`), rec)
	if _, out, err := s.listSpaces(ctx(), nil, ListSpacesInput{Zone: "us"}); err != nil || !out.OK {
		t.Fatalf("listSpaces: %v %+v", err, out)
	}
	if rec.path != "/v1/teams/team_1/spaces" || rec.query != "zone=us" {
		t.Fatalf("path %q", rec.path)
	}
}

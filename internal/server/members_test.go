package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestMembersTools(t *testing.T) {
	rec := &record{}
	s := testServer(t, 200, envelope(`{"current":1,"items":[{"id":"mem_1","role":"manager"}],"last":1}`), rec)
	if _, out, err := s.listTeamMembers(ctx(), nil, ListTeamMembersInput{TeamID: "team_1"}); err != nil || !out.OK {
		t.Fatalf("listTeamMembers: %v %+v", err, out)
	}
	if rec.path != "/teams/team_1/members" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 201, envelope(`{"id":"mem_2","status":"active"}`), rec)
	if _, out, err := s.addTeamMember(ctx(), nil, AddTeamMemberInput{TeamID: "team_1", InviteID: "inv_9"}); err != nil || !out.OK {
		t.Fatalf("addTeamMember: %v %+v", err, out)
	}
	if rec.method != http.MethodPost || !strings.Contains(rec.body, `"inviteId":"inv_9"`) {
		t.Fatalf("request %s body=%q", rec.method, rec.body)
	}

	s = testServer(t, 200, envelope(`{"id":"mem_1"}`), rec)
	if _, out, _ := s.getTeamMember(ctx(), nil, GetTeamMemberInput{TeamID: "team_1", MemberID: "mem_1"}); !out.OK {
		t.Fatal("getTeamMember not ok")
	}
	if rec.path != "/teams/team_1/members/mem_1" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, `{"message":"ok"}`, rec)
	if _, out, _ := s.updateTeamMember(ctx(), nil, UpdateTeamMemberInput{TeamID: "team_1", MemberID: "mem_1", Role: "guest"}); !out.OK {
		t.Fatal("updateTeamMember not ok")
	}
	if rec.method != http.MethodPatch || !strings.Contains(rec.body, `"role":"guest"`) {
		t.Fatalf("request %s body=%q", rec.method, rec.body)
	}

	s = testServer(t, 200, `{"message":"ok"}`, rec)
	if _, out, _ := s.removeTeamMember(ctx(), nil, RemoveTeamMemberInput{TeamID: "team_1", MemberID: "mem_1"}); !out.OK {
		t.Fatal("removeTeamMember not ok")
	}
	if rec.method != http.MethodDelete || rec.path != "/teams/team_1/members/mem_1" {
		t.Fatalf("request %s %s", rec.method, rec.path)
	}

	s = testServer(t, 200, envelope(`{"current":1,"items":[{"id":"mem_3","status":"invited"}],"last":1}`), rec)
	if _, out, err := s.listTeamInvites(ctx(), nil, ListTeamInvitesInput{TeamID: "team_1"}); err != nil || !out.OK {
		t.Fatalf("listTeamInvites: %v %+v", err, out)
	}
	if rec.path != "/teams/team_1/invites" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 201, envelope(`{"id":"mem_4","status":"invited"}`), rec)
	if _, out, err := s.createTeamInvite(ctx(), nil, CreateTeamInviteInput{TeamID: "team_1", Email: "a@b.c", Role: "guest"}); err != nil || !out.OK {
		t.Fatalf("createTeamInvite: %v %+v", err, out)
	}
	if !strings.Contains(rec.body, `"email":"a@b.c"`) || !strings.Contains(rec.body, `"role":"guest"`) {
		t.Fatalf("body=%q", rec.body)
	}

	s = testServer(t, 200, `{"message":"ok"}`, rec)
	if _, out, _ := s.revokeTeamInvite(ctx(), nil, RevokeTeamInviteInput{TeamID: "team_1", InviteID: "mem_4"}); !out.OK {
		t.Fatal("revokeTeamInvite not ok")
	}
	if rec.method != http.MethodDelete || rec.path != "/teams/team_1/invites/mem_4" {
		t.Fatalf("request %s %s", rec.method, rec.path)
	}
}

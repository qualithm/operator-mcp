package server

import (
	"net/http"
	"strings"
	"testing"

	operator "github.com/qualithm/operator-go"
)

func TestTeamsTools(t *testing.T) {
	rec := &record{}
	s := testServer(t, 200, envelope(`{"current":1,"items":[{"id":"team_1"}],"last":1}`), rec)
	if _, out, err := s.listTeams(ctx(), nil, ListTeamsInput{Page: 1, Limit: 5}); err != nil || !out.OK {
		t.Fatalf("listTeams: %v %+v", err, out)
	}
	if rec.method != http.MethodGet || rec.path != "/teams" {
		t.Fatalf("request %s %s", rec.method, rec.path)
	}

	s = testServer(t, 201, envelope(`{"id":"team_2","name":"Team abcd1234"}`), rec)
	if _, out, err := s.createTeam(ctx(), nil, CreateTeamInput{}); err != nil || !out.OK {
		t.Fatalf("createTeam: %v %+v", err, out)
	}
	if rec.method != http.MethodPost || rec.path != "/teams" || rec.body != "" {
		t.Fatalf("request %s %s body=%q", rec.method, rec.path, rec.body)
	}

	s = testServer(t, 200, envelope(`{"id":"team_2"}`), rec)
	if _, out, _ := s.getTeam(ctx(), nil, GetTeamInput{TeamID: "team_2"}); !out.OK {
		t.Fatal("getTeam not ok")
	}
	if rec.path != "/teams/team_2" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, `{"message":"ok"}`, rec)
	if _, out, _ := s.updateTeam(ctx(), nil, UpdateTeamInput{TeamID: "team_2", Name: "New"}); !out.OK {
		t.Fatal("updateTeam not ok")
	}
	if rec.method != http.MethodPatch || !strings.Contains(rec.body, `"name":"New"`) {
		t.Fatalf("request %s body=%q", rec.method, rec.body)
	}

	s = testServer(t, 200, `{"message":"ok"}`, rec)
	if _, out, _ := s.deleteTeam(ctx(), nil, DeleteTeamInput{TeamID: "team_2"}); !out.OK {
		t.Fatal("deleteTeam not ok")
	}
	if rec.method != http.MethodDelete || rec.path != "/teams/team_2" {
		t.Fatalf("request %s %s", rec.method, rec.path)
	}

	s = testServer(t, 200, envelope(`{"items":[{"deviceId":"dev_1","online":true,"lastSeenAt":1,"metrics":{"temp":21}}]}`), rec)
	_, out, err := s.getTeamDeviceState(ctx(), nil, GetTeamDeviceStateInput{TeamID: "team_2"})
	if err != nil || !out.OK {
		t.Fatalf("getTeamDeviceState: %v %+v", err, out)
	}
	if rec.path != "/teams/team_2/device-state" {
		t.Fatalf("path %q", rec.path)
	}
	snaps, ok := out.Data.([]operator.DeviceStateSnapshot)
	if !ok || len(snaps) != 1 || !snaps[0].Online {
		t.Fatalf("data = %T %+v", out.Data, out.Data)
	}
}

package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestDashboardsTools(t *testing.T) {
	rec := &record{}
	s := testServer(t, 200, envelope(`{"current":1,"items":[{"id":"dash_1","spaceId":null}],"last":1}`), rec)
	if _, out, err := s.listDashboards(ctx(), nil, ListDashboardsInput{}); err != nil || !out.OK {
		t.Fatalf("listDashboards: %v %+v", err, out)
	}
	if rec.path != "/dashboards" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 201, envelope(`{"id":"dash_2"}`), rec)
	if _, out, err := s.createDashboard(ctx(), nil, CreateDashboardInput{Name: "Board"}); err != nil || !out.OK {
		t.Fatalf("createDashboard: %v %+v", err, out)
	}
	if rec.method != http.MethodPost || !strings.Contains(rec.body, `"name":"Board"`) {
		t.Fatalf("request %s body=%q", rec.method, rec.body)
	}

	s = testServer(t, 200, envelope(`{"id":"dash_2"}`), rec)
	if _, out, _ := s.getDashboard(ctx(), nil, GetDashboardInput{DashboardID: "dash_2"}); !out.OK {
		t.Fatal("getDashboard not ok")
	}
	if rec.path != "/dashboards/dash_2" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, `{"message":"ok"}`, rec)
	if _, out, _ := s.updateDashboard(ctx(), nil, UpdateDashboardInput{DashboardID: "dash_2", Name: "B2"}); !out.OK {
		t.Fatal("updateDashboard not ok")
	}
	if rec.method != http.MethodPatch || !strings.Contains(rec.body, `"name":"B2"`) {
		t.Fatalf("request %s body=%q", rec.method, rec.body)
	}

	s = testServer(t, 200, `{"message":"ok"}`, rec)
	if _, out, _ := s.deleteDashboard(ctx(), nil, DeleteDashboardInput{DashboardID: "dash_2"}); !out.OK {
		t.Fatal("deleteDashboard not ok")
	}
	if rec.method != http.MethodDelete || rec.path != "/dashboards/dash_2" {
		t.Fatalf("request %s %s", rec.method, rec.path)
	}
}

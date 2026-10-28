package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestCommandsTools(t *testing.T) {
	rec := &record{}
	s := testServer(t, 200, envelope(`{"current":1,"items":[{"id":"cmd_1","status":"sent"}],"last":1}`), rec)
	if _, out, err := s.listDeviceCommands(ctx(), nil, ListDeviceCommandsInput{DeviceID: "dev_1"}); err != nil || !out.OK {
		t.Fatalf("listDeviceCommands: %v %+v", err, out)
	}
	if rec.path != "/v1/teams/team_1/devices/dev_1/commands" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 202, envelope(`{"id":"cmd_2","duplicate":false}`), rec)
	if _, out, err := s.sendDeviceCommand(ctx(), nil, SendDeviceCommandInput{DeviceID: "dev_1", Capability: "power", Value: true, DedupKey: "k1"}); err != nil || !out.OK {
		t.Fatalf("sendDeviceCommand: %v %+v", err, out)
	}
	if rec.method != http.MethodPost || !strings.Contains(rec.body, `"capability":"power"`) || !strings.Contains(rec.body, `"dedupKey":"k1"`) {
		t.Fatalf("request %s body=%q", rec.method, rec.body)
	}

	s = testServer(t, 200, envelope(`{"items":[{"id":"cap_1","key":"power","type":"onoff"}],"nextCursor":null}`), rec)
	if _, out, err := s.getDeviceCapabilities(ctx(), nil, GetDeviceCapabilitiesInput{DeviceID: "dev_1"}); err != nil || !out.OK {
		t.Fatalf("getDeviceCapabilities: %v %+v", err, out)
	}
	if rec.path != "/v1/teams/team_1/devices/dev_1/capabilities" {
		t.Fatalf("path %q", rec.path)
	}

	for _, parked := range []bool{true, false} {
		s = testServer(t, 204, ``, rec)
		if _, out, _ := s.updateDevice(ctx(), nil, UpdateDeviceInput{DeviceID: "dev_1", Parked: &parked}); !out.OK {
			t.Fatalf("parked=%v not ok", parked)
		}
		want := `"parked":false`
		if parked {
			want = `"parked":true`
		}
		if rec.method != http.MethodPatch || rec.path != "/v1/teams/team_1/devices/dev_1" || !strings.Contains(rec.body, want) {
			t.Fatalf("parked=%v: request %s %s body=%q", parked, rec.method, rec.path, rec.body)
		}
	}
}

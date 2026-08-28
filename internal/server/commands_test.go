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
	if rec.path != "/devices/dev_1/commands" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 202, envelope(`{"id":"cmd_2","duplicate":false}`), rec)
	if _, out, err := s.sendDeviceCommand(ctx(), nil, SendDeviceCommandInput{DeviceID: "dev_1", Capability: "power", Value: true, DedupKey: "k1"}); err != nil || !out.OK {
		t.Fatalf("sendDeviceCommand: %v %+v", err, out)
	}
	if rec.method != http.MethodPost || !strings.Contains(rec.body, `"capability":"power"`) || !strings.Contains(rec.body, `"dedupKey":"k1"`) {
		t.Fatalf("request %s body=%q", rec.method, rec.body)
	}

	s = testServer(t, 200, envelope(`[{"id":"cap_1","key":"power","type":"onoff"}]`), rec)
	if _, out, err := s.getDeviceCapabilities(ctx(), nil, GetDeviceCapabilitiesInput{DeviceID: "dev_1"}); err != nil || !out.OK {
		t.Fatalf("getDeviceCapabilities: %v %+v", err, out)
	}
	if rec.path != "/devices/dev_1/capabilities" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, `{"message":"ok"}`, rec)
	if _, out, _ := s.parkDevice(ctx(), nil, ParkDeviceInput{DeviceID: "dev_1"}); !out.OK {
		t.Fatal("parkDevice not ok")
	}
	if rec.method != http.MethodPost || rec.path != "/devices/dev_1/park" {
		t.Fatalf("request %s %s", rec.method, rec.path)
	}

	s = testServer(t, 200, `{"message":"ok"}`, rec)
	if _, out, _ := s.unparkDevice(ctx(), nil, UnparkDeviceInput{DeviceID: "dev_1"}); !out.OK {
		t.Fatal("unparkDevice not ok")
	}
	if rec.method != http.MethodPost || rec.path != "/devices/dev_1/unpark" {
		t.Fatalf("request %s %s", rec.method, rec.path)
	}
}

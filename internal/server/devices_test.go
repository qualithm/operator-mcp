package server

import (
	"encoding/json"
	"strings"
	"testing"
)

const virtualDevice = `{"id":"dev_1","virtual":true,"parkedAt":"2026-10-28T00:00:00Z",` +
	`"capabilitiesDeclared":true,"level":"development",` +
	`"checks":[{"id":"credential-certificate","pass":false,"evidence":"a token credential"}]}`

// TestDeviceToolsReturnEveryField proves an agent sees what the device API
// serves, not only the fields the tools happen to name.
func TestDeviceToolsReturnEveryField(t *testing.T) {
	t.Parallel()
	want := []string{
		`"virtual":true`,
		`"parkedAt":"2026-10-28T00:00:00Z"`,
		`"capabilitiesDeclared":true`,
		`"level":"development"`,
		`"evidence":"a token credential"`,
	}
	tests := []struct {
		name string
		body string
		call func(s *Server) (Result, error)
	}{
		{
			name: "get_device",
			body: envelope(virtualDevice),
			call: func(s *Server) (Result, error) {
				_, res, err := s.getDevice(ctx(), nil, GetDeviceInput{DeviceID: "dev_1"})
				return res, err
			},
		},
		{
			name: "list_devices",
			body: envelope(`{"items":[` + virtualDevice + `]}`),
			call: func(s *Server) (Result, error) {
				_, res, err := s.listDevices(ctx(), nil, ListDevicesInput{})
				return res, err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, err := tt.call(testServer(t, 200, tt.body, nil))
			if err != nil || !res.OK {
				t.Fatalf("err=%v ok=%v message=%q", err, res.OK, res.Message)
			}
			out, err := json.Marshal(res.Data)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			for _, field := range want {
				if !strings.Contains(string(out), field) {
					t.Errorf("result lacks %s: %s", field, out)
				}
			}
		})
	}
}

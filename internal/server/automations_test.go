package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestAutomationsTools(t *testing.T) {
	rec := &record{}
	s := testServer(t, 200, envelope(`{"current":1,"items":[{"id":"auto_1","enabled":true}],"last":1}`), rec)
	if _, out, err := s.listAutomations(ctx(), nil, ListAutomationsInput{}); err != nil || !out.OK {
		t.Fatalf("listAutomations: %v %+v", err, out)
	}
	if rec.path != "/automations" {
		t.Fatalf("path %q", rec.path)
	}

	payload := `{"trigger":{"type":"event","config":{"metric":"temp"}},"chain":[{"kind":"action","type":"notification","config":{}}]}`
	s = testServer(t, 201, envelope(`{"id":"auto_2","name":"A"}`), rec)
	if _, out, err := s.createAutomation(ctx(), nil, CreateAutomationInput{Name: "A", Payload: []byte(payload)}); err != nil || !out.OK {
		t.Fatalf("createAutomation: %v %+v", err, out)
	}
	if rec.method != http.MethodPost || !strings.Contains(rec.body, `"trigger"`) {
		t.Fatalf("request %s body=%q", rec.method, rec.body)
	}

	s = testServer(t, 400, `{"message":"Invalid request body"}`, rec)
	if _, out, err := s.createAutomation(ctx(), nil, CreateAutomationInput{Name: "A", Payload: []byte(`{bad`)}); err != nil || out.OK {
		t.Fatalf("invalid payload: want failure, got %v %+v", err, out)
	}

	s = testServer(t, 200, envelope(`{"id":"auto_2"}`), rec)
	if _, out, _ := s.getAutomation(ctx(), nil, GetAutomationInput{AutomationID: "auto_2"}); !out.OK {
		t.Fatal("getAutomation not ok")
	}
	if rec.path != "/automations/auto_2" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, `{"message":"ok"}`, rec)
	if _, out, _ := s.updateAutomation(ctx(), nil, UpdateAutomationInput{AutomationID: "auto_2", Name: "B"}); !out.OK {
		t.Fatal("updateAutomation not ok")
	}
	if rec.method != http.MethodPatch || !strings.Contains(rec.body, `"name":"B"`) {
		t.Fatalf("request %s body=%q", rec.method, rec.body)
	}

	s = testServer(t, 200, `{"message":"ok"}`, rec)
	if _, out, _ := s.updateAutomation(ctx(), nil, UpdateAutomationInput{AutomationID: "auto_2", Template: "tmpl_1", Params: []byte(`{"device":"dev_1"}`)}); !out.OK {
		t.Fatal("updateAutomation template not ok")
	}
	if !strings.Contains(rec.body, `"template":"tmpl_1"`) || !strings.Contains(rec.body, `"device":"dev_1"`) {
		t.Fatalf("template body=%q", rec.body)
	}

	s = testServer(t, 200, `{"message":"ok"}`, rec)
	if _, out, _ := s.deleteAutomation(ctx(), nil, DeleteAutomationInput{AutomationID: "auto_2"}); !out.OK {
		t.Fatal("deleteAutomation not ok")
	}
	if rec.method != http.MethodDelete || rec.path != "/automations/auto_2" {
		t.Fatalf("request %s %s", rec.method, rec.path)
	}

	for _, tc := range []struct {
		name string
		call func(*Server) (Result, error)
		path string
	}{
		{"enable", func(s *Server) (Result, error) {
			_, out, err := s.enableAutomation(ctx(), nil, EnableAutomationInput{AutomationID: "auto_2"})
			return out, err
		}, "/automations/auto_2/enable"},
		{"disable", func(s *Server) (Result, error) {
			_, out, err := s.disableAutomation(ctx(), nil, DisableAutomationInput{AutomationID: "auto_2"})
			return out, err
		}, "/automations/auto_2/disable"},
	} {
		s = testServer(t, 200, envelope(`{"id":"auto_2"}`), rec)
		out, err := tc.call(s)
		if err != nil || !out.OK {
			t.Fatalf("%s: %v %+v", tc.name, err, out)
		}
		if rec.method != http.MethodPost || rec.path != tc.path {
			t.Fatalf("%s: request %s %s", tc.name, rec.method, rec.path)
		}
	}

	s = testServer(t, 200, envelope(`{"id":"run_1"}`), rec)
	if _, out, err := s.runAutomation(ctx(), nil, RunAutomationInput{AutomationID: "auto_2"}); err != nil || !out.OK {
		t.Fatalf("runAutomation: %v %+v", err, out)
	}
	if rec.path != "/automations/auto_2/run" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"current":1,"items":[{"id":"run_1","status":"succeeded"}],"last":1}`), rec)
	if _, out, err := s.listAutomationRuns(ctx(), nil, ListAutomationRunsInput{AutomationID: "auto_2"}); err != nil || !out.OK {
		t.Fatalf("listAutomationRuns: %v %+v", err, out)
	}
	if rec.path != "/automations/auto_2/runs" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 201, envelope(`{"id":"auto_3"}`), rec)
	if _, out, err := s.createAutomationFromTemplate(ctx(), nil, CreateAutomationFromTemplateInput{Name: "T", Template: "tmpl_1", Params: []byte(`{"device":"dev_1"}`)}); err != nil || !out.OK {
		t.Fatalf("createAutomationFromTemplate: %v %+v", err, out)
	}
	if rec.path != "/automations/from-template" || !strings.Contains(rec.body, `"template":"tmpl_1"`) {
		t.Fatalf("path %q body=%q", rec.path, rec.body)
	}

	s = testServer(t, 200, envelope(`[{"id":"tmpl_1","slots":[{"name":"device","kind":"device"}]}]`), rec)
	if _, out, err := s.listAutomationTemplates(ctx(), nil, ListAutomationTemplatesInput{DeviceID: "dev_1"}); err != nil || !out.OK {
		t.Fatalf("listAutomationTemplates: %v %+v", err, out)
	}
	if rec.path != "/devices/dev_1/automation-templates" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"runId":"run_9"}`), rec)
	if _, out, err := s.triggerAutomation(ctx(), nil, TriggerAutomationInput{Secret: "trig_secret", Context: []byte(`{"k":"v"}`)}); err != nil || !out.OK {
		t.Fatalf("triggerAutomation: %v %+v", err, out)
	}
	if rec.path != "/automations/trigger" || !strings.Contains(rec.body, `"k":"v"`) {
		t.Fatalf("path %q body=%q", rec.path, rec.body)
	}

	s = testServer(t, 200, envelope(`{"url":"https://x","secret":"s3cr3t"}`), rec)
	if _, out, err := s.createAutomationTriggerSecret(ctx(), nil, CreateAutomationTriggerSecretInput{AutomationID: "auto_2"}); err != nil || !out.OK {
		t.Fatalf("createAutomationTriggerSecret: %v %+v", err, out)
	}
	if rec.path != "/automations/auto_2/trigger-secret" {
		t.Fatalf("path %q", rec.path)
	}
}

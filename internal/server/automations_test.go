package server

import (
	"encoding/json"
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
	if rec.path != "/v1/teams/team_1/automations" {
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
	if rec.path != "/v1/teams/team_1/automations/auto_2" {
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
	if rec.method != http.MethodDelete || rec.path != "/v1/teams/team_1/automations/auto_2" {
		t.Fatalf("request %s %s", rec.method, rec.path)
	}

	for _, enabled := range []bool{true, false} {
		s = testServer(t, 200, envelope(`{"id":"auto_2"}`), rec)
		if _, out, err := s.updateAutomation(ctx(), nil, UpdateAutomationInput{AutomationID: "auto_2", Enabled: &enabled}); err != nil || !out.OK {
			t.Fatalf("enabled=%v: %v %+v", enabled, err, out)
		}
		want := `"enabled":false`
		if enabled {
			want = `"enabled":true`
		}
		if rec.method != http.MethodPatch || rec.path != "/v1/teams/team_1/automations/auto_2" || !strings.Contains(rec.body, want) {
			t.Fatalf("enabled=%v: request %s %s body=%q", enabled, rec.method, rec.path, rec.body)
		}
	}

	s = testServer(t, 200, envelope(`{"id":"run_1"}`), rec)
	if _, out, err := s.runAutomation(ctx(), nil, RunAutomationInput{AutomationID: "auto_2"}); err != nil || !out.OK {
		t.Fatalf("runAutomation: %v %+v", err, out)
	}
	if rec.method != http.MethodPost || rec.path != "/v1/teams/team_1/automations/auto_2/runs" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`{"current":1,"items":[{"id":"run_1","status":"succeeded"}],"last":1}`), rec)
	if _, out, err := s.listAutomationRuns(ctx(), nil, ListAutomationRunsInput{AutomationID: "auto_2"}); err != nil || !out.OK {
		t.Fatalf("listAutomationRuns: %v %+v", err, out)
	}
	if rec.path != "/v1/teams/team_1/automations/auto_2/runs" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 201, envelope(`{"id":"auto_3"}`), rec)
	if _, out, err := s.createAutomation(ctx(), nil, CreateAutomationInput{Name: "T", Template: "tmpl_1", Params: []byte(`{"device":"dev_1"}`)}); err != nil || !out.OK {
		t.Fatalf("createAutomation from a template: %v %+v", err, out)
	}
	if rec.path != "/v1/teams/team_1/automations" || !strings.Contains(rec.body, `"template":"tmpl_1"`) || strings.Contains(rec.body, `"payload"`) {
		t.Fatalf("path %q body=%q", rec.path, rec.body)
	}

	s = testServer(t, 200, envelope(`[{"id":"tmpl_1","slots":[{"name":"device","kind":"device"}]}]`), rec)
	if _, out, err := s.listAutomationTemplates(ctx(), nil, ListAutomationTemplatesInput{DeviceID: "dev_1"}); err != nil || !out.OK {
		t.Fatalf("listAutomationTemplates: %v %+v", err, out)
	}
	if rec.path != "/v1/teams/team_1/devices/dev_1/automation-templates" {
		t.Fatalf("path %q", rec.path)
	}

	s = testServer(t, 200, envelope(`[{"id":"tmpl_1","needs":["sensor"],"usable":true,"devices":["dev_1"]}]`), rec)
	_, team, err := s.listTeamAutomationTemplates(ctx(), nil, ListTeamAutomationTemplatesInput{})
	if err != nil || !team.OK {
		t.Fatalf("listTeamAutomationTemplates: %v %+v", err, team)
	}
	got, _ := json.Marshal(team.Data)
	if rec.path != "/v1/teams/team_1/automation-templates" || !strings.Contains(string(got), `"needs":["sensor"]`) || !strings.Contains(string(got), `"usable":true`) || !strings.Contains(string(got), `"devices":["dev_1"]`) {
		t.Fatalf("path %q data %s", rec.path, got)
	}

	s = testServer(t, 200, envelope(`{"runId":"run_9"}`), rec)
	if _, out, err := s.triggerAutomation(ctx(), nil, TriggerAutomationInput{Secret: "trig_secret", Context: []byte(`{"k":"v"}`)}); err != nil || !out.OK {
		t.Fatalf("triggerAutomation: %v %+v", err, out)
	}
	if rec.path != "/v1/automations/trigger" || !strings.Contains(rec.body, `"k":"v"`) {
		t.Fatalf("path %q body=%q", rec.path, rec.body)
	}

	s = testServer(t, 200, envelope(`{"url":"https://x","secret":"s3cr3t"}`), rec)
	if _, out, err := s.createAutomationTriggerSecret(ctx(), nil, CreateAutomationTriggerSecretInput{AutomationID: "auto_2"}); err != nil || !out.OK {
		t.Fatalf("createAutomationTriggerSecret: %v %+v", err, out)
	}
	if rec.path != "/v1/teams/team_1/automations/auto_2/trigger-secret" {
		t.Fatalf("path %q", rec.path)
	}
}

func TestTestAutomation(t *testing.T) {
	result := envelope(`{"matched":false,"reason":"condition_halted","device":{"id":"dev_1","scopeId":"spc_1","online":true,"lastSeenAt":1763193600000},"reading":{"source":"device","metrics":{"temp":20},"seenAt":1763193600000},"trigger":{"type":"event","matched":true},"compared":[],"steps":[],"actions":[],"decisions":[]}`)
	rec := &record{}

	s := testServer(t, 200, result, rec)
	_, out, err := s.testAutomation(ctx(), nil, TestAutomationInput{AutomationID: "auto_1", DeviceID: "dev_1", Notify: true})
	if err != nil || !out.OK {
		t.Fatalf("saved: %v %+v", err, out)
	}
	if rec.method != http.MethodPost || rec.path != "/v1/teams/team_1/automations/auto_1/test" {
		t.Fatalf("saved request %s %s", rec.method, rec.path)
	}
	if !strings.Contains(rec.body, `"deviceId":"dev_1"`) || !strings.Contains(rec.body, `"notify":true`) {
		t.Fatalf("saved body=%q", rec.body)
	}

	s = testServer(t, 200, result, rec)
	_, out, err = s.testAutomation(ctx(), nil, TestAutomationInput{
		Template: "tmpl_1",
		Params:   []byte(`{"device":"dev_1"}`),
		ScopeID:  "spc_1",
		Sample:   []byte(`{"metrics":{"temp":40},"state":"online"}`),
	})
	if err != nil || !out.OK {
		t.Fatalf("draft: %v %+v", err, out)
	}
	if rec.path != "/v1/teams/team_1/automations/test" {
		t.Fatalf("draft path %q", rec.path)
	}
	for _, want := range []string{`"template":"tmpl_1"`, `"device":"dev_1"`, `"scopeId":"spc_1"`, `"temp":40`, `"state":"online"`} {
		if !strings.Contains(rec.body, want) {
			t.Fatalf("draft body %q lacks %s", rec.body, want)
		}
	}
	if strings.Contains(rec.body, `"name"`) {
		t.Fatalf("draft sent an empty name: %q", rec.body)
	}

	payload := `{"trigger":{"type":"event","config":{"metric":"temp"}},"chain":[]}`
	s = testServer(t, 200, result, rec)
	if _, out, err := s.testAutomation(ctx(), nil, TestAutomationInput{Name: "Draft", Payload: []byte(payload), DeviceID: "dev_1"}); err != nil || !out.OK {
		t.Fatalf("draft payload: %v %+v", err, out)
	}
	if !strings.Contains(rec.body, `"trigger"`) || !strings.Contains(rec.body, `"name":"Draft"`) {
		t.Fatalf("draft payload body=%q", rec.body)
	}

	for name, in := range map[string]TestAutomationInput{
		"sample":  {AutomationID: "auto_1", Sample: []byte(`{bad`)},
		"payload": {Payload: []byte(`{bad`)},
		"params":  {Template: "tmpl_1", Params: []byte(`{bad`)},
	} {
		s = testServer(t, 200, result, rec)
		if _, out, err := s.testAutomation(ctx(), nil, in); err != nil || out.OK {
			t.Fatalf("%s: want failure, got %v %+v", name, err, out)
		}
	}

	for _, in := range []TestAutomationInput{{AutomationID: "auto_1"}, {Template: "tmpl_1"}} {
		s = testServer(t, 400, `{"message":"test needs a reading"}`, rec)
		if _, out, err := s.testAutomation(ctx(), nil, in); err != nil || out.OK {
			t.Fatalf("platform refusal: want failure, got %v %+v", err, out)
		}
	}
}

package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/qualithm/operator-go"
)

// pagedServer serves two pages of devices: the first with nextCursor "c2",
// the second, asked for with cursor=c2, with a null nextCursor.
func pagedServer(t *testing.T, queries *[]string) *Server {
	t.Helper()
	return newServerWith(func(dryRun bool) (*operator.Client, error) {
		rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
			*queries = append(*queries, req.URL.RawQuery)
			body := envelope(`{"items":[{"id":"dev_1"}],"nextCursor":"c2"}`)
			if req.URL.Query().Get("cursor") == "c2" {
				body = envelope(`{"items":[{"id":"dev_2"}],"nextCursor":null}`)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		})
		return operator.New(tok,
			operator.WithDryRun(dryRun),
			operator.WithHTTPClient(&http.Client{Transport: rt}),
			operator.WithTeam("team_1"),
		)
	})
}

// TestListDevicesWalksTwoPages proves an agent can page a list tool to its
// end: it passes each page's nextCursor back as cursor until it is null.
func TestListDevicesWalksTwoPages(t *testing.T) {
	var queries []string
	s := pagedServer(t, &queries)

	var ids []string
	cursor := ""
	for range 3 {
		_, out, err := s.listDevices(ctx(), nil, ListDevicesInput{Cursor: cursor})
		if err != nil || !out.OK {
			t.Fatalf("listDevices: err=%v out=%+v", err, out)
		}
		raw, err := json.Marshal(out.Data)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var page struct {
			Items      []struct{ ID string } `json:"items"`
			NextCursor *string               `json:"nextCursor"`
		}
		if err := json.Unmarshal(raw, &page); err != nil {
			t.Fatalf("unmarshal %s: %v", raw, err)
		}
		for _, item := range page.Items {
			ids = append(ids, item.ID)
		}
		if page.NextCursor == nil {
			break
		}
		cursor = *page.NextCursor
	}

	if strings.Join(ids, ",") != "dev_1,dev_2" {
		t.Errorf("walked %v, want dev_1 then dev_2", ids)
	}
	if len(queries) != 2 || strings.Contains(queries[0], "cursor") || !strings.Contains(queries[1], "cursor=c2") {
		t.Errorf("queries = %q, want no cursor then cursor=c2", queries)
	}
}

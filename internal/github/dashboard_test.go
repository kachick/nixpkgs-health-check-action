package github

import (
	"context"
	"encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kachick/nixpkgs-health-check-action/internal/report"
)

func TestUpsertDashboardCreateAndAssign(t *testing.T) {
	var createdBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/repos/owner/repo/issues"):
			// Return an untrusted issue with the same marker to verify it is ignored.
			w.Write([]byte(`[
				{
					"number": 99,
					"title": "Spoofed",
					"body": "<!-- nixpkgs-health-check-dashboard: maintainer -->",
					"author_association": "NONE",
					"user": {"login": "random-user"}
				}
			]`))
		case r.Method == http.MethodPost && r.URL.Path == "/repos/owner/repo/issues":
			b, _ := io.ReadAll(r.Body)
			createdBody = string(b)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"number": 100, "html_url": "https://github.com/owner/repo/issues/100"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	client := &Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
		Token:      "test-token",
		Repo:       "owner/repo",
	}

	res, err := client.UpsertDashboard(context.Background(), UpsertOptions{
		DashboardID: "maintainer",
		Body:        report.Marker("maintainer") + "\n\nerror",
		Assignee:    "kachick",
		HasErrors:   true,
	})
	if err != nil {
		t.Fatalf("UpsertDashboard: %v", err)
	}
	if !res.Created || res.Number != 100 {
		t.Errorf("unexpected result: %+v", res)
	}
	if !strings.Contains(createdBody, `"assignees":["kachick"]`) {
		t.Errorf("expected assignee on create, got: %s", createdBody)
	}
}

func TestUpsertDashboardUnassignWhenHealthy(t *testing.T) {
	var patchBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/repos/owner/repo/issues"):
			w.Write([]byte(`[
				{
					"number": 7,
					"title": "Nixpkgs Health Check Dashboard (maintainer)",
					"body": "<!-- nixpkgs-health-check-dashboard: maintainer -->\nold",
					"author_association": "NONE",
					"user": {"login": "github-actions[bot]"},
					"assignees": [{"login": "kachick"}]
				}
			]`))
		case r.Method == http.MethodPatch && r.URL.Path == "/repos/owner/repo/issues/7":
			b, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(b, &patchBody); err != nil {
				t.Fatalf("unmarshal patch body: %v", err)
			}
			w.Write([]byte(`{"number": 7, "html_url": "https://github.com/owner/repo/issues/7"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	client := &Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
		Token:      "test-token",
		Repo:       "owner/repo",
	}

	res, err := client.UpsertDashboard(context.Background(), UpsertOptions{
		DashboardID: "maintainer",
		Body:        report.Marker("maintainer") + "\nhealthy",
		Assignee:    "kachick",
		HasErrors:   false,
	})
	if err != nil {
		t.Fatalf("UpsertDashboard: %v", err)
	}
	if res.Created || res.Number != 7 {
		t.Errorf("unexpected result: %+v", res)
	}
	assignees, ok := patchBody["assignees"].([]any)
	if !ok || len(assignees) != 0 {
		t.Errorf("expected empty assignees array to unassign, got: %#v", patchBody)
	}
}

func TestUpsertDashboardRejectsInvalidID(t *testing.T) {
	client := &Client{Token: "tok", Repo: "owner/repo"}
	_, err := client.UpsertDashboard(context.Background(), UpsertOptions{
		DashboardID: "bad --> id",
	})
	if err == nil {
		t.Fatal("expected error for invalid dashboard ID")
	}
}

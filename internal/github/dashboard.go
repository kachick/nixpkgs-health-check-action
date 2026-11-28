package github

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/kachick/nixpkgs-health-check-action/internal/report"
)

const defaultBaseURL = "https://api.github.com"

var (
	validDashboardID = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	validRepo        = regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)
)

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	Token      string
	Repo       string
}

type UpsertOptions struct {
	DashboardID string
	Title       string
	Body        string
	Assignee    string
	HasErrors   bool
}

type UpsertResult struct {
	Number  int
	HTMLURL string
	Created bool
}

type issue struct {
	Number            int          `json:"number"`
	Title             string       `json:"title"`
	Body              string       `json:"body"`
	HTMLURL           string       `json:"html_url"`
	AuthorAssociation string       `json:"author_association"`
	User              issueUser    `json:"user"`
	Assignees         []issueUser  `json:"assignees"`
	PullRequest       *pullRefStub `json:"pull_request"`
}

type issueUser struct {
	Login string `json:"login"`
}

type pullRefStub struct {
	URL string `json:"url"`
}

type createIssueRequest struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Assignees []string `json:"assignees,omitempty"`
}

type updateIssueRequest struct {
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Assignees *[]string `json:"assignees,omitzero"`
}

func (c *Client) UpsertDashboard(ctx context.Context, opts UpsertOptions) (*UpsertResult, error) {
	if !validDashboardID.MatchString(opts.DashboardID) {
		return nil, fmt.Errorf("invalid dashboard ID %q: must match %s", opts.DashboardID, validDashboardID.String())
	}
	if !validRepo.MatchString(c.Repo) {
		return nil, fmt.Errorf("invalid repository %q: expected owner/repo", c.Repo)
	}
	if c.Token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is required to update dashboard issue")
	}

	title := opts.Title
	if title == "" {
		title = fmt.Sprintf("Nixpkgs Health Check Dashboard (%s)", opts.DashboardID)
	}

	marker := report.Marker(opts.DashboardID)
	existing, err := c.findExistingIssue(ctx, marker)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		reqBody := createIssueRequest{
			Title: title,
			Body:  opts.Body,
		}
		if opts.HasErrors && opts.Assignee != "" {
			reqBody.Assignees = []string{opts.Assignee}
		}
		var created issue
		if err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/issues", c.Repo), reqBody, &created); err != nil {
			return nil, fmt.Errorf("create dashboard issue: %w", err)
		}
		return &UpsertResult{
			Number:  created.Number,
			HTMLURL: created.HTMLURL,
			Created: true,
		}, nil
	}

	reqBody := updateIssueRequest{
		Title: title,
		Body:  opts.Body,
	}
	if nextAssignees, changed := nextAssigneesList(existing.Assignees, opts.Assignee, opts.HasErrors); changed {
		reqBody.Assignees = &nextAssignees
	}

	var updated issue
	if err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("/repos/%s/issues/%d", c.Repo, existing.Number), reqBody, &updated); err != nil {
		return nil, fmt.Errorf("update dashboard issue #%d: %w", existing.Number, err)
	}
	return &UpsertResult{
		Number:  updated.Number,
		HTMLURL: updated.HTMLURL,
		Created: false,
	}, nil
}

func (c *Client) findExistingIssue(ctx context.Context, marker string) (*issue, error) {
	// Cap pagination to avoid unbounded API calls on large repos.
	const maxPages = 5
	for page := 1; page <= maxPages; page++ {
		path := fmt.Sprintf("/repos/%s/issues?state=open&per_page=100&page=%d", c.Repo, page)
		var issues []issue
		if err := c.doJSON(ctx, http.MethodGet, path, nil, &issues); err != nil {
			return nil, fmt.Errorf("list open issues (page %d): %w", page, err)
		}
		for i := range issues {
			iss := &issues[i]
			if iss.PullRequest != nil {
				continue
			}
			if !isTrustedAuthor(iss) {
				continue
			}
			if strings.Contains(iss.Body, marker) {
				return iss, nil
			}
		}
		if len(issues) < 100 {
			break
		}
	}
	return nil, nil
}

// isTrustedAuthor prevents untrusted users from hijacking the dashboard
// by opening an issue that embeds the same HTML comment marker.
func isTrustedAuthor(iss *issue) bool {
	if iss.User.Login == "github-actions[bot]" {
		return true
	}
	switch iss.AuthorAssociation {
	case "OWNER", "MEMBER", "COLLABORATOR":
		return true
	default:
		return false
	}
}

func nextAssigneesList(current []issueUser, target string, hasErrors bool) ([]string, bool) {
	if target == "" {
		return nil, false
	}

	logins := make([]string, 0, len(current)+1)
	alreadyAssigned := false
	for _, u := range current {
		if strings.EqualFold(u.Login, target) {
			alreadyAssigned = true
			if !hasErrors {
				continue
			}
		}
		logins = append(logins, u.Login)
	}

	if hasErrors {
		if alreadyAssigned {
			return nil, false
		}
		logins = append(logins, target)
		return logins, true
	}

	if !alreadyAssigned {
		return nil, false
	}
	return logins, true
}

func (c *Client) doJSON(ctx context.Context, method, path string, reqPayload any, respOut any) error {
	base := c.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	base = strings.TrimRight(base, "/")

	var bodyReader io.Reader
	if reqPayload != nil {
		b, err := json.Marshal(reqPayload)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, base+path, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if reqPayload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("github api %s %s returned %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(errBody)))
	}

	if respOut != nil {
		if err := json.UnmarshalRead(resp.Body, respOut); err != nil {
			return fmt.Errorf("decode response from %s %s: %w", method, path, err)
		}
	}
	return nil
}

package updates

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gopress/internal/buildinfo"
)

const (
	defaultBaseURL = "https://api.github.com"
	defaultTTL     = 5 * time.Minute
	maxCommits     = 30
)

// UpstreamError describes a non-200 response from the GitHub API.
type UpstreamError struct {
	Status  int
	Message string
}

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("github api returned %d: %s", e.Status, e.Message)
}

type Config struct {
	Repository string
	Branch     string
	Token      string
	BaseURL    string
	TTL        time.Duration
}

type Service struct {
	repository string
	branch     string
	token      string
	baseURL    string
	ttl        time.Duration
	client     *http.Client
	now        func() time.Time

	mu        sync.Mutex
	cached    *Status
	checkedAt time.Time
}

type Build struct {
	Version     string `json:"version"`
	Commit      string `json:"commit"`
	ShortCommit string `json:"short_commit"`
	BuildTime   string `json:"build_time"`
}

type Commit struct {
	SHA      string    `json:"sha"`
	ShortSHA string    `json:"short_sha"`
	Message  string    `json:"message"`
	Author   string    `json:"author"`
	Date     time.Time `json:"date"`
	URL      string    `json:"url"`
}

type Status struct {
	Repository      string    `json:"repository"`
	Branch          string    `json:"branch"`
	Current         Build     `json:"current"`
	Latest          *Commit   `json:"latest"`
	Behind          int       `json:"behind"`
	UpdateAvailable bool      `json:"update_available"`
	Comparable      bool      `json:"comparable"`
	CompareURL      string    `json:"compare_url"`
	Commits         []Commit  `json:"commits"`
	CheckedAt       time.Time `json:"checked_at"`
	Cached          bool      `json:"cached"`
}

func NewService(cfg Config) *Service {
	repository := strings.TrimSpace(cfg.Repository)
	if repository == "" {
		repository = "Bigzzh2022/GoPress"
	}
	branch := strings.TrimSpace(cfg.Branch)
	if branch == "" {
		branch = "main"
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = defaultTTL
	}
	return &Service{
		repository: repository,
		branch:     branch,
		token:      strings.TrimSpace(cfg.Token),
		baseURL:    baseURL,
		ttl:        ttl,
		client:     &http.Client{Timeout: 12 * time.Second},
		now:        time.Now,
	}
}

// Check compares the running build against the configured branch. Results are
// cached briefly so refreshing the admin page does not exhaust the anonymous
// GitHub API quota.
func (s *Service) Check(ctx context.Context, force bool) (Status, error) {
	s.mu.Lock()
	if !force && s.cached != nil && s.now().Sub(s.checkedAt) < s.ttl {
		status := *s.cached
		status.Cached = true
		s.mu.Unlock()
		return status, nil
	}
	s.mu.Unlock()

	status, err := s.fetch(ctx)
	if err != nil {
		return Status{}, err
	}

	s.mu.Lock()
	snapshot := status
	s.cached = &snapshot
	s.checkedAt = s.now()
	s.mu.Unlock()
	return status, nil
}

func (s *Service) fetch(ctx context.Context) (Status, error) {
	commits, err := s.fetchCommits(ctx)
	if err != nil {
		return Status{}, err
	}

	status := Status{
		Repository: s.repository,
		Branch:     s.branch,
		Current: Build{
			Version:     buildinfo.Version,
			Commit:      buildinfo.Commit,
			ShortCommit: buildinfo.ShortCommit(),
			BuildTime:   buildinfo.BuildTime,
		},
		Commits:   make([]Commit, 0),
		CheckedAt: s.now().UTC(),
	}
	if len(commits) > 0 {
		latest := commits[0]
		status.Latest = &latest
	}

	if !comparableCommit(buildinfo.Commit) {
		return status, nil
	}
	status.CompareURL = fmt.Sprintf("https://github.com/%s/compare/%s...%s", s.repository, url.PathEscape(buildinfo.Commit), url.PathEscape(s.branch))

	var comparison compareResponse
	path := fmt.Sprintf("/repos/%s/compare/%s...%s", s.repositoryPath(), url.PathEscape(buildinfo.Commit), url.PathEscape(s.branch))
	if err := s.getJSON(ctx, path, &comparison); err != nil {
		var upstream *UpstreamError
		if errors.As(err, &upstream) && upstream.Status == http.StatusNotFound {
			// The running commit is not in the repository (local build or a
			// commit that was force-pushed away); keep the latest commit only.
			return status, nil
		}
		return Status{}, err
	}

	status.Comparable = true
	if comparison.HTMLURL != "" {
		status.CompareURL = comparison.HTMLURL
	}
	switch comparison.Status {
	case "identical", "behind":
		status.Behind = 0
		status.UpdateAvailable = false
	case "ahead", "diverged":
		status.Behind = comparison.AheadBy
		status.UpdateAvailable = comparison.AheadBy > 0
	}

	if comparison.Status == "ahead" || comparison.Status == "diverged" {
		status.Commits = toCommits(comparison.Commits)
	}
	return status, nil
}

func (s *Service) fetchCommits(ctx context.Context) ([]Commit, error) {
	path := fmt.Sprintf("/repos/%s/commits?sha=%s&per_page=%d", s.repositoryPath(), url.QueryEscape(s.branch), maxCommits)
	var commits []githubCommit
	if err := s.getJSON(ctx, path, &commits); err != nil {
		return nil, err
	}
	return toCommits(commits), nil
}

func (s *Service) getJSON(ctx context.Context, path string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+path, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "GoPress-Update-Check")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if s.token != "" {
		request.Header.Set("Authorization", "Bearer "+s.token)
	}

	response, err := s.client.Do(request)
	if err != nil {
		return fmt.Errorf("github request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
		return &UpstreamError{Status: response.StatusCode, Message: strings.TrimSpace(string(body))}
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(target); err != nil {
		return fmt.Errorf("decode github response: %w", err)
	}
	return nil
}

func (s *Service) repositoryPath() string {
	parts := strings.Split(s.repository, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

func comparableCommit(commit string) bool {
	commit = strings.TrimSpace(commit)
	return commit != "" && commit != "unknown"
}

type githubCommit struct {
	SHA     string `json:"sha"`
	HTMLURL string `json:"html_url"`
	Commit  struct {
		Message string `json:"message"`
		Author  struct {
			Name string    `json:"name"`
			Date time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

type compareResponse struct {
	Status  string         `json:"status"`
	AheadBy int            `json:"ahead_by"`
	HTMLURL string         `json:"html_url"`
	Commits []githubCommit `json:"commits"`
}

func toCommits(source []githubCommit) []Commit {
	result := make([]Commit, 0, len(source))
	for _, item := range source {
		result = append(result, toCommit(item))
	}
	return result
}

func toCommit(item githubCommit) Commit {
	message := strings.TrimSpace(item.Commit.Message)
	if line := strings.SplitN(message, "\n", 2)[0]; line != "" {
		message = line
	}
	if runes := []rune(message); len(runes) > 200 {
		message = string(runes[:200]) + "…"
	}
	return Commit{
		SHA:      item.SHA,
		ShortSHA: shortSHA(item.SHA),
		Message:  message,
		Author:   item.Commit.Author.Name,
		Date:     item.Commit.Author.Date,
		URL:      item.HTMLURL,
	}
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

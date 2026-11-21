package updates

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gopress/internal/buildinfo"
)

const commitsPayload = `[
  {
    "sha": "ffffffffffffffffffffffffffffffffffffffff",
    "html_url": "https://github.com/acme/blog/commit/fffffff",
    "commit": {
      "message": "feat: latest change\n\nlong body",
      "author": {"name": "Ada", "date": "2026-11-20T08:00:00Z"}
    }
  }
]`

func TestCheckReportsBehindCommits(t *testing.T) {
	restoreBuildInfo(t, "abc1234", "v1.0.0")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/acme/blog/commits":
			_, _ = w.Write([]byte(commitsPayload))
		case "/repos/acme/blog/compare/abc1234...main":
			_, _ = w.Write([]byte(`{
              "status": "ahead",
              "ahead_by": 2,
              "html_url": "https://github.com/acme/blog/compare/abc1234...main",
              "commits": [
                {"sha": "bbbbb", "html_url": "u1", "commit": {"message": "one", "author": {"name": "Ada", "date": "2026-11-19T08:00:00Z"}}},
                {"sha": "ccccc", "html_url": "u2", "commit": {"message": "two", "author": {"name": "Bob", "date": "2026-11-20T08:00:00Z"}}}
              ]
            }`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	service := NewService(Config{Repository: "acme/blog", Branch: "main", BaseURL: server.URL, TTL: time.Minute})
	status, err := service.Check(context.Background(), true)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !status.Comparable || !status.UpdateAvailable {
		t.Fatalf("expected comparable update, got %+v", status)
	}
	if status.Behind != 2 || len(status.Commits) != 2 {
		t.Fatalf("expected 2 commits behind, got behind=%d commits=%d", status.Behind, len(status.Commits))
	}
	if status.Latest == nil || status.Latest.Message != "feat: latest change" {
		t.Fatalf("unexpected latest commit: %+v", status.Latest)
	}
	if status.Current.ShortCommit != "abc1234" || status.Current.Version != "v1.0.0" {
		t.Fatalf("unexpected current build: %+v", status.Current)
	}
}

func TestCheckWhenCompareCannotFindBuild(t *testing.T) {
	restoreBuildInfo(t, "deadbeef", "dev")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/acme/blog/commits" {
			_, _ = w.Write([]byte(commitsPayload))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	service := NewService(Config{Repository: "acme/blog", BaseURL: server.URL})
	status, err := service.Check(context.Background(), true)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if status.Comparable || status.UpdateAvailable {
		t.Fatalf("expected non-comparable status, got %+v", status)
	}
	if status.Latest == nil {
		t.Fatal("expected latest commit to be returned")
	}
}

func TestCheckIdenticalBuild(t *testing.T) {
	restoreBuildInfo(t, "abc1234", "v1.0.0")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/acme/blog/commits" {
			_, _ = w.Write([]byte(commitsPayload))
			return
		}
		_, _ = w.Write([]byte(`{"status":"identical","ahead_by":0,"commits":[]}`))
	}))
	defer server.Close()

	service := NewService(Config{Repository: "acme/blog", BaseURL: server.URL})
	status, err := service.Check(context.Background(), true)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !status.Comparable || status.UpdateAvailable || status.Behind != 0 {
		t.Fatalf("expected up-to-date status, got %+v", status)
	}
}

func TestCheckUsesCacheUntilForced(t *testing.T) {
	restoreBuildInfo(t, "abc1234", "v1.0.0")

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/acme/blog/commits" {
			requests++
			_, _ = w.Write([]byte(commitsPayload))
			return
		}
		_, _ = w.Write([]byte(`{"status":"identical","ahead_by":0,"commits":[]}`))
	}))
	defer server.Close()

	service := NewService(Config{Repository: "acme/blog", BaseURL: server.URL, TTL: time.Minute})
	if _, err := service.Check(context.Background(), false); err != nil {
		t.Fatalf("first check: %v", err)
	}
	cached, err := service.Check(context.Background(), false)
	if err != nil {
		t.Fatalf("cached check: %v", err)
	}
	if requests != 1 || !cached.Cached {
		t.Fatalf("expected cached result, requests=%d cached=%v", requests, cached.Cached)
	}
	if _, err := service.Check(context.Background(), true); err != nil {
		t.Fatalf("forced check: %v", err)
	}
	if requests != 2 {
		t.Fatalf("expected forced check to bypass cache, requests=%d", requests)
	}
}

func restoreBuildInfo(t *testing.T, commit, version string) {
	t.Helper()
	oldCommit, oldVersion := buildinfo.Commit, buildinfo.Version
	buildinfo.Commit, buildinfo.Version = commit, version
	t.Cleanup(func() {
		buildinfo.Commit, buildinfo.Version = oldCommit, oldVersion
	})
}

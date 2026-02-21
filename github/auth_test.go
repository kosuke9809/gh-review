package github_test

import (
	"testing"

	"github.com/kosuke9809/gh-review/github"
)

func TestResolveAuthHost(t *testing.T) {
	t.Setenv("GH_HOST", "GITHUB.EXAMPLE.COM")
	if got := github.ResolveAuthHost("api.github.com"); got != "github.com" {
		t.Fatalf("ResolveAuthHost(api.github.com) = %q, want %q", got, "github.com")
	}
	if got := github.ResolveAuthHost(""); got != "github.example.com" {
		t.Fatalf("ResolveAuthHost(\"\") = %q, want %q", got, "github.example.com")
	}
}

func TestTokenForHost_UsesGitHubTokenEnv(t *testing.T) {
	t.Setenv("GH_TOKEN", "gh-token")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_ENTERPRISE_TOKEN", "")
	t.Setenv("GITHUB_ENTERPRISE_TOKEN", "")

	token, source, err := github.TokenForHost("github.com")
	if err != nil {
		t.Fatalf("TokenForHost(github.com) error = %v", err)
	}
	if token != "gh-token" {
		t.Fatalf("token = %q, want %q", token, "gh-token")
	}
	if source != "GH_TOKEN" {
		t.Fatalf("source = %q, want %q", source, "GH_TOKEN")
	}
}

func TestTokenForHost_UsesEnterpriseTokenEnv(t *testing.T) {
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_ENTERPRISE_TOKEN", "ent-token")
	t.Setenv("GITHUB_ENTERPRISE_TOKEN", "")

	token, source, err := github.TokenForHost("github.example.com")
	if err != nil {
		t.Fatalf("TokenForHost(github.example.com) error = %v", err)
	}
	if token != "ent-token" {
		t.Fatalf("token = %q, want %q", token, "ent-token")
	}
	if source != "GH_ENTERPRISE_TOKEN" {
		t.Fatalf("source = %q, want %q", source, "GH_ENTERPRISE_TOKEN")
	}
}

package github

import (
	"fmt"
	"strings"

	ghauth "github.com/cli/go-gh/v2/pkg/auth"
)

// ResolveAuthHost normalizes the host used for auth.
// If remoteHost is empty, gh's default host resolution is used.
func ResolveAuthHost(remoteHost string) string {
	remoteHost = strings.TrimSpace(remoteHost)
	if remoteHost != "" {
		return ghauth.NormalizeHostname(remoteHost)
	}
	host, _ := ghauth.DefaultHost()
	return ghauth.NormalizeHostname(host)
}

// TokenForHost resolves a token for the remote host via gh-compatible rules.
func TokenForHost(remoteHost string) (token, source string, err error) {
	host := ResolveAuthHost(remoteHost)
	token, source = ghauth.TokenForHost(host)
	if token != "" {
		return token, source, nil
	}
	if host == "github.com" {
		return "", source, fmt.Errorf("no GitHub token found for %q; run `gh auth login` or set GH_TOKEN/GITHUB_TOKEN", host)
	}
	if ghauth.IsEnterprise(host) {
		return "", source, fmt.Errorf("no GitHub token found for %q; run `gh auth login --hostname %s` or set GH_ENTERPRISE_TOKEN/GITHUB_ENTERPRISE_TOKEN", host, host)
	}
	return "", source, fmt.Errorf("no GitHub token found for %q; run `gh auth login --hostname %s` or set GH_TOKEN/GITHUB_TOKEN", host, host)
}

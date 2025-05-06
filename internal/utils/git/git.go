package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"strings"
)

type RepoInfo struct {
	URLs      []string
	FullNames []string
}

func (r RepoInfo) MatchesName(fullName string) bool {
	for _, fn := range r.FullNames {
		if fn == fullName {
			return true
		}
	}
	return false
}

func CurrentRepo() (RepoInfo, error) {
	urls, err := getRepoURLs()
	if err != nil {
		return RepoInfo{}, err
	}
	seenNames := make(map[string]bool)
	var fullNames []string
	for _, u := range urls {
		fn, err := GetFullName(u)
		if err != nil {
			return RepoInfo{}, err
		}
		if seenNames[fn] {
			continue
		}
		seenNames[fn] = true
		fullNames = append(fullNames, fn)
	}
	return RepoInfo{URLs: urls, FullNames: fullNames}, nil
}

func GetFullName(url string) (string, error) {
	// Basic validation for URL format
	if !strings.Contains(url, "://") && !strings.Contains(url, "@") {
		return "", fmt.Errorf("invalid URL format: %s", url)
	}

	endpoint, err := transport.NewEndpoint(url)
	if err != nil {
		return "", err
	}
	name := strings.TrimSuffix(endpoint.Path, ".git")
	name = strings.TrimPrefix(name, "/")
	return name, nil
}

func GetRoot() (string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return "", err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return "", err
	}
	return wt.Filesystem.Root(), nil
}

func getRepoURLs() ([]string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, err
	}

	// Get the remote named "origin"
	remote, err := repo.Remote("origin")
	if err != nil {
		return nil, err
	}

	// Get the URL of the remote
	urls := remote.Config().URLs
	if len(urls) == 0 {
		return nil, fmt.Errorf("no remote URL found")
	}
	return urls, nil
}

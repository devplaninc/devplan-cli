package git

import (
	"errors"
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"os"
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

func IsNotInRepoErr(err error) bool {
	return errors.Is(err, git.ErrRepositoryNotExists)
}

func EnsureInRepo() RepoInfo {
	repo, err := CurrentRepo()
	if err == nil {
		return repo
	}
	if !IsNotInRepoErr(err) {
		out.Failf("Failed to get current repository: %v\n", err)
		os.Exit(1)
	}
	out.Pfail("Not in a git repository\n")
	fmt.Printf("Please, clone a repository first or navigate to the cloned one and run the command from inside the git repository.\n")
	os.Exit(1)
	return RepoInfo{}
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

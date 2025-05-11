package git

import (
	"errors"
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"io"
	"os"
	"os/exec"
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

func EnsureRepoPath(path string) RepoInfo {
	repo, err := RepoAtPath(path)
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

func EnsureInRepo() RepoInfo {
	return EnsureRepoPath(".")
}

func CurrentRepo() (RepoInfo, error) {
	return RepoAtPath(".")
}

func RepoAtPath(path string) (RepoInfo, error) {
	urls, err := getRepoURLs(path)
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

func Clone(repoURL string, targetPath string, outWriter io.Writer) error {
	cmd := exec.Command("git", "clone", repoURL, targetPath)
	cmd.Stdout = outWriter
	cmd.Stderr = outWriter
	return cmd.Run()
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

func getRepoURLs(path string) ([]string, error) {
	repo, err := git.PlainOpen(path)
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

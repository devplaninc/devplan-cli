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

func (r RepoInfo) GetFullName() string {
	return r.FullNames[0]
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

func GetRepoInfoFromURL(url string) (RepoInfo, error) {
	fn, err := GetFullName(url)
	if err != nil {
		return RepoInfo{}, err
	}
	return RepoInfo{URLs: []string{url}, FullNames: []string{fn}}, nil
}

func IsValidURL(url string) bool {
	return strings.Contains(url, "://") || strings.Contains(url, "@")
}

func GetFullName(url string) (string, error) {
	if !IsValidURL(url) {
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

type CloneOptions struct {
	RepoURL          string
	TargetPath       string
	CreateBranchName string
	OutWriter        io.Writer
}

func Clone(opt CloneOptions) error {
	cmd := exec.Command("git", "clone", opt.RepoURL, opt.TargetPath)
	if o := opt.OutWriter; o != nil {
		cmd.Stdout = o
		cmd.Stderr = o
	}
	err := cmd.Run()
	if err != nil {
		return err
	}

	if opt.CreateBranchName != "" {
		// Create and checkout new branch
		cmd = exec.Command("git", "-C", opt.TargetPath, "checkout", "-b", opt.CreateBranchName)
		if o := opt.OutWriter; o != nil {
			cmd.Stdout = o
			cmd.Stderr = o
		}
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
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

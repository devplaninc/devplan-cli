package git

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
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
		if err := SetupBranch(opt.TargetPath, opt.CreateBranchName); err != nil {
			return err
		}
	}

	return nil
}

// SetupBranch checks if a remote branch exists and checks it out, otherwise creates a new branch.
// This is the isolated branch setup logic used after cloning a repository.
func SetupBranch(repoPath, branchName string) error {
	// Check if remote branch exists
	remoteExists, err := RemoteBranchExists(repoPath, branchName)
	if err != nil {
		return fmt.Errorf("failed to check remote branch: %w", err)
	}

	if remoteExists {
		// Remote branch exists - checkout from remote
		if err := CheckoutRemoteBranch(repoPath, branchName, "origin"); err != nil {
			return fmt.Errorf("failed to checkout remote branch: %w", err)
		}
		return nil
	}

	// Remote branch doesn't exist - create new branch
	if err := CreateAndCheckoutBranch(repoPath, branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	return nil
}

func GetRoot() (string, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return "", err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return "", err
	}
	return wt.Filesystem.Root(), nil
}

// LocalBranchExists checks if a local branch with the given name exists
func LocalBranchExists(repoPath, branchName string) (bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 1 {
				return false, nil // Branch doesn't exist
			}
		}
		return false, err // Actual error
	}
	return true, nil // Branch exists
}

// RemoteBranchExists checks if a remote branch with the given name exists on origin
func RemoteBranchExists(repoPath, branchName string) (bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branchName)
	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 1 {
				return false, nil // Branch doesn't exist
			}
		}
		return false, err // Actual error
	}
	return true, nil // Branch exists
}

// FetchRemote fetches updates from a remote to ensure remote refs are up-to-date
func FetchRemote(repoPath, remoteName string) error {
	cmd := exec.Command("git", "-C", repoPath, "fetch", remoteName)
	return cmd.Run()
}

// CheckoutLocalBranch checks out an existing local branch
func CheckoutLocalBranch(repoPath, branchName string) error {
	cmd := exec.Command("git", "-C", repoPath, "checkout", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// CheckoutRemoteBranch creates and checks out a local branch that tracks a remote branch
func CheckoutRemoteBranch(repoPath, branchName, remoteName string) error {
	cmd := exec.Command("git", "-C", repoPath, "checkout", "-b", branchName, remoteName+"/"+branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// CreateAndCheckoutBranch creates a new branch and checks it out
func CreateAndCheckoutBranch(repoPath, branchName string) error {
	cmd := exec.Command("git", "-C", repoPath, "checkout", "-b", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// isValidBranchName checks if the branch name follows git naming rules
func isValidBranchName(name string) bool {
	if name == "" {
		return false
	}
	// Git branch name restrictions
	if strings.HasPrefix(name, "-") || strings.HasPrefix(name, ".") {
		return false
	}
	if strings.HasSuffix(name, ".") || strings.HasSuffix(name, ".lock") {
		return false
	}
	// Invalid characters and sequences
	invalidChars := []string{" ", "~", "^", ":", "?", "*", "[", "\\", ".."}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return false
		}
	}
	return true
}

// CheckoutBranch orchestrates the branch checkout logic based on branch existence.
// It handles the following scenarios:
// - Remote branch exists: creates local tracking branch
// - Only local branch exists: checks out local branch
// - Neither exists: creates new local branch
// Note: If already on the requested branch, git checkout succeeds (no-op).
// If checkout fails due to uncommitted changes, the git error is returned.
func CheckoutBranch(repoPath, branchName string) error {
	// Validate branch name
	if !isValidBranchName(branchName) {
		return fmt.Errorf("invalid branch name: %s", branchName)
	}

	// Fetch from origin to update remote refs
	if err := FetchRemote(repoPath, "origin"); err != nil {
		fmt.Println(out.Warnf("Could not fetch from remote, using local refs: %v", err))
		// Continue anyway - remote might not be available but local branches still work
	}

	// Check if remote branch exists
	remoteExists, err := RemoteBranchExists(repoPath, branchName)
	if err != nil {
		return fmt.Errorf("failed to check remote branch: %w", err)
	}

	if remoteExists {
		// Check if local branch already exists (tracking the remote)
		localExists, err := LocalBranchExists(repoPath, branchName)
		if err != nil {
			return fmt.Errorf("failed to check local branch: %w", err)
		}

		if localExists {
			// Local branch exists, just checkout
			if err := CheckoutLocalBranch(repoPath, branchName); err != nil {
				return err
			}
			out.Psuccessf("Checked out branch %s\n", out.H(branchName))
		} else {
			// Create local tracking branch from remote
			if err := CheckoutRemoteBranch(repoPath, branchName, "origin"); err != nil {
				return err
			}
			out.Psuccessf("Checked out branch %s from remote origin\n", out.H(branchName))
		}
		return nil
	}

	// Remote doesn't exist, check local
	localExists, err := LocalBranchExists(repoPath, branchName)
	if err != nil {
		return fmt.Errorf("failed to check local branch: %w", err)
	}

	if localExists {
		// Checkout local branch
		if err := CheckoutLocalBranch(repoPath, branchName); err != nil {
			return err
		}
		out.Psuccessf("Checked out local branch %s (note: no remote branch exists)\n", out.H(branchName))
		return nil
	}

	// Neither local nor remote exists, create new branch
	if err := CreateAndCheckoutBranch(repoPath, branchName); err != nil {
		return err
	}
	out.Psuccessf("Created and checked out new branch %s\n", out.H(branchName))
	return nil
}

func getRepoURLs(path string) ([]string, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
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

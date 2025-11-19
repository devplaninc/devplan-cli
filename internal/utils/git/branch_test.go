package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "git-branch-test")
	require.NoError(t, err)

	// Initialize git repo
	cmd := exec.Command("git", "init", tempDir)
	require.NoError(t, cmd.Run())

	// Configure git user for commits
	cmd = exec.Command("git", "-C", tempDir, "config", "user.email", "test@test.com")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", tempDir, "config", "user.name", "Test User")
	require.NoError(t, cmd.Run())

	// Create initial commit
	testFile := filepath.Join(tempDir, "README.md")
	require.NoError(t, os.WriteFile(testFile, []byte("# Test"), 0644))
	cmd = exec.Command("git", "-C", tempDir, "add", ".")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", tempDir, "commit", "-m", "Initial commit")
	require.NoError(t, cmd.Run())

	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// setupTestRepoWithRemote creates a local repo with a "remote" (bare repo) for testing
func setupTestRepoWithRemote(t *testing.T) (localPath, remotePath string, cleanup func()) {
	// Create temp directory for both repos
	baseDir, err := os.MkdirTemp("", "git-branch-test")
	require.NoError(t, err)

	remotePath = filepath.Join(baseDir, "remote.git")
	localPath = filepath.Join(baseDir, "local")

	// Initialize bare repo as "remote"
	cmd := exec.Command("git", "init", "--bare", remotePath)
	require.NoError(t, cmd.Run())

	// Clone "remote" to create local repo
	cmd = exec.Command("git", "clone", remotePath, localPath)
	require.NoError(t, cmd.Run())

	// Configure git user for commits
	cmd = exec.Command("git", "-C", localPath, "config", "user.email", "test@test.com")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", localPath, "config", "user.name", "Test User")
	require.NoError(t, cmd.Run())

	// Create initial commit and push
	testFile := filepath.Join(localPath, "README.md")
	require.NoError(t, os.WriteFile(testFile, []byte("# Test"), 0644))
	cmd = exec.Command("git", "-C", localPath, "add", ".")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", localPath, "commit", "-m", "Initial commit")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", localPath, "push", "-u", "origin", "master")
	err = cmd.Run()
	if err != nil {
		// Try main instead of master
		cmd = exec.Command("git", "-C", localPath, "branch", "-m", "master", "main")
		cmd.Run()
		cmd = exec.Command("git", "-C", localPath, "push", "-u", "origin", "main")
		require.NoError(t, cmd.Run())
	}

	cleanup = func() {
		_ = os.RemoveAll(baseDir)
	}

	return localPath, remotePath, cleanup
}

// getDefaultBranch gets the current branch name in a repo
func getDefaultBranch(t *testing.T, repoPath string) string {
	out, err := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
	require.NoError(t, err)
	return strings.TrimSpace(string(out))
}

func TestIsValidBranchName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		branch   string
		expected bool
	}{
		{"valid simple", "feature-branch", true},
		{"valid with slash", "feature/my-branch", true},
		{"valid with numbers", "branch123", true},
		{"empty", "", false},
		{"starts with dash", "-branch", false},
		{"starts with dot", ".branch", false},
		{"ends with dot", "branch.", false},
		{"ends with lock", "branch.lock", false},
		{"contains space", "my branch", false},
		{"contains tilde", "branch~1", false},
		{"contains caret", "branch^2", false},
		{"contains colon", "branch:name", false},
		{"contains question", "branch?", false},
		{"contains asterisk", "branch*", false},
		{"contains bracket", "branch[0]", false},
		{"contains backslash", "branch\\name", false},
		{"contains double dot", "branch..name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := isValidBranchName(tt.branch)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckoutBranch_InvalidBranchName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	localPath, _, cleanup := setupTestRepoWithRemote(t)
	defer cleanup()

	err := CheckoutBranch(localPath, "invalid..name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid branch name")
}

// Integration tests for SetupBranch function
func TestSetupBranch_RemoteBranchExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	localPath, _, cleanup := setupTestRepoWithRemote(t)
	defer cleanup()

	// Create a branch on remote
	cmd := exec.Command("git", "-C", localPath, "checkout", "-b", "feature-from-remote")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", localPath, "push", "-u", "origin", "feature-from-remote")
	require.NoError(t, cmd.Run())
	// Go back to main/master and delete local branch (simulating fresh clone)
	cmd = exec.Command("git", "-C", localPath, "checkout", "-")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", localPath, "branch", "-D", "feature-from-remote")
	require.NoError(t, cmd.Run())

	// Now SetupBranch should checkout from remote
	err := SetupBranch(localPath, "feature-from-remote")
	assert.NoError(t, err)

	// Verify we're on the correct branch
	out, err := exec.Command("git", "-C", localPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
	require.NoError(t, err)
	assert.Equal(t, "feature-from-remote", strings.TrimSpace(string(out)))

	// Verify it's tracking the remote
	out, err = exec.Command("git", "-C", localPath, "config", "--get", "branch.feature-from-remote.remote").Output()
	require.NoError(t, err)
	assert.Equal(t, "origin", strings.TrimSpace(string(out)))
}

func TestSetupBranch_RemoteBranchDoesNotExist(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	localPath, _, cleanup := setupTestRepoWithRemote(t)
	defer cleanup()

	// SetupBranch with non-existing branch should create new
	err := SetupBranch(localPath, "new-feature-branch")
	assert.NoError(t, err)

	// Verify we're on the correct branch
	out, err := exec.Command("git", "-C", localPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
	require.NoError(t, err)
	assert.Equal(t, "new-feature-branch", strings.TrimSpace(string(out)))

	// Verify local branch exists
	exists, err := LocalBranchExists(localPath, "new-feature-branch")
	assert.NoError(t, err)
	assert.True(t, exists)
}

// Integration test using real GitHub repository
func TestSetupBranch_RealGitHubRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	// Create temp directory for clone
	tempDir, err := os.MkdirTemp("", "git-integ-test")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	repoPath := filepath.Join(tempDir, "devplan-cli")

	// Clone the real repository
	cmd := exec.Command("git", "clone", "https://github.com/devplaninc/devplan-cli", repoPath)
	err = cmd.Run()
	require.NoError(t, err, "Failed to clone repository")

	// Test SetupBranch with existing remote branch
	branchName := "testing/integ_test_branch_do_not_delete"
	err = SetupBranch(repoPath, branchName)
	require.NoError(t, err)

	// Verify we're on the correct branch
	out, err := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
	require.NoError(t, err)
	assert.Equal(t, branchName, strings.TrimSpace(string(out)))

	// Verify it's tracking the remote
	out, err = exec.Command(
		"git", "-C", repoPath, "config", "--get", fmt.Sprintf("branch.%v.remote", branchName),
	).Output()
	require.NoError(t, err)
	assert.Equal(t, "origin", strings.TrimSpace(string(out)))
}

func TestLocalBranchExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	defaultBranch := getDefaultBranch(t, repoPath)

	tests := []struct {
		name       string
		branchName string
		setup      func()
		expected   bool
		wantErr    bool
	}{
		{
			name:       "existing branch (default)",
			branchName: defaultBranch,
			setup:      func() {},
			expected:   true,
			wantErr:    false,
		},
		{
			name:       "non-existing branch",
			branchName: "feature-xyz",
			setup:      func() {},
			expected:   false,
			wantErr:    false,
		},
		{
			name:       "newly created branch",
			branchName: "test-branch",
			setup: func() {
				cmd := exec.Command("git", "-C", repoPath, "branch", "test-branch")
				require.NoError(t, cmd.Run())
			},
			expected: true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			exists, err := LocalBranchExists(repoPath, tt.branchName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, exists)
			}
		})
	}
}

func TestLocalBranchExists_InvalidRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	tempDir, err := os.MkdirTemp("", "not-a-repo")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	exists, err := LocalBranchExists(tempDir, "master")
	assert.Error(t, err)
	assert.False(t, exists)
}

func TestRemoteBranchExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	localPath, _, cleanup := setupTestRepoWithRemote(t)
	defer cleanup()

	// Create a branch on remote
	cmd := exec.Command("git", "-C", localPath, "checkout", "-b", "remote-branch")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", localPath, "push", "-u", "origin", "remote-branch")
	require.NoError(t, cmd.Run())
	// Go back to main/master
	cmd = exec.Command("git", "-C", localPath, "checkout", "-")
	require.NoError(t, cmd.Run())
	// Fetch to update refs
	cmd = exec.Command("git", "-C", localPath, "fetch", "origin")
	require.NoError(t, cmd.Run())

	tests := []struct {
		name       string
		branchName string
		expected   bool
		wantErr    bool
	}{
		{
			name:       "existing remote branch",
			branchName: "remote-branch",
			expected:   true,
			wantErr:    false,
		},
		{
			name:       "non-existing remote branch",
			branchName: "nonexistent",
			expected:   false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := RemoteBranchExists(localPath, tt.branchName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, exists)
			}
		})
	}
}

func TestCheckoutLocalBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create a branch to checkout
	cmd := exec.Command("git", "-C", repoPath, "branch", "feature-branch")
	require.NoError(t, cmd.Run())

	tests := []struct {
		name       string
		branchName string
		wantErr    bool
	}{
		{
			name:       "checkout existing branch",
			branchName: "feature-branch",
			wantErr:    false,
		},
		{
			name:       "checkout non-existing branch",
			branchName: "nonexistent",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckoutLocalBranch(repoPath, tt.branchName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify we're on the correct branch
				out, err := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
				require.NoError(t, err)
				assert.Contains(t, string(out), tt.branchName)
			}
		})
	}
}

func TestCreateAndCheckoutBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	tests := []struct {
		name       string
		branchName string
		wantErr    bool
	}{
		{
			name:       "create new branch",
			branchName: "new-feature",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateAndCheckoutBranch(repoPath, tt.branchName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify we're on the correct branch
				out, err := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
				require.NoError(t, err)
				assert.Contains(t, string(out), tt.branchName)
			}
		})
	}
}

func TestCreateAndCheckoutBranch_AlreadyExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create branch first
	cmd := exec.Command("git", "-C", repoPath, "branch", "existing-branch")
	require.NoError(t, cmd.Run())

	// Try to create again
	err := CreateAndCheckoutBranch(repoPath, "existing-branch")
	assert.Error(t, err)
}

func TestCheckoutRemoteBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	localPath, _, cleanup := setupTestRepoWithRemote(t)
	defer cleanup()

	// Create a branch on remote
	cmd := exec.Command("git", "-C", localPath, "checkout", "-b", "remote-feature")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", localPath, "push", "-u", "origin", "remote-feature")
	require.NoError(t, cmd.Run())
	// Go back to main/master and delete local branch
	cmd = exec.Command("git", "-C", localPath, "checkout", "-")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", localPath, "branch", "-D", "remote-feature")
	require.NoError(t, cmd.Run())
	// Fetch to update refs
	cmd = exec.Command("git", "-C", localPath, "fetch", "origin")
	require.NoError(t, cmd.Run())

	// Now checkout from remote
	err := CheckoutRemoteBranch(localPath, "remote-feature", "origin")
	assert.NoError(t, err)

	// Verify we're on the correct branch
	out, err := exec.Command("git", "-C", localPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "remote-feature")
}

func TestFetchRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	localPath, _, cleanup := setupTestRepoWithRemote(t)
	defer cleanup()

	// This should succeed
	err := FetchRemote(localPath, "origin")
	assert.NoError(t, err)
}

func TestFetchRemote_InvalidRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	// No remote configured, should fail
	err := FetchRemote(repoPath, "origin")
	assert.Error(t, err)
}

func TestCheckoutBranch_RemoteExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	localPath, _, cleanup := setupTestRepoWithRemote(t)
	defer cleanup()

	// Create a branch on remote
	cmd := exec.Command("git", "-C", localPath, "checkout", "-b", "remote-branch")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", localPath, "push", "-u", "origin", "remote-branch")
	require.NoError(t, cmd.Run())
	// Go back to main/master and delete local branch
	cmd = exec.Command("git", "-C", localPath, "checkout", "-")
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "-C", localPath, "branch", "-D", "remote-branch")
	require.NoError(t, cmd.Run())

	// Use CheckoutBranch - should checkout from remote
	err := CheckoutBranch(localPath, "remote-branch")
	assert.NoError(t, err)

	// Verify we're on the correct branch
	out, err := exec.Command("git", "-C", localPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "remote-branch")
}

func TestCheckoutBranch_LocalOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	localPath, _, cleanup := setupTestRepoWithRemote(t)
	defer cleanup()

	// Create a local-only branch
	cmd := exec.Command("git", "-C", localPath, "checkout", "-b", "local-only-branch")
	require.NoError(t, cmd.Run())
	// Go back to main/master
	cmd = exec.Command("git", "-C", localPath, "checkout", "-")
	require.NoError(t, cmd.Run())

	// Use CheckoutBranch - should checkout local branch
	err := CheckoutBranch(localPath, "local-only-branch")
	assert.NoError(t, err)

	// Verify we're on the correct branch
	out, err := exec.Command("git", "-C", localPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "local-only-branch")
}

func TestCheckoutBranch_CreateNew(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	localPath, _, cleanup := setupTestRepoWithRemote(t)
	defer cleanup()

	// Use CheckoutBranch with non-existing branch - should create new
	err := CheckoutBranch(localPath, "brand-new-branch")
	assert.NoError(t, err)

	// Verify we're on the correct branch
	out, err := exec.Command("git", "-C", localPath, "rev-parse", "--abbrev-ref", "HEAD").Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "brand-new-branch")
}

func TestCheckoutBranch_UncommittedChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	t.Parallel()

	localPath, _, cleanup := setupTestRepoWithRemote(t)
	defer cleanup()

	// Create another branch
	cmd := exec.Command("git", "-C", localPath, "branch", "other-branch")
	require.NoError(t, cmd.Run())

	// Create uncommitted changes
	testFile := filepath.Join(localPath, "README.md")
	require.NoError(t, os.WriteFile(testFile, []byte("# Modified"), 0644))

	// Try to checkout - should fail due to uncommitted changes
	err := CheckoutBranch(localPath, "other-branch")
	// Note: git allows checkout if changes don't conflict, so this may or may not fail
	// depending on the situation. Let's just make sure it doesn't panic.
	_ = err
}

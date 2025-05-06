package cmd

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
)

var (
	repoURL      string
	repoPath     string
	repoBranch   string
	repoCheckout bool

	repoCmd = &cobra.Command{
		Use:   "repo",
		Short: "Manage git repositories",
		Long:  `Clone, update, or checkout git repositories for your development workflow.`,
	}

	repoCloneCmd = &cobra.Command{
		Use:   "clone",
		Short: "Clone a git repository",
		Long:  `Clone a git repository to a specified path.`,
		Run:   runRepoClone,
	}

	repoUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update a git repository",
		Long:  `Update (pull) the latest changes from a git repository.`,
		Run:   runRepoUpdate,
	}
)

//func init() {
//	rootCmd.AddCommand(repoCmd)
//	repoCmd.AddCommand(repoCloneCmd)
//	repoCmd.AddCommand(repoUpdateCmd)
//
//	// Clone command flags
//	repoCloneCmd.Flags().StringVarP(&repoURL, "url", "u", "", "URL of the git repository to clone (required)")
//	repoCloneCmd.Flags().StringVarP(&repoPath, "path", "p", "", "Path where the repository should be cloned (required)")
//	repoCloneCmd.Flags().StringVarP(&repoBranch, "branch", "b", "", "Branch to checkout (default is repository default branch)")
//	repoCloneCmd.MarkFlagRequired("url")
//	repoCloneCmd.MarkFlagRequired("path")
//
//	// Update command flags
//	repoUpdateCmd.Flags().StringVarP(&repoPath, "path", "p", "", "Path to the repository to update (required)")
//	repoUpdateCmd.Flags().StringVarP(&repoBranch, "branch", "b", "", "Branch to checkout after update")
//	repoUpdateCmd.MarkFlagRequired("path")
//}

func runRepoClone(cmd *cobra.Command, args []string) {
	fmt.Printf("Cloning repository from %s to %s...\n", repoURL, repoPath)

	// Create directory if it doesn't exist
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		err = os.MkdirAll(repoPath, 0755)
		if err != nil {
			fmt.Printf("Failed to create directory: %v\n", err)
			return
		}
	}

	// Clone options
	cloneOptions := &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	}

	// Clone the repository
	repo, err := git.PlainClone(repoPath, false, cloneOptions)
	if err != nil {
		fmt.Printf("Failed to clone repository: %v\n", err)
		return
	}

	// Checkout specific branch if specified
	if repoBranch != "" {
		w, err := repo.Worktree()
		if err != nil {
			fmt.Printf("Failed to get worktree: %v\n", err)
			return
		}

		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(repoBranch),
		})
		if err != nil {
			fmt.Printf("Failed to checkout branch %s: %v\n", repoBranch, err)
			return
		}
		fmt.Printf("Checked out branch: %s\n", repoBranch)
	}

	fmt.Println("Repository cloned successfully!")
}

func runRepoUpdate(cmd *cobra.Command, args []string) {
	fmt.Printf("Updating repository at %s...\n", repoPath)

	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		fmt.Printf("Failed to open repository: %v\n", err)
		return
	}

	// Get the worktree
	w, err := repo.Worktree()
	if err != nil {
		fmt.Printf("Failed to get worktree: %v\n", err)
		return
	}

	// Pull the latest changes
	err = w.Pull(&git.PullOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		fmt.Printf("Failed to pull repository: %v\n", err)
		return
	}

	if err == git.NoErrAlreadyUpToDate {
		fmt.Println("Repository is already up to date.")
	} else {
		fmt.Println("Repository updated successfully!")
	}

	// Checkout specific branch if specified
	if repoBranch != "" {
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(repoBranch),
		})
		if err != nil {
			fmt.Printf("Failed to checkout branch %s: %v\n", repoBranch, err)
			return
		}
		fmt.Printf("Checked out branch: %s\n", repoBranch)
	}
}

package gitws

import (
	"testing"

	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/devplaninc/devplan-cli/internal/utils/picker"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"github.com/stretchr/testify/assert"
)

func TestGenerateMetadata(t *testing.T) {
	repo := git.RepoInfo{
		URLs:      []string{"https://github.com/org/repo.git"},
		FullNames: []string{"org/repo"},
	}

	project := &documents.ProjectEntity{}
	project.SetId("proj-1")
	project.SetTitle("Project 1")
	project.SetNumericId(101)

	target := picker.DevTarget{
		ProjectWithDocs: &documents.ProjectWithDocs{},
	}
	target.ProjectWithDocs.SetProject(project)

	t.Run("project metadata only", func(t *testing.T) {
		meta := generateMetadata(repo, target, false)
		assert.Equal(t, "proj-1", meta.ProjectID)
		assert.Equal(t, "Project 1", meta.ProjectName)
		assert.Equal(t, "101", meta.ProjectNumericID)
		assert.Equal(t, "https://github.com/org/repo.git", meta.RepoURL)
		assert.Equal(t, "org/repo", meta.RepoName)
		assert.Empty(t, meta.TaskID)
	})

	t.Run("worktree metadata with task and story", func(t *testing.T) {
		story := &documents.DocumentEntity{}
		story.SetId("story-1")
		story.SetTitle("Story 1")
		story.SetNumericId(201)
		story.SetType(documents.DocumentType_FEATURE)

		task := &documents.DocumentEntity{}
		task.SetId("task-1")
		task.SetTitle("Task 1")
		task.SetNumericId(301)
		task.SetParentId("story-1")
		task.SetType(documents.DocumentType_TASK)

		target.Task = task
		target.ProjectWithDocs.SetDocs([]*documents.DocumentEntity{story, task})

		meta := generateMetadata(repo, target, true)
		assert.Equal(t, "proj-1", meta.ProjectID)
		assert.Equal(t, "101", meta.ProjectNumericID)
		assert.Equal(t, "task-1", meta.TaskID)
		assert.Equal(t, "Task 1", meta.TaskName)
		assert.Equal(t, "301", meta.TaskNumericID)
		assert.Equal(t, "story-1", meta.StoryID)
		assert.Equal(t, "Story 1", meta.StoryName)
		assert.Equal(t, "201", meta.StoryNumericID)
	})

	t.Run("worktree metadata with feature", func(t *testing.T) {
		feature := &documents.DocumentEntity{}
		feature.SetId("feat-1")
		feature.SetTitle("Feature 1")
		feature.SetNumericId(401)
		feature.SetType(documents.DocumentType_FEATURE)

		target.Task = nil
		target.SpecificFeature = feature

		meta := generateMetadata(repo, target, true)
		assert.Equal(t, "proj-1", meta.ProjectID)
		assert.Equal(t, "101", meta.ProjectNumericID)
		assert.Equal(t, "feat-1", meta.StoryID)
		assert.Equal(t, "Feature 1", meta.StoryName)
		assert.Equal(t, "401", meta.StoryNumericID)
		assert.Empty(t, meta.TaskID)
	})
}

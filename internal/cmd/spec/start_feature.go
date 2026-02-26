package spec

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/converters"
	"github.com/devplaninc/devplan-cli/internal/utils/gitws"
	"github.com/devplaninc/devplan-cli/internal/utils/metadata"
	"github.com/devplaninc/devplan-cli/internal/utils/workspace"
	"github.com/opensdd/osdd-api/clients/go/osdd/recipes"
)

type startFeatureResult struct {
	WorkspacePath string
	ExecRecipe    *recipes.ExecutableRecipe
}

func runStartFeature(ctx context.Context, cl *devplan.Client, companyID int32, featureID string, path string) startFeatureResult {
	docResp, err := cl.GetDocument(companyID, featureID)
	check(err)
	feature := docResp.GetDocument()

	details, err := converters.GetFeatureDetails(feature)
	check(err)
	if details == nil || len(details.GetRepoNames()) == 0 {
		check(fmt.Errorf("feature has no repositories configured"))
	}

	workspacePath := path
	if workspacePath == "" {
		project, err := resolveProjectInfo(cl, companyID, feature.GetProjectId())
		check(err)

		sanitizedProject := gitws.SanitizeName(project.Name, 30)
		sanitizedFeature := gitws.SanitizeName(feature.GetTitle(), 30)
		parentPath := workspace.GetFeatureWorkspacePath(sanitizedProject, sanitizedFeature)

		repos, err := gitws.ResolveRepos(details.GetRepoNames(), companyID)
		check(err)

		slog.Info("Cloning repositories for feature", "feature", feature.GetTitle(), "count", len(repos))
		cloneResult, err := gitws.CloneAllRepos(ctx, repos, parentPath, "feature/"+sanitizedFeature)
		check(err)

		out.Psuccessf("All %d repositories cloned successfully\n", len(repos))

		meta := metadata.Metadata{
			ProjectID:        feature.GetProjectId(),
			ProjectName:      project.Name,
			ProjectNumericID: fmt.Sprintf("%v", project.NumericID),
			StoryID:          feature.GetId(),
			StoryName:        feature.GetTitle(),
			StoryNumericID:   fmt.Sprintf("%v", feature.GetNumericId()),
		}
		if err := metadata.EnsureMetadataSetup(cloneResult.ParentPath, meta); err != nil {
			slog.Warn("Failed to setup feature workspace metadata", "err", err)
		}

		workspacePath = cloneResult.ParentPath
	}

	execRecipe, err := cl.GetFeatureExecRecipe(companyID, featureID)
	check(err)

	return startFeatureResult{
		WorkspacePath: workspacePath,
		ExecRecipe:    execRecipe,
	}
}

type resolvedProject struct {
	Name      string
	NumericID int32
}

func resolveProjectInfo(cl *devplan.Client, companyID int32, projectID string) (resolvedProject, error) {
	prResp, err := cl.GetCompanyProjects(companyID)
	if err != nil {
		return resolvedProject{}, fmt.Errorf("failed to get company projects: %w", err)
	}
	for _, p := range prResp.GetProjects() {
		if p.GetProject().GetId() == projectID {
			return resolvedProject{
				Name:      p.GetProject().GetTitle(),
				NumericID: p.GetProject().GetNumericId(),
			}, nil
		}
	}
	return resolvedProject{}, fmt.Errorf("project %s not found", projectID)
}

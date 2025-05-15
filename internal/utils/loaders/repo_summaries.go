package loaders

import (
	"context"
	"github.com/devplaninc/devplan-cli/internal/components/spinner"
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/devplaninc/devplan-cli/internal/utils/picker"
	company2 "github.com/devplaninc/webapp/golang/pb/api/devplan/services/web/company"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/artifacts"
)

type summariesResult struct {
	resp *company2.GetRepoSummariesResponse
	err  error
}

func RepoSummary(target picker.DevTarget, repo git.RepoInfo) (*artifacts.ArtifactRepoSummary, error) {
	ctx, cancel := context.WithCancel(context.Background())

	cl := devplan.NewClient(devplan.Config{})
	sumRespChan := make(chan summariesResult, 1)
	go func() {
		defer cancel()
		sumResp, err := cl.GetRepoSummaries(target.ProjectWithDocs.GetProject().GetCompanyId())
		if err != nil {
			sumRespChan <- summariesResult{err: err}
			return
		}
		sumRespChan <- summariesResult{resp: sumResp}
	}()
	err := spinner.Run(ctx, "Loading repo summaries", "Repo summaries loaded")
	if err != nil {
		return nil, err
	}
	res := <-sumRespChan
	if err := res.err; err != nil {
		return nil, err
	}
	return getMatchingSummary(repo, res.resp.GetSummaries()), nil
}

func getMatchingSummary(repo git.RepoInfo, summaries []*artifacts.ArtifactRepoSummary) *artifacts.ArtifactRepoSummary {
	for _, s := range summaries {
		if repo.MatchesName(s.GetRepoName()) {
			return s
		}
	}
	return nil
}

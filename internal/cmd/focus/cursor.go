package focus

import (
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/devplan-cli/internal/utils/ide/cursor"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/artifacts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"path/filepath"
)

func createCursorRules(featurePrompt *documents.DocumentEntity, repoSummary *artifacts.ArtifactRepoSummary) error {
	rules := []ide.Rule{
		{Name: "flow", Content: devFlowRule, Header: cursor.GetRuleHeader(cursor.Header{
			Description: devFlowRuleDescription,
		})},
		{Name: "rules", Content: rulesRule, Header: cursor.GetRuleHeader(cursor.Header{
			Description: generalRuleDescription, Globs: ".cursor/rules/*.mdc",
		})},
		{Name: "insights", Content: insightsRule, Header: cursor.GetRuleHeader(cursor.Header{
			Description: insightsRuleDescription,
		})},
	}
	if featurePrompt != nil {
		rules = append(rules, ide.Rule{
			Name:    "current_feature",
			Content: featurePrompt.GetContent(),
			Header: cursor.GetRuleHeader(cursor.Header{
				Description: currentFeatRuleDescription,
			}),
		})
	}
	if summary := repoSummary.GetSummary().GetContent(); summary != "" {
		rules = append(rules, ide.Rule{Name: "repo_overview", Content: summary, Header: cursor.GetRuleHeader(
			cursor.Header{Description: repoOverviewRuleDescription},
		)})
	}

	return ide.WriteRules(rules, filepath.Join(".cursor", "rules"), "mdc")
}

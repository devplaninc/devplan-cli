package ide

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/utils/prompts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/artifacts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"strings"
)

func mdAllOtherRules() string {
	return fmt.Sprintf(`Also refer to the following files for the details when needed:

- [Insights](devplan_insights.md) - %s
- [Rules](devplan_rules.md) - %s
- [Repo Overview](devplan_repo_overview.md) - %s (if present)
- [Current Feature](devplan_current_feature.md) - %s. Always review current feature if it is present.
`, insightsRuleDescription, generalRuleDescription, repoOverviewRuleDescription, currentFeatRuleDescription)
}

var mdReplacements = map[string]string{
	"@devplan_current_feature.mdc": "%v/devplan_current_feature.md",
	"@devplan_rules.mdc":           "%v/devplan_rules.md",
	"@devplan_insights.mdc":        "%v/devplan_insights.md",
	"@devplan_repo_overview.mdc":   "%v/devplan_repo_overview.md",
	"@devplan_tests.mdc":           "%v/devplan_tests.md",
}

func mdPathsReplace(str string, rulesPath string) string {
	result := str
	for k, v := range mdReplacements {
		vStr := fmt.Sprintf(v, rulesPath)
		result = strings.ReplaceAll(result, k, vStr)
	}
	return result
}

func createMdRules(rulesPath string, featurePrompt *documents.DocumentEntity, repoSummary *artifacts.ArtifactRepoSummary) error {
	rules := []Rule{
		{NoPrefix: true, Name: "guidelines", Content: mdPathsReplace(devFlowRule, rulesPath), Footer: mdAllOtherRules()},
		{Name: "rules", Content: rulesRule},
		{Name: "insights", Content: insightsRule},
	}
	if featurePrompt != nil {
		targetID, err := prompts.GetTargetID(featurePrompt)
		if err != nil {
			return fmt.Errorf("failed to get feature ID for feature prompt: %w", err)
		}
		rules = append(rules, Rule{
			Name:     "current_feature",
			Content:  featurePrompt.GetContent(),
			TargetID: targetID,
		})
	}
	if summary := repoSummary.GetSummary().GetContent(); summary != "" {
		rules = append(rules, Rule{Name: "repo_overview", Content: summary})
	}
	return WriteRules(rules, rulesPath, "md")
}

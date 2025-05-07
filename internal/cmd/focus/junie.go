package focus

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/artifacts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"path/filepath"
	"strings"
)

func junieAllOtherRules() string {
	return fmt.Sprintf(`Also refer to the following files for the details when needed:

- [Insights](devplan_insights.md) - %s
- [Rules](devplan_rules.md) - %s
- [Repo Overview](devplan_repo_overview.md) - %s (if present)
- [Current Feature](devplan_current_feature.md) - %s. Always review current feature if it is present.
`, insightsRuleDescription, generalRuleDescription, repoOverviewRuleDescription, currentFeatRuleDescription)
}

var replacements = map[string]string{
	"@devplan_current_feature.mdc": ".junie/devplan_current_feature.md",
	"@devplan_rules.mdc":           ".junie/devplan_rules.md",
	"@devplan_insights.mdc":        ".junie/devplan_insights.md",
	"@devplan_repo_overview.mdc":   ".junie/devplan_repo_overview.md",
	"@devplan_tests.mdc":           ".junie/devplan_tests.md",
}

func juniePathsReplace(str string) string {
	result := str
	for k, v := range replacements {
		result = strings.ReplaceAll(result, k, v)
	}
	return result
}

func createJunieRules(featurePrompt *documents.DocumentEntity, repoSummary *artifacts.ArtifactRepoSummary) error {
	rules := []ide.Rule{
		{NoPrefix: true, Name: "guidelines", Content: juniePathsReplace(devFlowRule), Footer: junieAllOtherRules()},
		{Name: "rules", Content: rulesRule},
		{Name: "insights", Content: insightsRule},
	}
	if featurePrompt != nil {
		rules = append(rules, ide.Rule{Name: "current_feature", Content: featurePrompt.GetContent()})
	}
	if summary := repoSummary.GetSummary().GetContent(); summary != "" {
		rules = append(rules, ide.Rule{Name: "repo_overview", Content: summary})
	}
	return ide.WriteRules(rules, filepath.Join(".junie"), "md")
}

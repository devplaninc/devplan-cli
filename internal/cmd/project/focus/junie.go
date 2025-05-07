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

func juniePathsReplace(str string) string {
	return strings.ReplaceAll(str, "@devplan_current_feature.mdc", ".junie/devplan_current_feature.md")
}

func createJunieRules(featurePrompt *documents.DocumentEntity, repoSummary *artifacts.ArtifactRepoSummary) error {
	if err := confirmRulesGeneration("Junie", featurePrompt, repoSummary); err != nil {
		return err
	}
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

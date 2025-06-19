package ide

import (
	"bytes"
	"fmt"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/integrations"
	"strings"
	"text/template"
)

func allOtherRulesSuffix(rulesPath string, ext string) string {
	const tmpl = `Also refer to the following files for the details when needed:

- [Insights]({{.Path}}/devplan_insights.{{.Ext}}) - {{.Insights}}
- [Rules]({{.Path}}/devplan_rules.{{.Ext}}) - {{.Rules}}
- [Repo Overview]({{.Path}}/devplan_repo_overview.{{.Ext}}) - {{.Overview}} (if present)
- [Current Feature]({{.Path}}/devplan_current_feature.{{.Ext}}) - {{.Feature}}. Always review current feature if it is present.
`
	data := struct {
		Path     string
		Insights string
		Rules    string
		Overview string
		Feature  string
		Ext      string
	}{
		Path:     rulesPath,
		Insights: insightsRuleDescription,
		Rules:    generalRuleDescription,
		Overview: repoOverviewRuleDescription,
		Feature:  currentFeatRuleDescription,
		Ext:      ext,
	}

	t := template.Must(template.New("rules").Parse(tmpl))
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

func replacePaths(str string, rulesPath string, ext string) string {
	mdReplacements := map[string]string{
		"@devplan_current_feature.mdc": "%v/devplan_current_feature." + ext,
		"@devplan_rules.mdc":           "%v/devplan_rules." + ext,
		"@devplan_insights.mdc":        "%v/devplan_insights." + ext,
		"@devplan_repo_overview.mdc":   "%v/devplan_repo_overview." + ext,
		"@devplan_tests.mdc":           "%v/devplan_tests." + ext,
	}
	result := str
	for k, v := range mdReplacements {
		vStr := fmt.Sprintf(v, rulesPath)
		result = strings.ReplaceAll(result, k, vStr)
	}
	return result
}

func createMdRules(rulesPath string, featurePrompt *documents.DocumentEntity, repoSummary *integrations.RepositorySummary) error {
	rules := []Rule{
		{NoPrefix: true, Name: "guidelines",
			Content: replacePaths(devFlowRule, rulesPath, "md"),
			Footer:  allOtherRulesSuffix(".", "md"),
		},
		{Name: "rules", Content: rulesRule},
		{Name: "insights", Content: insightsRule},
	}
	if featurePrompt != nil {
		cfRules, err := generateCurrentFeatureRules(
			rulePaths{dir: rulesPath, ext: "md"},
			Rule{},
			featurePrompt)
		if err != nil {
			return fmt.Errorf("failed to generate current feature rules: %w", err)
		}
		rules = append(rules, cfRules...)
	}
	if summary := repoSummary.GetSummary(); summary != "" {
		rules = append(rules, Rule{Name: "repo_overview", Content: summary})
	}
	return WriteRules(rules, rulesPath, "md")
}

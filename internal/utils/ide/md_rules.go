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

- [Repo Overview]({{.Path}}/devplan_repo_overview.{{.Ext}}) - {{.Overview}} (if present)
- [Current Feature]({{.Path}}/devplan_current_feature.{{.Ext}}) - {{.Feature}}. Always review current feature if it is present.
`
	data := struct {
		Path     string
		Overview string
		Feature  string
		Ext      string
	}{
		Path:     rulesPath,
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

type mdRulesParams struct {
	rulesPath     string
	featurePrompt *documents.DocumentEntity
	repoSummary   *integrations.RepositorySummary
	devFlowName   string
	devFlowPath   string
	backUpDevFlow bool
}

func createMdRules(p mdRulesParams) error {
	devFlowName := p.devFlowName
	if devFlowName == "" {
		devFlowName = "guidelines"
	}
	devFlowPath := p.devFlowPath
	if devFlowPath == "" {
		devFlowPath = p.rulesPath
	}
	guidelinesRule := Rule{
		NoPrefix:    true,
		Name:        devFlowName,
		Content:     replacePaths(devFlowRule, p.rulesPath, "md"),
		Footer:      allOtherRulesSuffix(p.rulesPath, "md"),
		NeedsBackup: p.backUpDevFlow,
	}
	var rules []Rule
	if p.featurePrompt != nil {
		cfRules, err := generateCurrentFeatureRules(
			rulePaths{dir: p.rulesPath, ext: "md"},
			Rule{},
			p.featurePrompt)
		if err != nil {
			return fmt.Errorf("failed to generate current feature rules: %w", err)
		}
		rules = append(rules, cfRules...)
	}
	if summary := p.repoSummary.GetSummary(); summary != "" {
		rules = append(rules, Rule{Name: "repo_overview", Content: summary})
	}
	if err := WriteRules(rules, p.rulesPath, "md"); err != nil {
		return err
	}
	if err := WriteRules([]Rule{guidelinesRule}, devFlowPath, "md"); err != nil {
		return err
	}
	return nil
}

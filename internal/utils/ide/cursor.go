package ide

import (
	"fmt"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/integrations"
)

func createCursorRulesFromPrompt(rulesPath string, prompt *documents.DocumentEntity, repoSummary *integrations.RepositorySummary) error {
	rules := []Rule{
		{Name: "flow",
			Content: replacePaths(devFlowRule, rulesPath, "mdc"),
			Header:  getCursorRuleHeader(CursorHeader{Description: devFlowRuleDescription}),
			Footer:  allOtherRulesSuffix(rulesPath, "mdc"),
		},
	}

	if prompt != nil {
		cfRules, err := generateCurrentFeatureRules(
			rulePaths{dir: rulesPath, ext: "mdc"},
			Rule{
				Header: getCursorRuleHeader(CursorHeader{Description: currentFeatRuleDescription}),
			},
			prompt)
		if err != nil {
			return fmt.Errorf("failed to generate current feature rules: %w", err)
		}
		rules = append(rules, cfRules...)
	}
	if summary := repoSummary.GetSummary(); summary != "" {
		rules = append(rules, Rule{Name: "repo_overview", Content: summary, Header: getCursorRuleHeader(
			CursorHeader{Description: repoOverviewRuleDescription},
		)})
	}

	return WriteRules(rules, rulesPath, "mdc")
}

type CursorHeader struct {
	Description        string
	Globs              string
	ConditionallyApply bool
}

func getCursorRuleHeader(props CursorHeader) string {
	globs := "**/*"
	if props.Globs != "" {
		globs = props.Globs
	}
	alwaysApply := !props.ConditionallyApply
	return fmt.Sprintf(`---
description: %v
globs: %v
alwaysApply: %v
---

`, props.Description, globs, alwaysApply)
}

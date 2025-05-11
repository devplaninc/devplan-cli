package ide

import (
	"fmt"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/artifacts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"path/filepath"
)

func createCursorRulesFromPrompt(featurePrompt *documents.DocumentEntity, repoSummary *artifacts.ArtifactRepoSummary) error {
	rules := []Rule{
		{Name: "flow", Content: devFlowRule, Header: GetRuleHeader(CursorHeader{
			Description: devFlowRuleDescription,
		})},
		{Name: "rules", Content: rulesRule, Header: GetRuleHeader(CursorHeader{
			Description: generalRuleDescription, Globs: ".cursor/rules/*.mdc",
		})},
		{Name: "insights", Content: insightsRule, Header: GetRuleHeader(CursorHeader{
			Description: insightsRuleDescription,
		})},
	}
	if featurePrompt != nil {
		rules = append(rules, Rule{
			Name:    "current_feature",
			Content: featurePrompt.GetContent(),
			Header: GetRuleHeader(CursorHeader{
				Description: currentFeatRuleDescription,
			}),
		})
	}
	if summary := repoSummary.GetSummary().GetContent(); summary != "" {
		rules = append(rules, Rule{Name: "repo_overview", Content: summary, Header: GetRuleHeader(
			CursorHeader{Description: repoOverviewRuleDescription},
		)})
	}

	return WriteRules(rules, filepath.Join(".cursor", "rules"), "mdc")
}

type CursorHeader struct {
	Description        string
	Globs              string
	ConditionallyApply bool
}

func GetRuleHeader(props CursorHeader) string {
	globs := "**/*"
	alwaysApply := !props.ConditionallyApply
	return fmt.Sprintf(`---
description: %v
globs: %v
alwaysApply: %v
---

`, props.Description, globs, alwaysApply)
}

package ide

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/utils/prompts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/artifacts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"path/filepath"
)

func createCursorRulesFromPrompt(prompt *documents.DocumentEntity, repoSummary *artifacts.ArtifactRepoSummary) error {
	rules := []Rule{
		{Name: "flow",
			Content: replacePaths(devFlowRule, ".cursor/rules", "mdc"),
			Header:  getCursorRuleHeader(CursorHeader{Description: devFlowRuleDescription}),
			Footer:  allOtherRulesSuffix(".cursor/rules", "mdc"),
		},
		{Name: "rules", Content: rulesRule, Header: getCursorRuleHeader(CursorHeader{
			Description: generalRuleDescription, Globs: ".cursor/rules/*.mdc",
		})},
		{Name: "insights", Content: insightsRule, Header: getCursorRuleHeader(CursorHeader{
			Description: insightsRuleDescription,
		})},
	}

	if prompt != nil {
		targetID, err := prompts.GetTargetID(prompt)
		if err != nil {
			return fmt.Errorf("failed to get feature ID for feature prompt: %w", err)
		}
		rules = append(rules, Rule{
			Name:    "current_feature",
			Content: prompt.GetContent(),
			Header: getCursorRuleHeader(CursorHeader{
				Description: currentFeatRuleDescription,
			}),
			TargetID: targetID,
		})
	}
	if summary := repoSummary.GetSummary().GetContent(); summary != "" {
		rules = append(rules, Rule{Name: "repo_overview", Content: summary, Header: getCursorRuleHeader(
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

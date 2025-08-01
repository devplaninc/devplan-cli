package ide

import (
	"fmt"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/integrations"
)

func createWindsurfRulesFromPrompt(rulesPath string, featurePrompt *documents.DocumentEntity, repoSummary *integrations.RepositorySummary) error {
	rules := []Rule{
		{Name: "flow", Content: replacePaths(devFlowRule, rulesPath, "md"),
			Header: getWindsurfRuleHeader(WindsurfHeader{Description: devFlowRuleDescription}),
			Footer: allOtherRulesSuffix(".", "md"),
		},
	}

	if featurePrompt != nil {
		cfRules, err := generateCurrentFeatureRules(
			rulePaths{dir: rulesPath, ext: "md"},
			Rule{
				Header: getWindsurfRuleHeader(WindsurfHeader{Description: currentFeatRuleDescription}),
			}, featurePrompt)
		if err != nil {
			return fmt.Errorf("failed to generate current feature rules: %w", err)
		}
		rules = append(rules, cfRules...)
	}
	if summary := repoSummary.GetSummary(); summary != "" {
		rules = append(rules, Rule{Name: "repo_overview",
			Content: summary,
			Header: getWindsurfRuleHeader(WindsurfHeader{
				Description: repoOverviewRuleDescription,
				Trigger:     "model_decision",
			})})
	}

	return WriteRules(rules, rulesPath, "md")
}

type WindsurfHeader struct {
	Description string
	Globs       string
	Trigger     string
}

func getWindsurfRuleHeader(props WindsurfHeader) string {
	globs := "**/*"
	if props.Globs != "" {
		globs = props.Globs
	}
	trigger := "always_on"
	if props.Trigger != "" {
		trigger = props.Trigger
	}
	return fmt.Sprintf(`---
description: %v
globs: %v
trigger: %v
---

`, props.Description, globs, trigger)
}

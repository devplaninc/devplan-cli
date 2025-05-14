package ide

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/utils/prompts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/artifacts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"path/filepath"
)

func createWindsurfRulesFromPrompt(featurePrompt *documents.DocumentEntity, repoSummary *artifacts.ArtifactRepoSummary) error {
	rulesPath := filepath.Join(".windsurf", "rules")
	rules := []Rule{
		{Name: "flow", Content: mdPathsReplace(devFlowRule, rulesPath), Header: getWindsurfRuleHeader(WindsurfHeader{
			Description: devFlowRuleDescription,
		})},
		{Name: "rules", Content: rulesRule, Header: getWindsurfRuleHeader(WindsurfHeader{
			Description: generalRuleDescription, Globs: ".windsurf/rules/*.md", Trigger: "glob",
		})},
		{Name: "insights", Content: insightsRule, Header: getWindsurfRuleHeader(WindsurfHeader{
			Description: insightsRuleDescription,
		})},
	}

	if featurePrompt != nil {
		featID, err := prompts.GetPromptFeatureID(featurePrompt)
		if err != nil {
			return fmt.Errorf("failed to get feature ID for feature prompt: %w", err)
		}
		rules = append(rules, Rule{
			Name:      "current_feature",
			Content:   featurePrompt.GetContent(),
			Header:    getWindsurfRuleHeader(WindsurfHeader{Description: currentFeatRuleDescription}),
			FeatureID: featID,
		})
	}
	if summary := repoSummary.GetSummary().GetContent(); summary != "" {
		rules = append(rules, Rule{Name: "repo_overview", Content: summary, Header: getWindsurfRuleHeader(
			WindsurfHeader{Description: repoOverviewRuleDescription},
		)})
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

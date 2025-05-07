package focus

import _ "embed"

//go:embed insights.txt
var insightsRule string

//go:embed dev_flow.txt
var devFlowRule string

//go:embed rules.txt
var rulesRule string

const currentFeatRuleDescription = "Description of the feature the implementation should focus on now"
const devFlowRuleDescription = "A general guideline for the development workflow"
const insightsRuleDescription = "Framework for systematic collection of insights from chat interactions, code execution, and pattern analysis."
const generalRuleDescription = "Guidelines for creating and maintaining rules to ensure consistency and effectiveness."
const repoOverviewRuleDescription = "High level overview of this repository"

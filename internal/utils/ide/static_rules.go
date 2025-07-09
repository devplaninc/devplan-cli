package ide

import _ "embed"

//go:embed dev_flow.txt
var devFlowRule string

const currentFeatRuleDescription = "Description of the feature the implementation should focus on now"
const devFlowRuleDescription = "A general guideline for the development workflow"
const repoOverviewRuleDescription = "High level overview of this repository"

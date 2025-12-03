package devplan

import (
	"fmt"
)

const apiPath = "api/v1"
const selfPath = apiPath + "/user"

func companyPath(companyID int32) string {
	return fmt.Sprintf("%v/company/%v", apiPath, companyID)
}

func projectsPath(companyID int32) string {
	return fmt.Sprintf("%v/projects", companyPath(companyID))
}

func projectDocsPath(companyID int32, projectID string) string {
	return fmt.Sprintf("%v/%v/docs", projectsPath(companyID), projectID)
}

func documentPath(companyID int32, documentID string) string {
	return fmt.Sprintf("%v/document/%v", companyPath(companyID), documentID)
}

func templatesPath(companyID int32) string {
	return fmt.Sprintf("%v/templates", companyPath(companyID))
}

func groupsPath(companyID int32) string {
	return fmt.Sprintf("%v/groups", companyPath(companyID))
}

func integrationPath(companyID int32, provider string) string {
	return fmt.Sprintf("%v/integration/%v", companyPath(companyID), provider)
}

func devRulePath(companyID int32, ruleName string) string {
	return fmt.Sprintf("%v/dev/rule/%v", companyPath(companyID), ruleName)
}

func devIDERecipePath(companyID int32) string {
	return fmt.Sprintf("%v/dev/ide", companyPath(companyID))
}

func devTaskPath(companyID int32, taskID string) string {
	return fmt.Sprintf("%v/dev/task/%v", companyPath(companyID), taskID)
}

func devTaskExecRecipePath(companyID int32, taskID string) string {
	return fmt.Sprintf("%v/dev/task/%v/executable", companyPath(companyID), taskID)
}

func submitWorkLogPath(companyID int32) string {
	return fmt.Sprintf("%v/worklog/submit", companyPath(companyID))
}

func groupPath(companyID int32, groupID string) string {
	return fmt.Sprintf("%v/%v", groupsPath(companyID), groupID)
}

func repoSummariesPath(companyID int32) string {
	return fmt.Sprintf("%v/repo-summaries", companyPath(companyID))
}

func taskSpecsPath(companyID int32, taskID string) string {
	return fmt.Sprintf("%v/specs", devTaskPath(companyID, taskID))
}

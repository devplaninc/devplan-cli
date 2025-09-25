package devplan

import "fmt"

const apiPath = "api/v1"
const selfPath = apiPath + "/user"

func companyPath(companyID int32) string {
	return fmt.Sprintf("%v/company/%v", apiPath, companyID)
}

func projectsPath(companyID int32) string {
	return fmt.Sprintf("%v/projects", companyPath(companyID))
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

func groupPath(companyID int32, groupID string) string {
	return fmt.Sprintf("%v/%v", groupsPath(companyID), groupID)
}

func repoSummariesPath(companyID int32) string {
	return fmt.Sprintf("%v/repo-summaries", companyPath(companyID))
}

package selector

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/utils/prefs"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
)

func Project(projects []*documents.ProjectEntity, props Props) (*documents.ProjectEntity, error) {
	// If there's only one project, use it
	if len(projects) == 1 {
		// Save the selected project ID to preferences
		prefs.SetLastProjectID(projects[0].GetId())
		return projects[0], nil
	}

	// Try to get the last selected project from preferences
	lastProjectID := prefs.GetLastProjectID()

	var items []item
	var defaultIndex int
	var foundLastProject bool

	for i, p := range projects {
		items = append(items, item{
			id: p.GetId(), title: p.GetTitle(), extra: fmt.Sprintf("Status: %v", p.GetDetails().GetStatus()),
		})

		// Check if this is the last selected project
		if p.GetId() == lastProjectID {
			defaultIndex = i
			foundLastProject = true
		}
	}

	// Only use the default index if we found the last project
	if !foundLastProject {
		defaultIndex = -1
	}

	projectID, err := runSelector("project", items, props, defaultIndex)
	if err != nil {
		return nil, err
	}
	if projectID == "" {
		return nil, nil
	}

	for _, p := range projects {
		if fmt.Sprintf("%v", p.GetId()) == projectID {
			// Save the selected project ID to preferences
			prefs.SetLastProjectID(p.GetId())
			return p, nil
		}
	}
	return nil, nil
}

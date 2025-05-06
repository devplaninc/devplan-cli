package selector

import (
	"fmt"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
)

func Project(projects []*documents.ProjectEntity, props Props) (*documents.ProjectEntity, error) {
	if len(projects) == 1 {
		return projects[0], nil
	}
	var items []item
	for _, p := range projects {
		items = append(items, item{
			id: p.GetId(), title: p.GetTitle(), extra: fmt.Sprintf("Status: %v", p.GetDetails().GetStatus()),
		})
	}
	projectID, err := runSelector("project", items, props)
	if err != nil {
		return nil, err
	}
	if projectID == "" {
		return nil, nil
	}
	for _, p := range projects {
		if fmt.Sprintf("%v", p.GetId()) == projectID {
			return p, nil
		}
	}
	return nil, nil
}

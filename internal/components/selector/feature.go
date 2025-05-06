package selector

import (
	"fmt"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
)

func Feature(features []*documents.DocumentEntity, props Props) (*documents.DocumentEntity, error) {
	if len(features) == 1 {
		return features[0], nil
	}
	var items []item
	for _, f := range features {
		items = append(items, item{
			id: f.GetId(), title: f.GetTitle(),
		})
	}
	featureID, err := runSelector("feature", items, props)
	if err != nil {
		return nil, err
	}
	if featureID == "" {
		return nil, nil
	}
	for _, f := range features {
		if fmt.Sprintf("%v", f.GetId()) == featureID {
			return f, nil
		}
	}
	return nil, nil
}

package selector

import (
	"fmt"

	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
)

func Document(
	title string,
	docs []*documents.DocumentEntity,
	props Props,
	chosen string,
) (*documents.DocumentEntity, error) {
	if chosen != "" {
		for _, d := range docs {
			if d.GetId() == chosen {
				return d, nil
			}
		}
		return nil, fmt.Errorf("%v [%v] not found among available %v entries", title, chosen, title)
	}
	if len(docs) == 1 {
		return docs[0], nil
	}
	var items []item
	for _, d := range docs {
		items = append(items, item{
			id: d.GetId(), title: d.GetTitle(),
		})
	}
	docID, err := runSelector(title, items, props)
	if err != nil {
		return nil, err
	}
	if docID == "" {
		return nil, nil
	}
	for _, d := range docs {
		if fmt.Sprintf("%v", d.GetId()) == docID {
			return d, nil
		}
	}
	return nil, nil
}

func Feature(features []*documents.DocumentEntity, props Props, chosen string) (*documents.DocumentEntity, error) {
	return Document("feature", features, props, chosen)
}

func Task(tasks []*documents.DocumentEntity, props Props, chosen string) (*documents.DocumentEntity, error) {
	return Document("feature", tasks, props, chosen)
}

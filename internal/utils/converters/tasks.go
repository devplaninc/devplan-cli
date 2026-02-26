package converters

import (
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/tasks"
	"google.golang.org/protobuf/encoding/protojson"
)

func GetTaskDetails(doc *documents.DocumentEntity) (*tasks.TaskDetails, error) {
	if doc.GetDetails() == "" {
		return nil, nil
	}
	details := &tasks.TaskDetails{}
	u := protojson.UnmarshalOptions{DiscardUnknown: true}
	return details, u.Unmarshal([]byte(doc.GetDetails()), details)
}

func GetFeatureDetails(doc *documents.DocumentEntity) (*tasks.FeatureDetails, error) {
	if doc.GetDetails() == "" {
		return nil, nil
	}
	details := &tasks.FeatureDetails{}
	u := protojson.UnmarshalOptions{DiscardUnknown: true}
	return details, u.Unmarshal([]byte(doc.GetDetails()), details)
}

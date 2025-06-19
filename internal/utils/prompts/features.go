package prompts

import (
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"google.golang.org/protobuf/encoding/protojson"
)

func GetDocDetails(promptDoc *documents.DocumentEntity) (*documents.CustomDocumentDetails, error) {
	if promptDoc.GetDetails() == "" {
		return nil, nil
	}
	details := &documents.CustomDocumentDetails{}
	u := protojson.UnmarshalOptions{DiscardUnknown: true}
	return details, u.Unmarshal([]byte(promptDoc.GetDetails()), details)
}

type Target struct {
	FeatureID string
	Stepped   bool
	Combined  bool
}

func GetTarget(promptDoc *documents.DocumentEntity) (*Target, error) {
	details, err := GetDocDetails(promptDoc)
	if err != nil {
		return nil, err
	}
	featureID := details.GetExtraPromptParams()["feature_id"]
	steppedFeatureID := details.GetExtraPromptParams()["stepped_feature_id"]
	if featureID != "" {
		return &Target{FeatureID: featureID}, nil
	}
	if steppedFeatureID != "" {
		return &Target{FeatureID: steppedFeatureID, Stepped: true}, nil
	}
	if details.GetExtraPromptParams()["combined"] == "true" {
		return &Target{Combined: true}, nil
	}
	return nil, nil
}

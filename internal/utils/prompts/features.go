package prompts

import (
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"google.golang.org/protobuf/encoding/protojson"
)

const CombinedFeatureID = "combined-prompt"

func GetDocDetails(promptDoc *documents.DocumentEntity) (*documents.CustomDocumentDetails, error) {
	if promptDoc.GetDetails() == "" {
		return nil, nil
	}
	details := &documents.CustomDocumentDetails{}
	u := protojson.UnmarshalOptions{DiscardUnknown: true}
	return details, u.Unmarshal([]byte(promptDoc.GetDetails()), details)
}

func GetTargetID(promptDoc *documents.DocumentEntity) (string, error) {
	details, err := GetDocDetails(promptDoc)
	if err != nil {
		return "", err
	}
	featureID := details.GetExtraPromptParams()["feature_id"]
	if featureID != "" {
		return featureID, nil
	}
	if details.GetExtraPromptParams()["combined"] == "true" {
		return CombinedFeatureID, nil
	}
	return "", nil
}

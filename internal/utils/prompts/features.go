package prompts

import (
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"google.golang.org/protobuf/encoding/protojson"
)

func GetPromptFeatureID(promptDoc *documents.DocumentEntity) (string, error) {
	if promptDoc.GetDetails() == "" {
		return "", nil
	}
	details := &documents.CustomDocumentDetails{}
	err := protojson.Unmarshal([]byte(promptDoc.GetDetails()), details)
	if err != nil {
		return "", err
	}
	return details.GetExtraPromptParams()["feature_id"], nil
}

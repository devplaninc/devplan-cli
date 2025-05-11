package picker

import (
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"google.golang.org/protobuf/encoding/protojson"
)

func GetFeaturePrompt(featureID string, docs []*documents.DocumentEntity) (*documents.DocumentEntity, error) {
	codeAssist := getCodingAssistant(docs)
	if codeAssist == nil {
		return nil, nil
	}
	for _, d := range docs {
		if d.GetParentId() != codeAssist.GetId() {
			continue
		}
		details := &documents.CustomDocumentDetails{}
		err := protojson.Unmarshal([]byte(d.GetDetails()), details)
		if err != nil {
			return nil, err
		}
		if details.GetExtraPromptParams()["feature_id"] == featureID {
			return d, nil
		}
	}
	return nil, nil
}

func getCodingAssistant(docs []*documents.DocumentEntity) *documents.DocumentEntity {
	for _, d := range docs {
		if d.GetType() == documents.DocumentType_CODING_ASSISTANT {
			return d
		}
	}
	return nil
}

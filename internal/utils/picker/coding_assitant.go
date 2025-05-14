package picker

import (
	"github.com/devplaninc/devplan-cli/internal/utils/prompts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
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
		featID, err := prompts.GetPromptFeatureID(d)
		if err != nil {
			return nil, err
		}
		if featID == featureID {
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

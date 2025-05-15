package picker

import (
	"github.com/devplaninc/devplan-cli/internal/utils/prompts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
)

func GetTargetPrompt(target DevTarget, docs []*documents.DocumentEntity) (*documents.DocumentEntity, error) {
	codeAssist := getCodingAssistant(docs)
	if codeAssist == nil {
		return nil, nil
	}
	for _, d := range docs {
		if d.GetParentId() != codeAssist.GetId() {
			continue
		}
		targetID, err := prompts.GetTargetID(d)
		if err != nil {
			return nil, err
		}
		if (target.SingleShot && targetID == prompts.CombinedFeatureID) ||
			(!target.SingleShot && targetID == target.SpecificFeature.GetId()) {
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

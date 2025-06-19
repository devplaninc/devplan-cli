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
	featureID := target.SpecificFeature.GetId()
	for _, d := range docs {
		if d.GetParentId() != codeAssist.GetId() {
			continue
		}
		t, err := prompts.GetTarget(d)
		if err != nil {
			return nil, err
		}
		if t == nil {
			continue
		}

		if target.SingleShot && t.Combined {
			return d, nil
		}
		if t.FeatureID == "" {
			continue
		}
		if t.FeatureID == featureID && t.Stepped == target.Steps {
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

package picker

import (
	"github.com/devplaninc/devplan-cli/internal/utils/prompts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
)

func GetTargetPrompt(target DevTarget, docs []*documents.DocumentEntity) (*prompts.Info, error) {
	taskPrompt, err := pickTask(target, docs)
	if err != nil {
		return nil, err
	}
	if taskPrompt != nil {
		return taskPrompt, nil
	}
	return pickFeature(target, docs)
}

func pickTask(target DevTarget, docs []*documents.DocumentEntity) (*prompts.Info, error) {
	task := target.Task
	promptTemplateId := target.Template.GetDetails().GetTaskPrompt().GetInfo().GetTemplateId()
	if task == nil || promptTemplateId == "" {
		return nil, nil
	}
	for _, d := range docs {
		if d.GetParentId() == task.GetId() && d.GetTemplateId() == promptTemplateId {
			return &prompts.Info{
				Doc:    d,
				Target: prompts.Target{TaskID: task.GetId()},
			}, nil
		}
	}
	return nil, nil
}

func pickFeature(target DevTarget, docs []*documents.DocumentEntity) (*prompts.Info, error) {
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
			return &prompts.Info{Doc: d, Target: *t}, nil
		}
		if t.FeatureID == "" {
			continue
		}
		if t.FeatureID == featureID && t.Stepped == target.Steps {
			return &prompts.Info{Doc: d, Target: *t}, nil
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

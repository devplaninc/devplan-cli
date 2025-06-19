package ide

import (
	"bytes"
	"fmt"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"google.golang.org/protobuf/encoding/protojson"
	"text/template"
)

const stepsDescription = `Bellow are step-by-step instructions for the current feature.

## Summary:

{{.Summary}}

## Steps:
{{range .Steps}}
{{.Index}}. {{.Name}} (file: [{{.FileName}}]({{.FullPath}}))
{{- end }}
`

const stepDescription = `# {{.Name}}

{{.Instructions}}
`

type steps struct {
	Summary string
	Steps   []step
}

type step struct {
	Index        int
	Name         string
	RuleName     string
	FileName     string
	Instructions string
	FullPath     string
}

func generateSteppedRules(
	paths rulePaths, base Rule, featurePrompt *documents.DocumentEntity,
) ([]Rule, error) {
	c := &documents.SteppedFeaturePromptContent{}
	if err := protojson.Unmarshal([]byte(featurePrompt.GetContent()), c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal feature prompt content: %w", err)
	}
	st := &steps{Summary: c.GetRequirementsSummary()}
	for i, s := range c.GetSteps() {
		newStep := step{
			Index:        i + 1,
			Name:         s.GetName(),
			RuleName:     fmt.Sprintf("current_feature_step_%d", i+1),
			Instructions: s.GetInstructions(),
		}
		newStep.FileName = fmt.Sprintf("%v%v.%v", ruleFileNamePrefix, newStep.RuleName, paths.ext)
		newStep.FullPath = fmt.Sprintf("%v/%v", paths.dir, newStep.FileName)
		st.Steps = append(st.Steps, newStep)
	}
	var rules []Rule
	stepsRule, err := ruleFromTemplate("current_feature", base, stepsDescription, st)
	if err != nil {
		return nil, fmt.Errorf("failed to generate overall steps rule: %w", err)
	}
	rules = append(rules, stepsRule)

	for _, s := range st.Steps {
		stepRule, err := ruleFromTemplate(s.RuleName, base, stepDescription, s)
		if err != nil {
			return nil, fmt.Errorf("failed to generate step rule: %w", err)
		}
		rules = append(rules, stepRule)
	}
	return rules, nil
}

func ruleFromTemplate(name string, base Rule, tmplStr string, data any) (Rule, error) {
	tmpl, err := template.New("steps").Parse(tmplStr)
	if err != nil {
		return Rule{}, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return Rule{}, fmt.Errorf("failed to execute template: %w", err)
	}

	content := buf.String()
	return Rule{
		Name:     name,
		Content:  content,
		Target:   base.Target,
		NoPrefix: base.NoPrefix,
		Footer:   base.Footer,
		Header:   base.Header,
	}, nil
}

package ide

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/devplaninc/devplan-cli/internal/utils/prompts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/integrations"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

const ruleFileNamePrefix = "devplan_"

type Rule struct {
	NoPrefix bool
	Name     string
	Content  string
	Header   string
	Footer   string
	Target   *prompts.Target
}

func WriteMultiIDE(
	assistants []Assistant,
	featPrompt *documents.DocumentEntity,
	summary *integrations.RepositorySummary,
	yes bool,
) error {
	err := confirmRulesGeneration(assistants, featPrompt, summary, yes)
	if err != nil {
		return err
	}
	for _, name := range assistants {
		fmt.Println()
		if err := processAssistant(name, featPrompt, summary); err != nil {
			return err
		}
	}
	return nil
}

func processAssistant(asst Assistant, featPrompt *documents.DocumentEntity, summary *integrations.RepositorySummary) error {
	var err error
	switch asst {
	case JunieAI:
		err = createJunieRules(featPrompt, summary)
	case CursorAI:
		err = createCursorRulesFromPrompt(featPrompt, summary)
	case WindsurfAI:
		err = createWindsurfRulesFromPrompt(featPrompt, summary)
	default:
		err = fmt.Errorf("unknown assistant: %v", asst)
	}
	if err != nil {
		return err
	}
	err = git.UpdateIgnore()
	if err != nil {
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}
	fmt.Println(out.Successf("%s rules created successfully!", asst))
	return nil
}

func confirmRulesGeneration(
	assistants []Assistant,
	featurePrompt *documents.DocumentEntity,
	repoSummary *integrations.RepositorySummary,
	yes bool,
) error {
	if yes {
		return nil
	}
	if featurePrompt == nil && repoSummary == nil {
		return fmt.Errorf("neither feature prompt nor repo summary found for the feature and repository")
	}
	root, err := git.GetRoot()
	if err != nil {
		return err
	}
	var assistantsStr []string
	for _, assistant := range assistants {
		assistantsStr = append(assistantsStr, string(assistant))
	}
	fmt.Print(out.H(fmt.Sprintf(
		"\nRules for %v will be generated for the selected feature in the current repository %v.\n\n",
		strings.Join(assistantsStr, ", "), root)))
	prompt := promptui.Prompt{
		Label:     "Create rules",
		IsConfirm: true,
	}
	result, err := prompt.Run()
	if err != nil {
		return err
	}
	if result != "y" {
		return errors.New("aborted:" + result)
	}
	return nil
}

func WriteRules(rules []Rule, path string, extension string) error {
	root, err := git.GetRoot()
	if err != nil {
		return err
	}
	rulesDir := filepath.Join(root, path)

	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		err = os.MkdirAll(rulesDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create rules directory: %w", err)
		}
	}

	for _, rule := range rules {
		fileName := fmt.Sprintf("%s.%s", rule.Name, extension)
		if !rule.NoPrefix {
			fileName = fmt.Sprintf("%v%s", ruleFileNamePrefix, fileName)
		}
		filePath := filepath.Join(rulesDir, fileName)

		// Write the rule content to the file
		content := ""
		if h := rule.Header; h != "" {
			content = fmt.Sprintf("%v\n\n%v", h, content)
		}

		if t := rule.Target; t != nil && t.FeatureID != "" {
			content = fmt.Sprintf("%v<!-- feature_id: %v -->\n\n", content, t.FeatureID)
		}

		content = fmt.Sprintf("%v%v", content, rule.Content)

		if f := rule.Footer; f != "" {
			content = fmt.Sprintf("%v\n%v", content, f)
		}

		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			return fmt.Errorf("failed to write rule file %s: %w", rule.Name, err)
		}
		relativeFile := filepath.Join(path, fileName)
		fmt.Printf("%v %v\n", out.Check, relativeFile)
	}
	return nil
}

type rulePaths struct {
	dir string
	ext string
}

func generateCurrentFeatureRules(
	paths rulePaths,
	base Rule,
	prompt *documents.DocumentEntity,
) ([]Rule, error) {
	target, err := prompts.GetTarget(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get target info for feature prompt: %w", err)
	}
	if target == nil || !target.Stepped {
		rule := Rule{
			Name:     "current_feature",
			Content:  prompt.GetContent(),
			Header:   base.Header,
			Footer:   base.Footer,
			NoPrefix: base.NoPrefix,
			Target:   target,
		}
		return []Rule{rule}, nil
	}
	return generateSteppedRules(paths, base, prompt)
}

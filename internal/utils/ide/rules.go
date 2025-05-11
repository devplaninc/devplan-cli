package ide

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/artifacts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"github.com/manifoldco/promptui"
	"os"
	"path/filepath"
	"strings"
)

type Rule struct {
	NoPrefix bool
	Name     string
	Content  string
	Header   string
	Footer   string
}

func WriteMultiIDE(
	ides []string,
	featPrompt *documents.DocumentEntity,
	summary *artifacts.ArtifactRepoSummary,
	yes bool,
) error {
	err := confirmRulesGeneration(ides, featPrompt, summary, yes)
	if err != nil {
		return err
	}
	for _, name := range ides {
		fmt.Println()
		if err := processIDE(name, featPrompt, summary); err != nil {
			return err
		}
	}
	return nil
}

func processIDE(ideName string, featPrompt *documents.DocumentEntity, summary *artifacts.ArtifactRepoSummary) error {
	var err error
	switch ideName {
	case Junie:
		err = createJunieRules(featPrompt, summary)
	case Cursor:
		err = createCursorRulesFromPrompt(featPrompt, summary)
	default:
		err = fmt.Errorf("unknown ide: %v", ideName)
	}
	if err != nil {
		return err
	}
	err = git.UpdateIgnore()
	if err != nil {
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}
	fmt.Println(out.Successf("%s rules created successfully!", ideName))
	return nil
}

func confirmRulesGeneration(
	ideNames []string,
	featurePrompt *documents.DocumentEntity,
	repoSummary *artifacts.ArtifactRepoSummary,
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
	fmt.Printf(out.H(fmt.Sprintf(
		"\nRules for %v will be generated for the selected feature in the current repository %v.\n\n",
		strings.Join(ideNames, ", "), root)))
	prompt := promptui.Prompt{
		Label:     "Create rules",
		IsConfirm: true,
	}
	result, err := prompt.Run()
	if err != nil {
		return err
	}
	if result != "y" {
		return fmt.Errorf("aborted:" + result)
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
			fileName = fmt.Sprintf("devplan_%s", fileName)
		}
		filePath := filepath.Join(rulesDir, fileName)

		// Write the rule content to the file
		content := rule.Content
		if h := rule.Header; h != "" {
			content = fmt.Sprintf("%v\n\n%v", h, content)
		}
		if f := rule.Footer; f != "" {
			content = fmt.Sprintf("%v\n\n%v", content, f)
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

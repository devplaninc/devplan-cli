package ide

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"os"
	"path/filepath"
)

type Rule struct {
	NoPrefix bool
	Name     string
	Content  string
	Header   string
	Footer   string
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

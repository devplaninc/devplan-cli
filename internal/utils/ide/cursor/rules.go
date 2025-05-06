package cursor

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"os"
	"path/filepath"

	"github.com/devplaninc/devplan-cli/internal/utils/ide"
)

func CreateRules(rules []ide.Rule) error {
	root, err := git.GetRoot()
	if err != nil {
		return err
	}
	relativePath := filepath.Join(".cursor", "rules")
	rulesDir := filepath.Join(root, relativePath)

	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		err = os.MkdirAll(rulesDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create cursor rules directory: %w", err)
		}
	}

	for _, rule := range rules {
		fileName := "devplan_" + rule.Name + ".mdc"
		filePath := filepath.Join(rulesDir, fileName)

		// Write the rule content to the file
		content := rule.Content
		if h := rule.Header; h != "" {
			content = fmt.Sprintf("%v\n\n%v", h, content)
		}
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			return fmt.Errorf("failed to write rule file %s: %w", rule.Name, err)
		}
		relativeFile := filepath.Join(relativePath, fileName)
		fmt.Printf("%v %v\n", out.Check, relativeFile)
	}
	return nil
}

type Header struct {
	Description        string
	Globs              string
	ConditionallyApply bool
}

func GetRuleHeader(props Header) string {
	globs := "**/*"
	alwaysApply := !props.ConditionallyApply
	return fmt.Sprintf(`---
description: %v
globs: %v
alwaysApply: %v
---

`, props.Description, globs, alwaysApply)
}

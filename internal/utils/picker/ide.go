package picker

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/manifoldco/promptui"
	"slices"
)

func GetAllowedIDEs() []string {
	return []string{ide.Cursor, ide.Junie}
}

func IDEs(ideName string) ([]string, error) {
	ides, err := getIDEsToProvision(ideName)
	if err != nil {
		return nil, err
	}
	allowedIDEs := GetAllowedIDEs()
	if len(ides) == 0 {
		fmt.Printf("No IDEs auto detected or provided.\n")
		prompt := promptui.Select{Label: "Select IDE", Items: allowedIDEs}
		_, ideName, err = prompt.Run()
		if err != nil {
			return nil, err
		}
		if ideName == "" || !slices.Contains(allowedIDEs, ideName) {
			return nil, fmt.Errorf("no valid ide selected")
		}
		ides = []string{ideName}
	}
	return ides, nil
}

func getIDEsToProvision(ideName string) ([]string, error) {
	if ideName != "" {
		return []string{ideName}, nil
	}
	return ide.Detect()
}

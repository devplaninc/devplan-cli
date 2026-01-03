package picker

import (
	"fmt"
	"slices"

	"github.com/charmbracelet/huh"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/devplan-cli/internal/utils/prefs"
)

func AssistantForIDE(ideName string) ([]ide.Assistant, error) {
	assistants, err := getAssistantsToProvision(ideName)
	if err != nil {
		return nil, err
	}
	if len(assistants) > 0 {
		return assistants, nil
	}
	if lastAsst, ok := prefs.GetLastAssistant(); ok {
		return []ide.Assistant{ide.Assistant(lastAsst)}, nil
	}
	allowedAssistants := ide.GetAssistants()
	fmt.Printf("No AssistantForIDE auto detected or provided.\n")
	var asst ide.Assistant
	err = huh.NewSelect[ide.Assistant]().
		Title("Select AssistantForIDE").
		Options(huh.NewOptions(allowedAssistants...)...).
		Value(&asst).
		Run()
	if err != nil {
		return nil, err
	}
	if asst == "" || !slices.Contains(allowedAssistants, asst) {
		return nil, fmt.Errorf("no valid assistant selected")
	}
	prefs.SetLastAssistant(string(asst))
	return []ide.Assistant{asst}, nil
}

func Assistant(asst string) ([]ide.Assistant, error) {
	if asst != "" {
		if slices.Contains(ide.GetAssistants(), ide.Assistant(asst)) {
			return []ide.Assistant{ide.Assistant(asst)}, nil
		}
		return nil, fmt.Errorf("invalid assistant selected: [%v]", asst)
	}
	assistants, err := getAssistantsToProvision("")
	if err != nil {
		return nil, err
	}
	if len(assistants) > 0 {
		return assistants, nil
	}

	allowedAssistants := ide.GetAssistants()
	var selected ide.Assistant
	err = huh.NewSelect[ide.Assistant]().
		Title("Select Assistant").
		Options(huh.NewOptions(allowedAssistants...)...).
		Value(&selected).
		Run()
	if err != nil {
		return nil, err
	}
	if selected == "" || !slices.Contains(allowedAssistants, selected) {
		return nil, fmt.Errorf("no valid assistant selected")
	}
	return []ide.Assistant{selected}, nil
}

func getAssistantsToProvision(ideName string) ([]ide.Assistant, error) {
	if ideName != "" {
		res, err := ide.GetAssistant(ide.IDE(ideName))
		if err != nil {
			return nil, err
		}
		return []ide.Assistant{res}, nil
	}

	// First, try to detect AssistantForIDE based on repository files
	result, err := ide.DetectAssistant()
	if err != nil {
		return nil, err
	}

	// Then, detect installed AssistantForIDE on the system
	installedIDEs, err := ide.DetectInstalledIDEs()
	if err != nil {
		// If there's an error detecting installed AssistantForIDE, just use the repository-based detection
		return result, nil
	}

	allowed := ide.GetKnown()
	// Add installed AssistantForIDE that aren't already in the result
	for idN := range installedIDEs {
		id := idN
		if !slices.Contains(allowed, id) {
			continue
		}
		if asst, err := ide.GetAssistant(id); err == nil {
			if !slices.Contains(result, asst) {
				result = append(result, asst)
			}
		}
	}

	return result, nil
}

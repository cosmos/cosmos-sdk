package prompt

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

func Select(label string, options []string) (string, error) {
	selectUi := promptui.Select{
		Label: label,
		Items: options,
	}

	_, selectedProposalType, err := selectUi.Run()
	if err != nil {
		return "", fmt.Errorf("failed to prompt proposal types: %w", err)
	}

	return selectedProposalType, nil
}

func PromptString(label string, validate func(string) error) (string, error) {
	promptUi := promptui.Prompt{
		Label:    label,
		Validate: validate,
	}

	return promptUi.Run()
}

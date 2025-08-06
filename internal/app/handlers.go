package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rufex/sftui/internal/models"
)

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return a.handleWindowSize(msg)
	case tea.KeyMsg:
		return a.handleKeyMsg(msg)
	}
	return a, nil
}

func (a *App) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	a.Model.Width = msg.Width
	a.Model.Height = msg.Height
	return a, nil
}

func (a *App) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.Model.ShowHelp {
		a.Model.ShowHelp = false
		return a, nil
	}

	if a.Model.ShowActionPopup {
		return a.handleActionPopup(msg)
	}

	if a.Model.ShowFirmPopup {
		return a.handleFirmPopup(msg)
	}

	if a.Model.ShowHostPopup {
		return a.handleHostPopup(msg)
	}

	if a.Model.ShowReconciliationTypePopup {
		return a.handleReconciliationTypePopup(msg)
	}

	if a.Model.ShowTextPartPopup {
		return a.handleTextPartPopup(msg)
	}

	if a.Model.ShowInPlaceEdit {
		return a.handleInPlaceEdit(msg)
	}

	if msg.Alt {
		return a, nil
	}

	if a.Model.SearchMode {
		return a.handleSearchMode(msg)
	}

	return a.handleMainKeys(msg)
}

func (a *App) handleActionPopup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.Model.ShowActionPopup = false
		a.Model.SelectedAction = 0
		a.Model.Output = "Action cancelled"
		return a, nil
	case "up", "k":
		a.Model.SelectedAction = (a.Model.SelectedAction - 1 + 4) % 4
		return a, nil
	case "down", "j":
		a.Model.SelectedAction = (a.Model.SelectedAction + 1) % 4
		return a, nil
	case "enter":
		actions := []string{"create", "import", "update", "cancel"}
		selectedActionName := actions[a.Model.SelectedAction]

		if selectedActionName == "cancel" {
			a.Model.ShowActionPopup = false
			a.Model.SelectedAction = 0
			a.Model.Output = "Action cancelled"
		} else {
			selectedCount := len(a.Model.SelectedTemplates)
			a.Model.Output = fmt.Sprintf("Action '%s' selected for %d templates (not implemented yet)", selectedActionName, selectedCount)
			a.Model.ShowActionPopup = false
			a.Model.SelectedAction = 0
		}
		return a, nil
	}
	return a, nil
}

func (a *App) handleFirmPopup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.Model.ShowFirmPopup = false
		a.Model.SelectedFirm = 0
		a.Model.Output = "Firm selection cancelled"
		return a, nil
	case "up", "k":
		if len(a.Model.FirmOptions) > 0 {
			a.Model.SelectedFirm = (a.Model.SelectedFirm - 1 + len(a.Model.FirmOptions)) % len(a.Model.FirmOptions)
		}
		return a, nil
	case "down", "j":
		if len(a.Model.FirmOptions) > 0 {
			a.Model.SelectedFirm = (a.Model.SelectedFirm + 1) % len(a.Model.FirmOptions)
		}
		return a, nil
	case "enter":
		if len(a.Model.FirmOptions) > 0 && a.Model.SelectedFirm < len(a.Model.FirmOptions) {
			selectedOption := a.Model.FirmOptions[a.Model.SelectedFirm]

			err := a.configManager.SetDefaultFirm(selectedOption.ID)
			if err != nil {
				a.Model.Output = fmt.Sprintf("Error setting default firm: %v", err)
			} else {
				a.Model.Firm = fmt.Sprintf("%s (%s)", selectedOption.Name, selectedOption.ID)
				a.Model.Output = fmt.Sprintf("Default firm set to %s", selectedOption.Name)
			}

			a.Model.ShowFirmPopup = false
			a.Model.SelectedFirm = 0
		}
		return a, nil
	}
	return a, nil
}

func (a *App) handleHostPopup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.Model.ShowHostPopup = false
		a.Model.HostTextInput.Blur()
		a.Model.Output = "Host edit cancelled"
		return a, nil
	case "enter":
		newHost := a.Model.HostTextInput.Value()
		err := a.configManager.SetHost(newHost)
		if err != nil {
			a.Model.Output = fmt.Sprintf("Error setting host: %v", err)
		} else {
			a.Model.Host = newHost
			a.Model.Output = "Host updated successfully"
		}

		a.Model.ShowHostPopup = false
		a.Model.HostTextInput.Blur()
		return a, nil
	default:
		var cmd tea.Cmd
		a.Model.HostTextInput, cmd = a.Model.HostTextInput.Update(msg)
		return a, cmd
	}
}

func (a *App) handleReconciliationTypePopup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.Model.ShowReconciliationTypePopup = false
		a.Model.SelectedReconciliationType = 0
		a.Model.Output = "Reconciliation type edit cancelled"
		return a, nil
	case "up", "k":
		a.Model.SelectedReconciliationType = (a.Model.SelectedReconciliationType - 1 + 3) % 3
		return a, nil
	case "down", "j":
		a.Model.SelectedReconciliationType = (a.Model.SelectedReconciliationType + 1) % 3
		return a, nil
	case "enter":
		reconciliationTypes := []string{"can_be_reconciled_without_data", "reconciliation_not_necessary", "only_reconciled_with_data"}
		selectedType := reconciliationTypes[a.Model.SelectedReconciliationType]

		if len(a.Model.FilteredTemplates) > 0 && a.Model.SelectedTemplate < len(a.Model.FilteredTemplates) {
			actualIndex := a.Model.FilteredTemplates[a.Model.SelectedTemplate]
			template := a.Model.Templates[actualIndex]

			err := a.configManager.UpdateReconciliationType(template.Path, selectedType)
			if err != nil {
				a.Model.Output = fmt.Sprintf("Error updating reconciliation type: %v", err)
			} else {
				a.Model.Templates[actualIndex].Config["reconciliation_type"] = selectedType
				a.Model.Output = fmt.Sprintf("Reconciliation type set to: %s", selectedType)
			}
		}

		a.Model.ShowReconciliationTypePopup = false
		a.Model.SelectedReconciliationType = 0
		return a, nil
	}
	return a, nil
}

func (a *App) handleTextPartPopup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.Model.ShowTextPartPopup = false
		a.Model.TextPartEditMode = ""
		a.Model.Output = "Text part edit cancelled"
		return a, nil
	case "tab":
		if a.Model.TextPartEditMode == "name" {
			a.Model.TextPartEditMode = "path"
			a.Model.TextPartNameInput.Blur()
			a.Model.TextPartPathInput.Focus()
		} else {
			a.Model.TextPartEditMode = "name"
			a.Model.TextPartPathInput.Blur()
			a.Model.TextPartNameInput.Focus()
		}
		return a, nil
	case "enter":
		newName := a.Model.TextPartNameInput.Value()
		newPath := a.Model.TextPartPathInput.Value()

		a.Model.ShowTextPartPopup = false
		a.Model.TextPartEditMode = ""
		a.Model.Output = fmt.Sprintf("Text part updated: %s -> %s", newName, newPath)
		return a, nil
	default:
		var cmd tea.Cmd
		if a.Model.TextPartEditMode == "name" {
			a.Model.TextPartNameInput, cmd = a.Model.TextPartNameInput.Update(msg)
		} else {
			a.Model.TextPartPathInput, cmd = a.Model.TextPartPathInput.Update(msg)
		}
		return a, cmd
	}
}

func (a *App) handleInPlaceEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if len(a.Model.FilteredTemplates) > 0 && a.Model.SelectedTemplate < len(a.Model.FilteredTemplates) {
			actualIndex := a.Model.FilteredTemplates[a.Model.SelectedTemplate]
			a.Model.Templates[actualIndex].Config[a.Model.InPlaceEditField] = a.Model.InPlaceEditOriginalValue
		}
		a.Model.ShowInPlaceEdit = false
		a.Model.InPlaceEditField = ""
		a.Model.InPlaceEditOptions = nil
		a.Model.InPlaceEditOriginalValue = nil
		a.Model.InPlaceEditSelectedIndex = 0
		a.Model.Output = "Edit cancelled"
		return a, nil
	case "up", "k", "left", "h":
		if len(a.Model.InPlaceEditOptions) > 0 {
			a.Model.InPlaceEditSelectedIndex = (a.Model.InPlaceEditSelectedIndex - 1 + len(a.Model.InPlaceEditOptions)) % len(a.Model.InPlaceEditOptions)
			selectedValue := a.Model.InPlaceEditOptions[a.Model.InPlaceEditSelectedIndex]
			a.Model.Output = fmt.Sprintf("%s: %s (↑/↓ to change, Enter to save, Esc to cancel)", a.Model.InPlaceEditField, selectedValue)
		}
		return a, nil
	case "down", "j", "right", "l":
		if len(a.Model.InPlaceEditOptions) > 0 {
			a.Model.InPlaceEditSelectedIndex = (a.Model.InPlaceEditSelectedIndex + 1) % len(a.Model.InPlaceEditOptions)
			selectedValue := a.Model.InPlaceEditOptions[a.Model.InPlaceEditSelectedIndex]
			a.Model.Output = fmt.Sprintf("%s: %s (↑/↓ to change, Enter to save, Esc to cancel)", a.Model.InPlaceEditField, selectedValue)
		}
		return a, nil
	case "enter":
		if len(a.Model.FilteredTemplates) > 0 && a.Model.SelectedTemplate < len(a.Model.FilteredTemplates) && len(a.Model.InPlaceEditOptions) > 0 {
			actualIndex := a.Model.FilteredTemplates[a.Model.SelectedTemplate]
			template := a.Model.Templates[actualIndex]
			newValue := a.Model.InPlaceEditOptions[a.Model.InPlaceEditSelectedIndex]

			err := a.updateConfigField(template.Path, a.Model.InPlaceEditField, newValue)
			if err != nil {
				a.Model.Output = fmt.Sprintf("Error updating %s: %v", a.Model.InPlaceEditField, err)
			} else {
				if a.Model.InPlaceEditField == "reconciliation_type" || a.Model.InPlaceEditField == "encoding" {
					a.Model.Templates[actualIndex].Config[a.Model.InPlaceEditField] = newValue
				} else {
					a.Model.Templates[actualIndex].Config[a.Model.InPlaceEditField] = newValue == "true"
				}
				a.Model.Output = fmt.Sprintf("%s updated to: %s", a.Model.InPlaceEditField, newValue)
			}
		}

		a.Model.ShowInPlaceEdit = false
		a.Model.InPlaceEditField = ""
		a.Model.InPlaceEditOptions = nil
		a.Model.InPlaceEditOriginalValue = nil
		a.Model.InPlaceEditSelectedIndex = 0
		return a, nil
	}
	return a, nil
}

func (a *App) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.Model.SearchMode = false
		a.Model.SearchQuery = ""
		a.Model.FilteredTemplates = a.templateManager.FilterTemplates(a.Model.Templates, a.Model.SearchQuery)
		a.Model.SelectedTemplate = 0
		a.Model.TemplatesOffset = 0
		a.Model.Output = "Search cancelled"
		return a, nil
	case "enter":
		a.Model.SearchMode = false
		if len(a.Model.FilteredTemplates) > 0 {
			a.Model.Output = fmt.Sprintf("Found %d templates", len(a.Model.FilteredTemplates))
		} else {
			a.Model.Output = "No templates found"
		}
		return a, nil
	case "backspace":
		if len(a.Model.SearchQuery) > 0 {
			a.Model.SearchQuery = a.Model.SearchQuery[:len(a.Model.SearchQuery)-1]
			a.Model.FilteredTemplates = a.templateManager.FilterTemplates(a.Model.Templates, a.Model.SearchQuery)
			a.Model.SelectedTemplate = 0
			a.Model.TemplatesOffset = 0
		}
		return a, nil
	default:
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
			char := msg.Runes[0]
			if char >= 32 && char < 127 {
				a.Model.SearchQuery += string(char)
				a.Model.FilteredTemplates = a.templateManager.FilterTemplates(a.Model.Templates, a.Model.SearchQuery)
				a.Model.SelectedTemplate = 0
				a.Model.TemplatesOffset = 0
			}
		}
		return a, nil
	}
}

func (a *App) handleMainKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "?":
		a.Model.ShowHelp = true
		return a, nil
	case "tab":
		a.navHandler.NextSection(a.Model)
		return a, nil
	case "shift+tab":
		a.navHandler.PrevSection(a.Model)
		return a, nil
	case "shift+up", "K":
		a.navHandler.HandleVerticalUp(a.Model)
		return a, nil
	case "shift+down", "J":
		a.navHandler.HandleVerticalDown(a.Model)
		return a, nil
	case "shift+left", "H":
		a.navHandler.PrevSection(a.Model)
		return a, nil
	case "shift+right", "L":
		a.navHandler.NextSection(a.Model)
		return a, nil
	case "up", "k":
		if a.Model.CurrentSection == models.TemplatesSection {
			a.navHandler.HandleTemplateNavigation(a.Model, "up")
		} else if a.Model.CurrentSection == models.DetailsSection {
			a.navHandler.HandleDetailsNavigation(a.Model, "up")
		}
		return a, nil
	case "down", "j":
		if a.Model.CurrentSection == models.TemplatesSection {
			a.navHandler.HandleTemplateNavigation(a.Model, "down")
		} else if a.Model.CurrentSection == models.DetailsSection {
			a.navHandler.HandleDetailsNavigation(a.Model, "down")
		}
		return a, nil
	case "left", "h", "right", "l":
		return a, nil
	case " ":
		return a.handleSpaceKey()
	case "/":
		return a.handleSearchKey()
	case "backspace":
		return a.handleBackspaceKey()
	case "enter":
		return a.handleEnterKey()
	}
	return a, nil
}

func (a *App) handleSpaceKey() (tea.Model, tea.Cmd) {
	if a.Model.CurrentSection == models.TemplatesSection && len(a.Model.FilteredTemplates) > 0 && a.Model.SelectedTemplate < len(a.Model.FilteredTemplates) {
		actualIndex := a.Model.FilteredTemplates[a.Model.SelectedTemplate]

		if a.Model.SelectedTemplates[actualIndex] {
			delete(a.Model.SelectedTemplates, actualIndex)
		} else {
			a.Model.SelectedTemplates[actualIndex] = true
		}

		selectedCount := len(a.Model.SelectedTemplates)
		switch selectedCount {
		case 0:
			a.Model.Output = "No templates selected"
		case 1:
			a.Model.Output = "1 template selected"
		default:
			a.Model.Output = fmt.Sprintf("%d templates selected", selectedCount)
		}
	}
	return a, nil
}

func (a *App) handleSearchKey() (tea.Model, tea.Cmd) {
	if a.Model.CurrentSection == models.TemplatesSection {
		a.Model.SearchMode = true
		a.Model.SearchQuery = ""
		a.Model.Output = "Search mode - type to filter templates"
	}
	return a, nil
}

func (a *App) handleBackspaceKey() (tea.Model, tea.Cmd) {
	if a.Model.CurrentSection == models.TemplatesSection && len(a.Model.SelectedTemplates) > 0 {
		a.Model.SelectedTemplates = make(map[int]bool)
		a.Model.Output = "All templates deselected"
	}
	return a, nil
}

func (a *App) handleEnterKey() (tea.Model, tea.Cmd) {
	switch a.Model.CurrentSection {
	case models.FirmSection:
		a.Model.ShowFirmPopup = true
		a.Model.SelectedFirm = 0
		a.Model.Output = "Select a firm or partner"
	case models.HostSection:
		a.Model.ShowHostPopup = true
		a.Model.HostTextInput.SetValue(a.Model.Host)
		a.Model.HostTextInput.Focus()
		a.Model.HostTextInput.CursorEnd()
		a.Model.Output = "Edit host URL"
	case models.DetailsSection:
		return a.handleDetailsEnter()
	case models.TemplatesSection:
		if len(a.Model.SelectedTemplates) > 0 {
			a.Model.ShowActionPopup = true
			a.Model.SelectedAction = 0
			selectedCount := len(a.Model.SelectedTemplates)
			if selectedCount == 1 {
				a.Model.Output = "1 template selected - choose action"
			} else {
				a.Model.Output = fmt.Sprintf("%d templates selected - choose action", selectedCount)
			}
		}
	}
	return a, nil
}

func (a *App) handleDetailsEnter() (tea.Model, tea.Cmd) {
	if len(a.Model.FilteredTemplates) > 0 && a.Model.SelectedTemplate < len(a.Model.FilteredTemplates) {
		actualIndex := a.Model.FilteredTemplates[a.Model.SelectedTemplate]
		template := a.Model.Templates[actualIndex]

		configFieldCount := a.GetConfigFieldCount(template)

		if a.Model.SelectedDetailField < configFieldCount {
			selectedConfigField := a.GetSelectedConfigField(template, a.Model.SelectedDetailField)

			if a.IsFieldEditable(selectedConfigField) {
				currentValue := template.Config[selectedConfigField]
				options := a.GetFieldEditOptions(selectedConfigField, currentValue)
				currentIndex := 0
				currentValueStr := fmt.Sprintf("%v", currentValue)
				for i, option := range options {
					if option == currentValueStr {
						currentIndex = i
						break
					}
				}

				a.Model.ShowInPlaceEdit = true
				a.Model.InPlaceEditField = selectedConfigField
				a.Model.InPlaceEditOptions = options
				a.Model.InPlaceEditOriginalValue = currentValue
				a.Model.InPlaceEditSelectedIndex = currentIndex
				a.Model.Output = fmt.Sprintf("Select %s value (↑/↓ to change, Enter to save, Esc to cancel)", selectedConfigField)
			}
		} else if template.Category != "shared_parts" {
			return a.handleTextPartEdit(template, configFieldCount)
		}
	}
	return a, nil
}

func (a *App) handleTextPartEdit(template models.Template, configFieldCount int) (tea.Model, tea.Cmd) {
	if textPartsInterface, exists := template.Config["text_parts"]; exists {
		if textParts, ok := textPartsInterface.(map[string]interface{}); ok {
			textPartIndex := a.Model.SelectedDetailField - configFieldCount

			type textPart struct {
				name string
				path string
			}
			var partsList []textPart
			for name, pathInterface := range textParts {
				if pathStr, ok := pathInterface.(string); ok {
					partsList = append(partsList, textPart{name: name, path: pathStr})
				}
			}

			for i := 0; i < len(partsList); i++ {
				for j := i + 1; j < len(partsList); j++ {
					if partsList[i].name > partsList[j].name {
						partsList[i], partsList[j] = partsList[j], partsList[i]
					}
				}
			}

			if textPartIndex < len(partsList) {
				selectedPart := partsList[textPartIndex]
				a.Model.ShowTextPartPopup = true
				a.Model.SelectedTextPart = textPartIndex

				a.Model.TextPartNameInput.SetValue(selectedPart.name)
				a.Model.TextPartPathInput.SetValue(selectedPart.path)
				a.Model.TextPartNameInput.Focus()
				a.Model.TextPartEditMode = "name"

				a.Model.Output = "Edit text part name and path"
			}
		}
	}
	return a, nil
}

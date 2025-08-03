package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/rufex/sftui/internal/models"
	"github.com/rufex/sftui/internal/navigation"
	"github.com/rufex/sftui/internal/template"
	"github.com/rufex/sftui/internal/ui"
)

type app struct {
	model           *models.Model
	templateManager *template.Manager
	configManager   *template.ConfigManager
	navHandler      *navigation.Handler
	uiRenderer      *ui.Renderer
}

func newApp() *app {
	return &app{
		model:           &models.Model{},
		templateManager: template.NewManager(),
		configManager:   template.NewConfigManager(),
		navHandler:      navigation.NewHandler(),
		uiRenderer:      ui.NewRenderer(),
	}
}

func (a *app) initialModel() *models.Model {
	// Initialize host text input
	hostTextInput := textinput.New()
	hostTextInput.Placeholder = "Enter host URL (e.g., https://api.example.com)"
	hostTextInput.CharLimit = 256
	hostTextInput.Width = 50

	// Initialize text part inputs
	textPartNameInput := textinput.New()
	textPartNameInput.Placeholder = "Enter text part name"
	textPartNameInput.CharLimit = 100
	textPartNameInput.Width = 40

	textPartPathInput := textinput.New()
	textPartPathInput.Placeholder = "Enter text part path"
	textPartPathInput.CharLimit = 256
	textPartPathInput.Width = 40

	a.model = &models.Model{
		CurrentSection:    models.TemplatesSection,
		SelectedTemplate:  0,
		Templates:         []models.Template{},
		SelectedTemplates: make(map[int]bool),
		Firm:              "No firm set",
		Host:              "No host set",
		HostTextInput:     hostTextInput,
		TextPartNameInput: textPartNameInput,
		TextPartPathInput: textPartPathInput,
		ShowHelp:          false,
		Output:            "Ready",
		SharedPartsUsage:  make(map[string][]string),
	}

	// Load Silverfin config
	firm, host, output := a.configManager.LoadSilverfinConfig()
	a.model.Firm = firm
	a.model.Host = host
	a.model.Output = output

	// Load firm options
	firmOptions, err := a.configManager.LoadFirmOptions()
	if err != nil {
		a.model.Output = "Error loading firm options"
	} else {
		a.model.FirmOptions = firmOptions
	}

	// Load templates
	a.model.Templates = a.templateManager.LoadTemplates()

	// Build shared parts usage mapping
	a.buildSharedPartsMapping()

	// Initialize filtered templates to show all templates
	a.model.FilteredTemplates = a.templateManager.FilterTemplates(a.model.Templates, "")

	return a.model
}

func (a *app) Init() tea.Cmd {
	return nil
}

func (a *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.model.Width = msg.Width
		a.model.Height = msg.Height
		return a, nil

	case tea.KeyMsg:
		if a.model.ShowHelp {
			a.model.ShowHelp = false
			return a, nil
		}

		// Handle action popup
		if a.model.ShowActionPopup {
			switch msg.String() {
			case "esc":
				a.model.ShowActionPopup = false
				a.model.SelectedAction = 0
				a.model.Output = "Action cancelled"
				return a, nil
			case "up", "k":
				a.model.SelectedAction = (a.model.SelectedAction - 1 + 4) % 4
				return a, nil
			case "down", "j":
				a.model.SelectedAction = (a.model.SelectedAction + 1) % 4
				return a, nil
			case "enter":
				actions := []string{"create", "import", "update", "cancel"}
				selectedActionName := actions[a.model.SelectedAction]

				if selectedActionName == "cancel" {
					a.model.ShowActionPopup = false
					a.model.SelectedAction = 0
					a.model.Output = "Action cancelled"
				} else {
					// Placeholder implementation - just show which action was selected
					selectedCount := len(a.model.SelectedTemplates)
					a.model.Output = fmt.Sprintf("Action '%s' selected for %d templates (not implemented yet)", selectedActionName, selectedCount)
					a.model.ShowActionPopup = false
					a.model.SelectedAction = 0
				}
				return a, nil
			}
			return a, nil
		}

		// Handle firm popup
		if a.model.ShowFirmPopup {
			switch msg.String() {
			case "esc":
				a.model.ShowFirmPopup = false
				a.model.SelectedFirm = 0
				a.model.Output = "Firm selection cancelled"
				return a, nil
			case "up", "k":
				if len(a.model.FirmOptions) > 0 {
					a.model.SelectedFirm = (a.model.SelectedFirm - 1 + len(a.model.FirmOptions)) % len(a.model.FirmOptions)
				}
				return a, nil
			case "down", "j":
				if len(a.model.FirmOptions) > 0 {
					a.model.SelectedFirm = (a.model.SelectedFirm + 1) % len(a.model.FirmOptions)
				}
				return a, nil
			case "enter":
				if len(a.model.FirmOptions) > 0 && a.model.SelectedFirm < len(a.model.FirmOptions) {
					selectedOption := a.model.FirmOptions[a.model.SelectedFirm]

					// Update config file
					err := a.configManager.SetDefaultFirm(selectedOption.ID)
					if err != nil {
						a.model.Output = fmt.Sprintf("Error setting default firm: %v", err)
					} else {
						// Update display and close popup
						a.model.Firm = fmt.Sprintf("%s (%s)", selectedOption.Name, selectedOption.ID)
						a.model.Output = fmt.Sprintf("Default firm set to %s", selectedOption.Name)
					}

					a.model.ShowFirmPopup = false
					a.model.SelectedFirm = 0
				}
				return a, nil
			}
			return a, nil
		}

		// Handle host popup
		if a.model.ShowHostPopup {
			switch msg.String() {
			case "esc":
				a.model.ShowHostPopup = false
				a.model.HostTextInput.Blur()
				a.model.Output = "Host edit cancelled"
				return a, nil
			case "enter":
				// Save the new host
				newHost := a.model.HostTextInput.Value()
				err := a.configManager.SetHost(newHost)
				if err != nil {
					a.model.Output = fmt.Sprintf("Error setting host: %v", err)
				} else {
					// Update display and close popup
					a.model.Host = newHost
					a.model.Output = "Host updated successfully"
				}

				a.model.ShowHostPopup = false
				a.model.HostTextInput.Blur()
				return a, nil
			default:
				// Handle all other input through the textinput model
				var cmd tea.Cmd
				a.model.HostTextInput, cmd = a.model.HostTextInput.Update(msg)
				return a, cmd
			}
		}

		// Handle reconciliation type popup
		if a.model.ShowReconciliationTypePopup {
			switch msg.String() {
			case "esc":
				a.model.ShowReconciliationTypePopup = false
				a.model.SelectedReconciliationType = 0
				a.model.Output = "Reconciliation type edit cancelled"
				return a, nil
			case "up", "k":
				a.model.SelectedReconciliationType = (a.model.SelectedReconciliationType - 1 + 3) % 3
				return a, nil
			case "down", "j":
				a.model.SelectedReconciliationType = (a.model.SelectedReconciliationType + 1) % 3
				return a, nil
			case "enter":
				// Get the selected reconciliation type
				reconciliationTypes := []string{"can_be_reconciled_without_data", "reconciliation_not_necessary", "only_reconciled_with_data"}
				selectedType := reconciliationTypes[a.model.SelectedReconciliationType]

				// Get the current template
				if len(a.model.FilteredTemplates) > 0 && a.model.SelectedTemplate < len(a.model.FilteredTemplates) {
					actualIndex := a.model.FilteredTemplates[a.model.SelectedTemplate]
					template := a.model.Templates[actualIndex]

					// Update the config and save to file
					err := a.configManager.UpdateReconciliationType(template.Path, selectedType)
					if err != nil {
						a.model.Output = fmt.Sprintf("Error updating reconciliation type: %v", err)
					} else {
						// Update the in-memory template config
						a.model.Templates[actualIndex].Config["reconciliation_type"] = selectedType
						a.model.Output = fmt.Sprintf("Reconciliation type set to: %s", selectedType)
					}
				}

				a.model.ShowReconciliationTypePopup = false
				a.model.SelectedReconciliationType = 0
				return a, nil
			}
			return a, nil
		}

		// Handle text part popup
		if a.model.ShowTextPartPopup {
			switch msg.String() {
			case "esc":
				a.model.ShowTextPartPopup = false
				a.model.TextPartEditMode = ""
				a.model.Output = "Text part edit cancelled"
				return a, nil
			case "tab":
				// Switch between name and path fields
				if a.model.TextPartEditMode == "name" {
					a.model.TextPartEditMode = "path"
					a.model.TextPartNameInput.Blur()
					a.model.TextPartPathInput.Focus()
				} else {
					a.model.TextPartEditMode = "name"
					a.model.TextPartPathInput.Blur()
					a.model.TextPartNameInput.Focus()
				}
				return a, nil
			case "enter":
				// Save the text part changes (placeholder for now)
				newName := a.model.TextPartNameInput.Value()
				newPath := a.model.TextPartPathInput.Value()

				a.model.ShowTextPartPopup = false
				a.model.TextPartEditMode = ""
				a.model.Output = fmt.Sprintf("Text part updated: %s -> %s", newName, newPath)
				return a, nil
			default:
				// Handle text input
				var cmd tea.Cmd
				if a.model.TextPartEditMode == "name" {
					a.model.TextPartNameInput, cmd = a.model.TextPartNameInput.Update(msg)
				} else {
					a.model.TextPartPathInput, cmd = a.model.TextPartPathInput.Update(msg)
				}
				return a, cmd
			}
		}

		// Handle special cases first
		if msg.Alt {
			return a, nil // Skip alt combinations
		}

		// Handle search mode input
		if a.model.SearchMode {
			switch msg.String() {
			case "esc":
				// Exit search mode
				a.model.SearchMode = false
				a.model.SearchQuery = ""
				a.model.FilteredTemplates = a.templateManager.FilterTemplates(a.model.Templates, a.model.SearchQuery)
				a.model.SelectedTemplate = 0
				a.model.TemplatesOffset = 0
				a.model.Output = "Search cancelled"
				return a, nil
			case "enter":
				// Exit search mode but keep filter
				a.model.SearchMode = false
				if len(a.model.FilteredTemplates) > 0 {
					a.model.Output = fmt.Sprintf("Found %d templates", len(a.model.FilteredTemplates))
				} else {
					a.model.Output = "No templates found"
				}
				return a, nil
			case "backspace":
				// Remove last character
				if len(a.model.SearchQuery) > 0 {
					a.model.SearchQuery = a.model.SearchQuery[:len(a.model.SearchQuery)-1]
					a.model.FilteredTemplates = a.templateManager.FilterTemplates(a.model.Templates, a.model.SearchQuery)
					a.model.SelectedTemplate = 0
					a.model.TemplatesOffset = 0
				}
				return a, nil
			default:
				// Add typed character to search query
				if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
					char := msg.Runes[0]
					// Only allow printable characters
					if char >= 32 && char < 127 {
						a.model.SearchQuery += string(char)
						a.model.FilteredTemplates = a.templateManager.FilterTemplates(a.model.Templates, a.model.SearchQuery)
						a.model.SelectedTemplate = 0
						a.model.TemplatesOffset = 0
					}
				}
				return a, nil
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit

		case "?":
			a.model.ShowHelp = true
			return a, nil

		case "tab":
			a.navHandler.NextSection(a.model)
			return a, nil

		case "shift+tab":
			a.navHandler.PrevSection(a.model)
			return a, nil

		case "shift+up", "K":
			a.navHandler.HandleVerticalUp(a.model)
			return a, nil

		case "shift+down", "J":
			a.navHandler.HandleVerticalDown(a.model)
			return a, nil

		case "shift+left", "H":
			a.navHandler.PrevSection(a.model)
			return a, nil

		case "shift+right", "L":
			a.navHandler.NextSection(a.model)
			return a, nil

		case "up", "k":
			if a.model.CurrentSection == models.TemplatesSection {
				a.navHandler.HandleTemplateNavigation(a.model, "up")
			} else if a.model.CurrentSection == models.DetailsSection {
				a.navHandler.HandleDetailsNavigation(a.model, "up")
			}
			return a, nil

		case "down", "j":
			if a.model.CurrentSection == models.TemplatesSection {
				a.navHandler.HandleTemplateNavigation(a.model, "down")
			} else if a.model.CurrentSection == models.DetailsSection {
				a.navHandler.HandleDetailsNavigation(a.model, "down")
			}
			return a, nil

		case "left", "h":
			// Only allow within Templates section for now (could be used for future horizontal navigation within templates)
			// No section navigation without shift
			return a, nil

		case "right", "l":
			// Only allow within Templates section for now (could be used for future horizontal navigation within templates)
			// No section navigation without shift
			return a, nil

		case " ": // Space key for template selection
			if a.model.CurrentSection == models.TemplatesSection && len(a.model.FilteredTemplates) > 0 && a.model.SelectedTemplate < len(a.model.FilteredTemplates) {
				// Get actual template index from filtered list
				actualIndex := a.model.FilteredTemplates[a.model.SelectedTemplate]

				// Toggle selection for current template
				if a.model.SelectedTemplates[actualIndex] {
					delete(a.model.SelectedTemplates, actualIndex)
				} else {
					a.model.SelectedTemplates[actualIndex] = true
				}

				// Update output to show selection count
				selectedCount := len(a.model.SelectedTemplates)
				if selectedCount == 0 {
					a.model.Output = "No templates selected"
				} else if selectedCount == 1 {
					a.model.Output = "1 template selected"
				} else {
					a.model.Output = fmt.Sprintf("%d templates selected", selectedCount)
				}
			}
			return a, nil

		case "/": // Enter search mode
			if a.model.CurrentSection == models.TemplatesSection {
				a.model.SearchMode = true
				a.model.SearchQuery = ""
				a.model.Output = "Search mode - type to filter templates"
			}
			return a, nil

		case "backspace": // Deselect all templates
			if a.model.CurrentSection == models.TemplatesSection && len(a.model.SelectedTemplates) > 0 {
				a.model.SelectedTemplates = make(map[int]bool)
				a.model.Output = "All templates deselected"
			}
			return a, nil

		case "enter": // Trigger action popup when templates are selected, firm popup when in firm section, or host popup when in host section
			if a.model.CurrentSection == models.FirmSection {
				a.model.ShowFirmPopup = true
				a.model.SelectedFirm = 0
				a.model.Output = "Select a firm or partner"
			} else if a.model.CurrentSection == models.HostSection {
				a.model.ShowHostPopup = true
				// Set current host value and focus the input
				a.model.HostTextInput.SetValue(a.model.Host)
				a.model.HostTextInput.Focus()
				a.model.HostTextInput.CursorEnd()
				a.model.Output = "Edit host URL"
			} else if a.model.CurrentSection == models.DetailsSection {
				// Handle enter in Details section for config field editing
				if len(a.model.FilteredTemplates) > 0 && a.model.SelectedTemplate < len(a.model.FilteredTemplates) {
					actualIndex := a.model.FilteredTemplates[a.model.SelectedTemplate]
					template := a.model.Templates[actualIndex]

					// Get config field count using the same logic as renderer and navigation
					configFieldCount := a.getConfigFieldCount(template)

					if a.model.SelectedDetailField < configFieldCount {
						// We're on a config field - determine which specific field
						selectedConfigField := a.getSelectedConfigField(template, a.model.SelectedDetailField)
						
						// Only show reconciliation_type popup if that specific field is selected
						if selectedConfigField == "reconciliation_type" && template.Category == "reconciliation_texts" {
							a.model.ShowReconciliationTypePopup = true
							a.model.SelectedReconciliationType = 0
							// Set current selection based on current value
							currentValue := template.Config["reconciliation_type"]
							reconciliationTypes := []string{"can_be_reconciled_without_data", "reconciliation_not_necessary", "only_reconciled_with_data"}
							for i, rType := range reconciliationTypes {
								if currentValue == rType {
									a.model.SelectedReconciliationType = i
									break
								}
							}
							a.model.Output = "Select reconciliation type"
						}
					} else if template.Category != "shared_parts" {
						// We're on a text part (only for templates that support them)
						if textPartsInterface, exists := template.Config["text_parts"]; exists {
							if textParts, ok := textPartsInterface.(map[string]interface{}); ok {
								textPartIndex := a.model.SelectedDetailField - configFieldCount

								// Convert map to sorted slice to get the correct text part
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

								// Sort by name for consistency
								for i := 0; i < len(partsList); i++ {
									for j := i + 1; j < len(partsList); j++ {
										if partsList[i].name > partsList[j].name {
											partsList[i], partsList[j] = partsList[j], partsList[i]
										}
									}
								}

								if textPartIndex < len(partsList) {
									selectedPart := partsList[textPartIndex]
									a.model.ShowTextPartPopup = true
									a.model.SelectedTextPart = textPartIndex

									// Initialize text inputs with current values
									a.model.TextPartNameInput.SetValue(selectedPart.name)
									a.model.TextPartPathInput.SetValue(selectedPart.path)
									a.model.TextPartNameInput.Focus()
									a.model.TextPartEditMode = "name"

									a.model.Output = "Edit text part name and path"
								}
							}
						}
					}
				}
			} else if a.model.CurrentSection == models.TemplatesSection && len(a.model.SelectedTemplates) > 0 {
				a.model.ShowActionPopup = true
				a.model.SelectedAction = 0
				selectedCount := len(a.model.SelectedTemplates)
				if selectedCount == 1 {
					a.model.Output = "1 template selected - choose action"
				} else {
					a.model.Output = fmt.Sprintf("%d templates selected - choose action", selectedCount)
				}
			}
			return a, nil
		}
	}

	return a, nil
}

func (a *app) View() string {
	if a.model.ShowHelp {
		return a.uiRenderer.HelpView(a.model)
	}

	if a.model.ShowActionPopup {
		return a.uiRenderer.ActionPopupView(a.model)
	}

	if a.model.ShowFirmPopup {
		return a.uiRenderer.FirmPopupView(a.model)
	}

	if a.model.ShowHostPopup {
		return a.uiRenderer.HostPopupView(a.model)
	}

	if a.model.ShowReconciliationTypePopup {
		return a.uiRenderer.ReconciliationTypePopupView(a.model)
	}

	if a.model.ShowTextPartPopup {
		return a.uiRenderer.TextPartPopupView(a.model)
	}

	if a.model.Height < 10 {
		return "Terminal too small"
	}

	// Add search bar if in search mode
	var searchBar string
	searchBarHeight := 0
	if a.model.SearchMode {
		searchBarHeight = 3
		searchStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("6")). // Cyan like active sections
			Padding(0, 1)

		searchContent := fmt.Sprintf("Search: %s_", a.model.SearchQuery) // Add cursor
		searchBar = searchStyle.Width(a.model.Width - 4).Render(searchContent)
	}

	// Use much smaller dimensions to account for lipgloss borders and padding
	// Each bordered section uses roughly: content + title + 3 lines for borders
	topContentHeight := 1
	outputContentHeight := 1
	statusHeight := 1

	// Reserve 4 lines per section (1 content + 3 for borders/title)
	reservedLines := (4 * 2) + 4 + statusHeight + searchBarHeight // top sections + output + status + search
	availableContentHeight := a.model.Height - reservedLines
	if availableContentHeight < 1 {
		availableContentHeight = 1
	}

	// Use smaller widths to account for borders
	halfWidth := (a.model.Width - 6) / 2 // Account for borders and spacing
	fullWidth := a.model.Width - 4       // Account for borders

	// Top row: Firm and Host
	firmBox := a.uiRenderer.RenderSection(a.model, models.FirmSection, "Firm", a.uiRenderer.FirmView(a.model), halfWidth, topContentHeight)
	hostBox := a.uiRenderer.RenderSection(a.model, models.HostSection, "Host", a.uiRenderer.HostView(a.model), halfWidth, topContentHeight)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, firmBox, hostBox)

	// Main row: Templates and Details with height and width constraints
	selectedCount := len(a.model.SelectedTemplates)
	var templatesTitle string

	templateCount := len(a.model.FilteredTemplates)
	totalCount := len(a.model.Templates)

	if a.model.SearchMode && a.model.SearchQuery != "" {
		if selectedCount > 0 {
			templatesTitle = fmt.Sprintf("Templates (%d/%d) - %d selected", templateCount, totalCount, selectedCount)
		} else {
			templatesTitle = fmt.Sprintf("Templates (%d/%d)", templateCount, totalCount)
		}
	} else {
		if selectedCount > 0 {
			templatesTitle = fmt.Sprintf("Templates (%d) - %d selected", totalCount, selectedCount)
		} else {
			templatesTitle = fmt.Sprintf("Templates (%d)", totalCount)
		}
	}
	templatesContent := a.uiRenderer.TemplatesViewWithHeightAndWidth(a.model, availableContentHeight, halfWidth)
	detailsContent := a.uiRenderer.DetailsViewWithHeightAndWidth(a.model, availableContentHeight, halfWidth)

	templatesBox := a.uiRenderer.RenderSection(a.model, models.TemplatesSection, templatesTitle, templatesContent, halfWidth, availableContentHeight)
	detailsBox := a.uiRenderer.RenderSection(a.model, models.DetailsSection, "Details", detailsContent, halfWidth, availableContentHeight)
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, templatesBox, detailsBox)

	// Output row
	outputBox := a.uiRenderer.RenderSection(a.model, models.OutputSection, "Output", a.uiRenderer.OutputView(a.model), fullWidth, outputContentHeight)

	// Status bar (no borders)
	statusBar := a.uiRenderer.StatusBarView(a.model)

	// Assemble final view
	if a.model.SearchMode {
		return lipgloss.JoinVertical(lipgloss.Left, searchBar, topRow, mainRow, outputBox, statusBar)
	} else {
		return lipgloss.JoinVertical(lipgloss.Left, topRow, mainRow, outputBox, statusBar)
	}
}

// getConfigFieldCount returns the number of config fields that will be displayed for a template
func (a *app) getConfigFieldCount(template models.Template) int {
	count := 0

	// Define the fixed keys to show for all template types (same as in renderer)
	configKeys := []string{
		"public",
		"reconciliation_type",
		"virtual_account_number",
		"allow_duplicate_reconciliation",
		"is_active",
		"use_full_width",
		"downloadable_as_docx",
		"encoding",
		"published",
		"hide_code",
		"externally_managed",
	}

	// Count how many of these keys exist in the template config
	for _, key := range configKeys {
		if _, exists := template.Config[key]; exists {
			count++
		}
	}

	return count
}

// getSelectedConfigField returns the config field name at the given index
func (a *app) getSelectedConfigField(template models.Template, fieldIndex int) string {
	currentIndex := 0

	// Define the fixed keys to show for all template types (same as in renderer)
	configKeys := []string{
		"public",
		"reconciliation_type",
		"virtual_account_number",
		"allow_duplicate_reconciliation",
		"is_active",
		"use_full_width",
		"downloadable_as_docx",
		"encoding",
		"published",
		"hide_code",
		"externally_managed",
	}

	// Find which config field corresponds to the given index
	for _, key := range configKeys {
		if _, exists := template.Config[key]; exists {
			if currentIndex == fieldIndex {
				return key
			}
			currentIndex++
		}
	}

	return "" // Field index out of range
}

// buildSharedPartsMapping builds a mapping of which shared parts each template uses
// by analyzing the used_in field of all shared parts
func (a *app) buildSharedPartsMapping() {
	// Initialize the mapping
	a.model.SharedPartsUsage = make(map[string][]string)

	// Find all shared parts templates
	for _, template := range a.model.Templates {
		if template.Category == "shared_parts" {
			// Check if this shared part has a used_in section
			if usedInInterface, exists := template.Config["used_in"]; exists {
				if usedInArray, ok := usedInInterface.([]interface{}); ok {
					// Process each usage entry
					for _, usageInterface := range usedInArray {
						if usage, ok := usageInterface.(map[string]interface{}); ok {
							// Get the type and handle from the usage entry
							if typeStr, typeExists := usage["type"].(string); typeExists {
								if handleStr, handleExists := usage["handle"].(string); handleExists {
									// Convert type to match template categories
									var templateCategory string
									switch typeStr {
									case "reconciliation_text":
										templateCategory = "reconciliation_texts"
									case "export_file":
										templateCategory = "export_files"
									case "account_template":
										templateCategory = "account_templates"
									default:
										continue // Skip unknown types
									}

									// Create a unique template identifier
									templateKey := templateCategory + "/" + handleStr

									// Add this shared part to the template's list
									if a.model.SharedPartsUsage[templateKey] == nil {
										a.model.SharedPartsUsage[templateKey] = []string{}
									}
									a.model.SharedPartsUsage[templateKey] = append(a.model.SharedPartsUsage[templateKey], template.Name)
								}
							}
						}
					}
				}
			}
		}
	}
}

func main() {
	app := newApp()
	app.initialModel()

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}

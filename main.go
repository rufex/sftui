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

	a.model = &models.Model{
		CurrentSection:    models.TemplatesSection,
		SelectedTemplate:  0,
		Templates:         []models.Template{},
		SelectedTemplates: make(map[int]bool),
		Firm:              "No firm set",
		Host:              "No host set",
		HostTextInput:     hostTextInput,
		ShowHelp:          false,
		Output:            "Ready",
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
			a.navHandler.HandleTemplateNavigation(a.model, "up")
			return a, nil

		case "down", "j":
			a.navHandler.HandleTemplateNavigation(a.model, "down")
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

func main() {
	app := newApp()
	app.initialModel()

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}

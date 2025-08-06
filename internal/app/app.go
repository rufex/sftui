package app

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/rufex/sftui/internal/models"
	"github.com/rufex/sftui/internal/navigation"
	"github.com/rufex/sftui/internal/template"
	"github.com/rufex/sftui/internal/ui"
)

type App struct {
	Model           *models.Model
	templateManager *template.Manager
	configManager   *template.ConfigManager
	navHandler      *navigation.Handler
	uiRenderer      *ui.Renderer
}

func New() *App {
	return &App{
		Model:           &models.Model{},
		templateManager: template.NewManager(),
		configManager:   template.NewConfigManager(),
		navHandler:      navigation.NewHandler(),
		uiRenderer:      ui.NewRenderer(),
	}
}

func (a *App) InitialModel() *models.Model {
	hostTextInput := textinput.New()
	hostTextInput.Placeholder = "Enter host URL (e.g., https://api.example.com)"
	hostTextInput.CharLimit = 256
	hostTextInput.Width = 50

	textPartNameInput := textinput.New()
	textPartNameInput.Placeholder = "Enter text part name"
	textPartNameInput.CharLimit = 100
	textPartNameInput.Width = 40

	textPartPathInput := textinput.New()
	textPartPathInput.Placeholder = "Enter text part path"
	textPartPathInput.CharLimit = 256
	textPartPathInput.Width = 40

	a.Model = &models.Model{
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

	firm, host, output := a.configManager.LoadSilverfinConfig()
	a.Model.Firm = firm
	a.Model.Host = host
	a.Model.Output = output

	firmOptions, err := a.configManager.LoadFirmOptions()
	if err != nil {
		a.Model.Output = "Error loading firm options"
	} else {
		a.Model.FirmOptions = firmOptions
	}

	a.Model.Templates = a.templateManager.LoadTemplates()
	a.buildSharedPartsMapping()
	a.Model.FilteredTemplates = a.templateManager.FilterTemplates(a.Model.Templates, "")

	return a.Model
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) buildSharedPartsMapping() {
	a.Model.SharedPartsUsage = make(map[string][]string)

	for _, template := range a.Model.Templates {
		if template.Category == "shared_parts" {
			if usedInInterface, exists := template.Config["used_in"]; exists {
				if usedInArray, ok := usedInInterface.([]interface{}); ok {
					for _, usageInterface := range usedInArray {
						if usage, ok := usageInterface.(map[string]interface{}); ok {
							if typeStr, typeExists := usage["type"].(string); typeExists {
								if handleStr, handleExists := usage["handle"].(string); handleExists {
									var templateCategory string
									switch typeStr {
									case "reconciliation_text":
										templateCategory = "reconciliation_texts"
									case "export_file":
										templateCategory = "export_files"
									case "account_template":
										templateCategory = "account_templates"
									default:
										continue
									}

									templateKey := templateCategory + "/" + handleStr

									if a.Model.SharedPartsUsage[templateKey] == nil {
										a.Model.SharedPartsUsage[templateKey] = []string{}
									}
									a.Model.SharedPartsUsage[templateKey] = append(a.Model.SharedPartsUsage[templateKey], template.Name)
								}
							}
						}
					}
				}
			}
		}
	}
}

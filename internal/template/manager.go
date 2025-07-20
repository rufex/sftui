package template

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rufex/sftui/internal/models"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) LoadTemplates() []models.Template {
	var templates []models.Template

	// Check if fixtures directory exists, use it for testing
	if _, err := os.Stat("fixtures/market-repo"); err == nil {
		templates = m.scanDirectory("fixtures/market-repo")
	} else {
		// Scan current directory for templates
		templates = m.scanDirectory(".")
	}

	return templates
}

func (m *Manager) scanDirectory(rootPath string) []models.Template {
	var templates []models.Template

	categories := []string{"account_templates", "reconciliation_texts", "export_files", "shared_parts"}

	for _, category := range categories {
		categoryPath := filepath.Join(rootPath, category)
		if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
			continue
		}

		filepath.WalkDir(categoryPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if d.Name() == "config.json" {
				templateDir := filepath.Dir(path)
				templateName := filepath.Base(templateDir)

				// Load config
				config := make(map[string]interface{})
				if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
					json.Unmarshal(data, &config)
				}

				template := models.Template{
					Name:     templateName,
					Path:     templateDir,
					Type:     category,
					Category: category,
					Config:   config,
				}
				templates = append(templates, template)
			}
			return nil
		})
	}

	return templates
}

func (m *Manager) GetCategoryPrefix(category string) string {
	switch category {
	case "account_templates":
		return "AT"
	case "export_files":
		return "EF"
	case "reconciliation_texts":
		return "RT"
	case "shared_parts":
		return "SP"
	default:
		return "?"
	}
}

func (m *Manager) GetCategoryDisplayName(category string) string {
	switch category {
	case "account_templates":
		return "Account Template"
	case "export_files":
		return "Export File"
	case "reconciliation_texts":
		return "Reconciliation Text"
	case "shared_parts":
		return "Shared Part"
	default:
		return category
	}
}

func (m *Manager) FuzzyMatch(query, target string) bool {
	if query == "" {
		return true
	}

	query = strings.ToLower(query)
	target = strings.ToLower(target)

	queryIdx := 0
	for i := 0; i < len(target) && queryIdx < len(query); i++ {
		if target[i] == query[queryIdx] {
			queryIdx++
		}
	}

	return queryIdx == len(query)
}

func (m *Manager) FilterTemplates(templates []models.Template, searchQuery string) []int {
	if searchQuery == "" {
		// No search query, show all templates
		filteredTemplates := make([]int, len(templates))
		for i := range templates {
			filteredTemplates[i] = i
		}
		return filteredTemplates
	}

	filteredTemplates := []int{}
	for i, template := range templates {
		// Match against template name, category, and path
		if m.FuzzyMatch(searchQuery, template.Name) ||
			m.FuzzyMatch(searchQuery, template.Category) ||
			m.FuzzyMatch(searchQuery, template.Path) {
			filteredTemplates = append(filteredTemplates, i)
		}
	}
	return filteredTemplates
}

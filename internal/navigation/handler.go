package navigation

import (
	"github.com/rufex/sftui/internal/models"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) AdjustScrolling(m *models.Model) {
	// Calculate available height for templates content
	searchBarHeight := 0
	if m.SearchMode {
		searchBarHeight = 3
	}
	reservedLines := (4 * 2) + 4 + 1 + searchBarHeight // top sections + output + status + search
	availableContentHeight := m.Height - reservedLines
	if availableContentHeight < 1 {
		availableContentHeight = 1
	}

	// Ensure selected template is visible
	if m.SelectedTemplate < m.TemplatesOffset {
		m.TemplatesOffset = m.SelectedTemplate
	} else if m.SelectedTemplate >= m.TemplatesOffset+availableContentHeight {
		m.TemplatesOffset = m.SelectedTemplate - availableContentHeight + 1
	}

	// Ensure offset is within bounds (use filtered templates count)
	maxOffset := max(0, len(m.FilteredTemplates)-availableContentHeight)
	if m.TemplatesOffset > maxOffset {
		m.TemplatesOffset = maxOffset
	}
	if m.TemplatesOffset < 0 {
		m.TemplatesOffset = 0
	}
}

func (h *Handler) HandleVerticalUp(m *models.Model) {
	switch m.CurrentSection {
	case models.TemplatesSection:
		m.CurrentSection = models.FirmSection
	case models.DetailsSection:
		m.CurrentSection = models.HostSection
	case models.OutputSection:
		// From Output, go to Templates (left side of main row)
		m.CurrentSection = models.TemplatesSection
	case models.FirmSection:
		// Already at top row, no vertical movement
	case models.HostSection:
		// Already at top row, no vertical movement
	}
}

func (h *Handler) HandleVerticalDown(m *models.Model) {
	switch m.CurrentSection {
	case models.FirmSection:
		m.CurrentSection = models.TemplatesSection
	case models.HostSection:
		m.CurrentSection = models.DetailsSection
	case models.TemplatesSection:
		m.CurrentSection = models.OutputSection
	case models.DetailsSection:
		m.CurrentSection = models.OutputSection
	case models.OutputSection:
		// Already at bottom, no vertical movement
	}
}

func (h *Handler) HandleTemplateNavigation(m *models.Model, direction string) {
	if m.CurrentSection != models.TemplatesSection || len(m.FilteredTemplates) == 0 {
		return
	}

	switch direction {
	case "up":
		m.SelectedTemplate = (m.SelectedTemplate - 1 + len(m.FilteredTemplates)) % len(m.FilteredTemplates)
		h.AdjustScrolling(m)
	case "down":
		m.SelectedTemplate = (m.SelectedTemplate + 1) % len(m.FilteredTemplates)
		h.AdjustScrolling(m)
	}
}

func (h *Handler) HandleDetailsNavigation(m *models.Model, direction string) {
	if m.CurrentSection != models.DetailsSection || len(m.FilteredTemplates) == 0 {
		return
	}

	// Get actual template index from filtered list
	actualIndex := m.FilteredTemplates[m.SelectedTemplate]
	template := m.Templates[actualIndex]

	// Count available config fields for navigation
	fieldCount := h.GetConfigFieldCount(template)
	if fieldCount == 0 {
		return
	}

	switch direction {
	case "up":
		m.SelectedDetailField = (m.SelectedDetailField - 1 + fieldCount) % fieldCount
	case "down":
		m.SelectedDetailField = (m.SelectedDetailField + 1) % fieldCount
	}
}

func (h *Handler) GetConfigFieldCount(template models.Template) int {
	// For reconciliation_texts, we only show reconciliation_type field
	if template.Category == "reconciliation_texts" {
		if _, exists := template.Config["reconciliation_type"]; exists {
			return 1
		}
	}
	return 0
}

func (h *Handler) NextSection(m *models.Model) {
	m.CurrentSection = models.Section((int(m.CurrentSection) + 1) % 5)
}

func (h *Handler) PrevSection(m *models.Model) {
	m.CurrentSection = models.Section((int(m.CurrentSection) - 1 + 5) % 5)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

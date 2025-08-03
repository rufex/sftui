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

	// Count config fields (fixed keys for all template types)
	configFieldCount := h.GetActualConfigFieldCount(template)

	// Count text parts
	textPartsCount := h.GetTextPartsCount(template)

	// Count shared parts
	sharedPartsCount := h.GetSharedPartsCount(template, m.SharedPartsUsage)

	totalFieldCount := configFieldCount + textPartsCount + sharedPartsCount
	if totalFieldCount == 0 {
		return
	}

	switch direction {
	case "up":
		m.SelectedDetailField = (m.SelectedDetailField - 1 + totalFieldCount) % totalFieldCount
	case "down":
		m.SelectedDetailField = (m.SelectedDetailField + 1) % totalFieldCount
	}

	// Update text part selection based on which field is selected
	if m.SelectedDetailField >= configFieldCount {
		// We're in the text parts section
		m.SelectedTextPart = m.SelectedDetailField - configFieldCount
	} else {
		// We're in the config section
		m.SelectedTextPart = -1 // No text part selected
	}
}

func (h *Handler) GetActualConfigFieldCount(template models.Template) int {
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

func (h *Handler) GetConfigFieldCount(template models.Template, sharedPartsUsage map[string][]string) int {
	count := h.GetActualConfigFieldCount(template)

	// Add text parts count (for templates that support them, excluding shared_parts)
	if template.Category != "shared_parts" {
		if textPartsInterface, exists := template.Config["text_parts"]; exists {
			if textParts, ok := textPartsInterface.(map[string]interface{}); ok {
				count += len(textParts)
			}
		}
	}

	// Add shared parts count
	count += h.GetSharedPartsCount(template, sharedPartsUsage)

	return count
}

func (h *Handler) GetSharedPartsCount(template models.Template, sharedPartsUsage map[string][]string) int {
	if template.Category == "shared_parts" {
		return 0
	}

	templateKey := template.Category + "/" + template.Name
	if sharedParts, exists := sharedPartsUsage[templateKey]; exists {
		return len(sharedParts)
	}
	return 0
}

func (h *Handler) GetTextPartsCount(template models.Template) int {
	if template.Category == "shared_parts" {
		return 0
	}

	if textPartsInterface, exists := template.Config["text_parts"]; exists {
		if textParts, ok := textPartsInterface.(map[string]interface{}); ok {
			return len(textParts)
		}
	}
	return 0
}

func (h *Handler) HandleTextPartNavigation(m *models.Model, direction string) {
	if m.CurrentSection != models.DetailsSection || len(m.FilteredTemplates) == 0 {
		return
	}

	// Get actual template index from filtered list
	actualIndex := m.FilteredTemplates[m.SelectedTemplate]
	template := m.Templates[actualIndex]

	textPartsCount := h.GetTextPartsCount(template)
	if textPartsCount == 0 {
		return
	}

	switch direction {
	case "up":
		m.SelectedTextPart = (m.SelectedTextPart - 1 + textPartsCount) % textPartsCount
	case "down":
		m.SelectedTextPart = (m.SelectedTextPart + 1) % textPartsCount
	}
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

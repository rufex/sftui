package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/rufex/sftui/internal/models"
	"github.com/rufex/sftui/internal/template"
)

type Renderer struct {
	templateManager *template.Manager
}

func NewRenderer() *Renderer {
	return &Renderer{
		templateManager: template.NewManager(),
	}
}

func (r *Renderer) RenderSection(m *models.Model, section models.Section, title, content string, width, height int) string {
	style := models.InactiveBorderStyle
	if m.CurrentSection == section {
		style = models.ActiveBorderStyle
	}

	titleStr := models.TitleStyle.Render(title)
	contentWithTitle := lipgloss.JoinVertical(lipgloss.Left, titleStr, content)

	return style.Width(width).Height(height).Render(contentWithTitle)
}

func (r *Renderer) FirmView(m *models.Model) string {
	return m.Firm
}

func (r *Renderer) HostView(m *models.Model) string {
	if m.Host == "" {
		return "No host set"
	}
	return m.Host
}

func (r *Renderer) TemplatesView(m *models.Model) string {
	return r.templatesViewWithHeightAndWidth(m, -1, -1)
}

func (r *Renderer) TemplatesViewWithHeight(m *models.Model, maxHeight int) string {
	return r.templatesViewWithHeightAndWidth(m, maxHeight, -1)
}

func (r *Renderer) TemplatesViewWithHeightAndWidth(m *models.Model, maxHeight, maxWidth int) string {
	return r.templatesViewWithHeightAndWidth(m, maxHeight, maxWidth)
}

func (r *Renderer) templatesViewWithHeightAndWidth(m *models.Model, maxHeight, maxWidth int) string {
	if len(m.FilteredTemplates) == 0 {
		if m.SearchMode && m.SearchQuery != "" {
			return "No templates match search"
		}
		return "No templates found"
	}

	var lines []string

	for i, templateIdx := range m.FilteredTemplates {
		template := m.Templates[templateIdx]
		prefix := r.templateManager.GetCategoryPrefix(template.Category)

		// Add selection indicator
		selectionIndicator := "  " // Two spaces for unselected
		if m.SelectedTemplates[templateIdx] {
			selectionIndicator = "✓ " // Checkmark for selected
		}

		line := fmt.Sprintf("%s[%s] %s", selectionIndicator, prefix, template.Name)

		// Apply horizontal truncation if width limit is specified
		if maxWidth > 0 {
			line = r.TruncateText(line, maxWidth)
		}

		if i == m.SelectedTemplate {
			line = models.SelectedItemStyle.Render(line)
		}
		lines = append(lines, line)
	}

	// If no height limit, return all lines
	if maxHeight <= 0 {
		return strings.Join(lines, "\n")
	}

	// Apply scrolling and height limit
	startIdx := m.TemplatesOffset
	endIdx := startIdx + maxHeight

	if startIdx >= len(lines) {
		startIdx = max(0, len(lines)-maxHeight)
	}
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	visibleLines := lines[startIdx:endIdx]

	// Pad with empty lines if needed to maintain consistent height
	for len(visibleLines) < maxHeight {
		visibleLines = append(visibleLines, "")
	}

	return strings.Join(visibleLines, "\n")
}

func (r *Renderer) DetailsView(m *models.Model) string {
	return r.detailsViewWithHeightAndWidth(m, -1, -1)
}

func (r *Renderer) DetailsViewWithHeight(m *models.Model, maxHeight int) string {
	return r.detailsViewWithHeightAndWidth(m, maxHeight, -1)
}

func (r *Renderer) DetailsViewWithHeightAndWidth(m *models.Model, maxHeight, maxWidth int) string {
	return r.detailsViewWithHeightAndWidth(m, maxHeight, maxWidth)
}

func (r *Renderer) detailsViewWithHeightAndWidth(m *models.Model, maxHeight, maxWidth int) string {
	if len(m.FilteredTemplates) == 0 || m.SelectedTemplate >= len(m.FilteredTemplates) {
		return "No template selected"
	}

	// Get actual template index from filtered list
	actualIndex := m.FilteredTemplates[m.SelectedTemplate]
	template := m.Templates[actualIndex]

	var details []string

	// Apply horizontal truncation to each line if width limit is specified
	nameStr := fmt.Sprintf("Name: %s", template.Name)
	if maxWidth > 0 {
		nameStr = r.TruncateText(nameStr, maxWidth)
	}
	details = append(details, nameStr)

	displayName := r.templateManager.GetCategoryDisplayName(template.Category)
	categoryStr := fmt.Sprintf("Type: %s", displayName)
	if maxWidth > 0 {
		categoryStr = r.TruncateText(categoryStr, maxWidth)
	}
	details = append(details, categoryStr)

	pathStr := fmt.Sprintf("Path: %s", template.Path)
	if maxWidth > 0 {
		pathStr = r.TruncateText(pathStr, maxWidth)
	}
	details = append(details, pathStr)

	details = append(details, "")
	details = append(details, "Configuration:")

	if len(template.Config) == 0 {
		details = append(details, "  (empty)")
	} else {
		// Show structured config fields for all template types
		structuredConfig := r.renderStructuredConfig(template, m, maxWidth)
		if structuredConfig != "" {
			// Split the structured config into individual lines
			configLines := strings.Split(structuredConfig, "\n")
			details = append(details, configLines...)
		}
	}

	// Show text parts as a separate section for templates that support them (but not shared_parts)
	if template.Category != "shared_parts" {
		textPartsSection := r.renderTextPartsSection(template, m, maxWidth)
		if textPartsSection != "" {
			details = append(details, "")
			// Split the text parts section into individual lines
			textPartsLines := strings.Split(textPartsSection, "\n")
			details = append(details, textPartsLines...)
		}
	}

	// Show shared parts as a separate section for templates that can use them (but not shared_parts themselves)
	if template.Category != "shared_parts" {
		sharedPartsSection := r.renderSharedPartsSection(template, m, maxWidth)
		if sharedPartsSection != "" {
			details = append(details, "")
			// Split the shared parts section into individual lines
			sharedPartsLines := strings.Split(sharedPartsSection, "\n")
			details = append(details, sharedPartsLines...)
		}
	}

	// If no height limit, return all details
	if maxHeight <= 0 {
		return strings.Join(details, "\n")
	}

	// Apply height limit - truncate if needed
	if len(details) > maxHeight {
		details = details[:maxHeight]
	}

	// Pad with empty lines if needed to maintain consistent height
	for len(details) < maxHeight {
		details = append(details, "")
	}

	return strings.Join(details, "\n")
}

func (r *Renderer) renderStructuredConfig(template models.Template, m *models.Model, maxWidth int) string {
	var lines []string
	fieldIndex := 0

	// Define the fixed keys to show for all template types
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

	// Show each config key if it exists
	for _, key := range configKeys {
		if value, exists := template.Config[key]; exists {
			line := fmt.Sprintf("  %s: %v", key, value)
			if maxWidth > 0 {
				line = r.TruncateText(line, maxWidth)
			}

			// Highlight if this field is selected and we're in Details section
			if m.CurrentSection == models.DetailsSection && m.SelectedDetailField == fieldIndex {
				line = models.SelectedItemStyle.Render(line)
			}

			lines = append(lines, line)
			fieldIndex++
		}
	}

	return strings.Join(lines, "\n")
}

// GetConfigFieldCount returns the number of config fields that will be displayed
func (r *Renderer) GetConfigFieldCount(template models.Template) int {
	count := 0

	// Define the fixed keys to show for all template types (same as in renderStructuredConfig)
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

func (r *Renderer) renderTextPartsSection(template models.Template, m *models.Model, maxWidth int) string {
	// Check if template has text_parts in config
	textPartsInterface, exists := template.Config["text_parts"]
	if !exists {
		return ""
	}

	textParts, ok := textPartsInterface.(map[string]interface{})
	if !ok || len(textParts) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, "Text Parts:")

	// Convert map to sorted slice for consistent ordering
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

	// Calculate field index - config fields come first
	configFieldCount := r.GetConfigFieldCount(template)

	// Render each text part (only show name)
	for i, part := range partsList {
		line := fmt.Sprintf("  %s", part.name)
		if maxWidth > 0 {
			line = r.TruncateText(line, maxWidth)
		}

		// Highlight if this text part is selected and we're in Details section
		fieldIndex := configFieldCount + i
		if m.CurrentSection == models.DetailsSection && m.SelectedDetailField == fieldIndex {
			line = models.SelectedItemStyle.Render(line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (r *Renderer) renderSharedPartsSection(template models.Template, m *models.Model, maxWidth int) string {
	// Only show for templates that can use shared parts (not shared parts themselves)
	if template.Category == "shared_parts" {
		return ""
	}

	// Create template key to look up shared parts
	templateKey := template.Category + "/" + template.Name

	// Get shared parts for this template
	sharedParts, exists := m.SharedPartsUsage[templateKey]
	if !exists || len(sharedParts) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, "Shared Parts:")

	// Calculate field index - config fields and text parts come first
	configFieldCount := r.GetConfigFieldCount(template)

	textPartsCount := 0
	if textPartsInterface, exists := template.Config["text_parts"]; exists {
		if textParts, ok := textPartsInterface.(map[string]interface{}); ok {
			textPartsCount = len(textParts)
		}
	}

	// Sort shared parts for consistent display
	sortedSharedParts := make([]string, len(sharedParts))
	copy(sortedSharedParts, sharedParts)
	for i := 0; i < len(sortedSharedParts); i++ {
		for j := i + 1; j < len(sortedSharedParts); j++ {
			if sortedSharedParts[i] > sortedSharedParts[j] {
				sortedSharedParts[i], sortedSharedParts[j] = sortedSharedParts[j], sortedSharedParts[i]
			}
		}
	}

	// Render each shared part (only show name)
	for i, sharedPartName := range sortedSharedParts {
		line := fmt.Sprintf("  %s", sharedPartName)
		if maxWidth > 0 {
			line = r.TruncateText(line, maxWidth)
		}

		// Highlight if this shared part is selected and we're in Details section
		fieldIndex := configFieldCount + textPartsCount + i
		if m.CurrentSection == models.DetailsSection && m.SelectedDetailField == fieldIndex {
			line = models.SelectedItemStyle.Render(line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (r *Renderer) ReconciliationTypePopupView(m *models.Model) string {
	// Build popup content
	var content strings.Builder
	content.WriteString("Select Reconciliation Type\n\n")

	// Reconciliation type options
	reconciliationTypes := []string{"can_be_reconciled_without_data", "reconciliation_not_necessary", "only_reconciled_with_data"}
	for i, rType := range reconciliationTypes {
		if i == m.SelectedReconciliationType {
			// Highlight selected option
			content.WriteString(models.SelectedItemStyle.Render(fmt.Sprintf("> %s", rType)))
		} else {
			content.WriteString(fmt.Sprintf("  %s", rType))
		}
		content.WriteString("\n")
	}

	content.WriteString("\nUse ↑/↓ or k/j to navigate")
	content.WriteString("\nPress ENTER to select, ESC to cancel")

	// Calculate popup dimensions
	popupWidth := 50
	popupHeight := 10

	// Center the popup
	leftMargin := max(0, (m.Width-popupWidth)/2)
	topMargin := max(0, (m.Height-popupHeight)/2)

	popupBox := models.ActiveBorderStyle.
		Width(popupWidth).
		Height(popupHeight).
		Padding(1).
		Render(content.String())

	// Add top margin using newlines
	centeredPopup := strings.Repeat("\n", topMargin) + popupBox

	// Add left margin using spaces
	lines := strings.Split(centeredPopup, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = strings.Repeat(" ", leftMargin) + line
		}
	}

	return strings.Join(lines, "\n")
}

func (r *Renderer) OutputView(m *models.Model) string {
	return m.Output
}

func (r *Renderer) TruncateText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return text
	}

	// Account for borders and padding (approximately 4 characters)
	availableWidth := maxWidth - 4
	if availableWidth <= 3 {
		availableWidth = maxWidth // Fallback for very narrow sections
	}

	if len(text) <= availableWidth {
		return text
	}

	// Truncate and add ellipsis
	if availableWidth <= 3 {
		return text[:min(len(text), availableWidth)]
	}

	return text[:availableWidth-3] + "..."
}

func (r *Renderer) StatusBarView(m *models.Model) string {
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Padding(0, 1)

	status := "TAB: Next section • Shift+TAB: Prev section • ↑/k,↓/j: Navigate list • ?: Help • q: Quit"
	return statusStyle.Width(m.Width).Render(status)
}

func (r *Renderer) ActionPopupView(m *models.Model) string {
	selectedCount := len(m.SelectedTemplates)

	// Build popup content
	var content strings.Builder

	// Title with selected count
	if selectedCount == 1 {
		content.WriteString("1 template selected\n\n")
	} else {
		content.WriteString(fmt.Sprintf("%d templates selected\n\n", selectedCount))
	}

	// Action options
	actions := []string{"create", "import", "update", "cancel"}
	for i, action := range actions {
		if i == m.SelectedAction {
			// Highlight selected action
			content.WriteString(models.SelectedItemStyle.Render(fmt.Sprintf("> %s", action)))
		} else {
			content.WriteString(fmt.Sprintf("  %s", action))
		}
		content.WriteString("\n")
	}

	content.WriteString("\nUse ↑/↓ or k/j to navigate")
	content.WriteString("\nPress ENTER to select, ESC to cancel")

	// Calculate popup dimensions
	popupWidth := 40
	popupHeight := 12

	// Center the popup
	leftMargin := max(0, (m.Width-popupWidth)/2)
	topMargin := max(0, (m.Height-popupHeight)/2)

	popupBox := models.ActiveBorderStyle.
		Width(popupWidth).
		Height(popupHeight).
		Padding(1).
		Render(content.String())

	// Add top margin using newlines
	centeredPopup := strings.Repeat("\n", topMargin) + popupBox

	// Add left margin using spaces
	lines := strings.Split(centeredPopup, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = strings.Repeat(" ", leftMargin) + line
		}
	}

	return strings.Join(lines, "\n")
}

func (r *Renderer) FirmPopupView(m *models.Model) string {
	// Build popup content
	var content strings.Builder
	content.WriteString("Select Default Firm/Partner\n\n")

	// Firm and partner options
	for i, option := range m.FirmOptions {
		prefix := "  "
		if i == m.SelectedFirm {
			// Highlight selected option
			optionText := fmt.Sprintf("> [%s] %s (%s)", option.Type, option.Name, option.ID)
			content.WriteString(models.SelectedItemStyle.Render(optionText))
		} else {
			content.WriteString(fmt.Sprintf("%s[%s] %s (%s)", prefix, option.Type, option.Name, option.ID))
		}
		content.WriteString("\n")
	}

	if len(m.FirmOptions) == 0 {
		content.WriteString("No firms or partners available\n")
	}

	content.WriteString("\nUse ↑/↓ or k/j to navigate")
	content.WriteString("\nPress ENTER to select, ESC to cancel")

	// Calculate popup dimensions (wider for firm names)
	popupWidth := 50
	popupHeight := min(15, len(m.FirmOptions)+8) // Dynamic height based on options

	// Center the popup
	leftMargin := max(0, (m.Width-popupWidth)/2)
	topMargin := max(0, (m.Height-popupHeight)/2)

	popupBox := models.ActiveBorderStyle.
		Width(popupWidth).
		Height(popupHeight).
		Padding(1).
		Render(content.String())

	// Add top margin using newlines
	centeredPopup := strings.Repeat("\n", topMargin) + popupBox

	// Add left margin using spaces
	lines := strings.Split(centeredPopup, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = strings.Repeat(" ", leftMargin) + line
		}
	}

	return strings.Join(lines, "\n")
}

func (r *Renderer) HostPopupView(m *models.Model) string {
	// Build popup content
	var content strings.Builder
	content.WriteString("Edit Host URL\n\n")

	// Use the textinput view for a much better experience
	content.WriteString("Host: ")
	content.WriteString(m.HostTextInput.View())
	content.WriteString("\n\n")

	content.WriteString("Type to edit the host URL")
	content.WriteString("\nUse ←/→ to move cursor, Ctrl+A/E for start/end")
	content.WriteString("\nPress ENTER to save, ESC to cancel")

	// Calculate popup dimensions (wider for URL input)
	popupWidth := 70
	popupHeight := 10

	// Center the popup
	leftMargin := max(0, (m.Width-popupWidth)/2)
	topMargin := max(0, (m.Height-popupHeight)/2)

	popupBox := models.ActiveBorderStyle.
		Width(popupWidth).
		Height(popupHeight).
		Padding(1).
		Render(content.String())

	// Add top margin using newlines
	centeredPopup := strings.Repeat("\n", topMargin) + popupBox

	// Add left margin using spaces
	lines := strings.Split(centeredPopup, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = strings.Repeat(" ", leftMargin) + line
		}
	}

	return strings.Join(lines, "\n")
}

func (r *Renderer) HelpView(m *models.Model) string {
	help := `Key Bindings:

Navigation:
  TAB                     Navigate to next section
  Shift+TAB               Navigate to previous section
  Shift+↑/↓, Shift+K/J    Navigate vertically between sections
  Shift+←/→, Shift+H/L    Navigate horizontally between sections
  ↑/k, ↓/j, h, l          Navigate within Templates section only
  Space                   Select/deselect template (Templates section)
  Backspace               Deselect all templates (Templates section)
  /                       Enter search mode (Templates section)
  Enter                   Show actions for selected templates
  
Search Mode:
  Type                    Filter templates by name, category, or path
  Backspace               Remove last character
  Enter                   Exit search mode (keep filter)
  Escape                  Exit search mode (clear filter)
  
Actions:
  ?                       Show/hide this help
  q / Ctrl+C              Quit application

Sections:
  Firm                    Current firm configuration
  Host                    Current host configuration
  Templates               Lists all available templates by category
  Details                 Shows selected template configuration
  Output                  Shows application output

Press any key to close this help...`

	// Calculate center position
	helpWidth := 70
	helpHeight := 20

	// Center the popup
	leftMargin := max(0, (m.Width-helpWidth)/2)
	topMargin := max(0, (m.Height-helpHeight)/2)

	helpBox := models.ActiveBorderStyle.
		Width(helpWidth).
		Height(helpHeight).
		Padding(1).
		Render(help)

	// Add top margin using newlines
	centeredHelp := strings.Repeat("\n", topMargin) + helpBox

	// Add left margin using spaces
	lines := strings.Split(centeredHelp, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = strings.Repeat(" ", leftMargin) + line
		}
	}

	return strings.Join(lines, "\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (r *Renderer) TextPartPopupView(m *models.Model) string {
	// Get current template and text part
	if len(m.FilteredTemplates) == 0 || m.SelectedTemplate >= len(m.FilteredTemplates) {
		return ""
	}

	actualIndex := m.FilteredTemplates[m.SelectedTemplate]
	template := m.Templates[actualIndex]

	textPartsInterface, exists := template.Config["text_parts"]
	if !exists {
		return ""
	}

	textParts, ok := textPartsInterface.(map[string]interface{})
	if !ok || len(textParts) == 0 {
		return ""
	}

	// Convert to sorted slice
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

	// Sort by name
	for i := 0; i < len(partsList); i++ {
		for j := i + 1; j < len(partsList); j++ {
			if partsList[i].name > partsList[j].name {
				partsList[i], partsList[j] = partsList[j], partsList[i]
			}
		}
	}

	if m.SelectedTextPart >= len(partsList) {
		return ""
	}

	// Build popup content
	var content strings.Builder
	content.WriteString("Edit Text Part\n\n")

	content.WriteString("Name: ")
	content.WriteString(m.TextPartNameInput.View())
	content.WriteString("\n\n")

	content.WriteString("Path: ")
	content.WriteString(m.TextPartPathInput.View())
	content.WriteString("\n\n")

	content.WriteString("Use TAB to switch between fields")
	content.WriteString("\nPress ENTER to save, ESC to cancel")

	// Calculate popup dimensions
	popupWidth := 60
	popupHeight := 12

	// Center the popup
	leftMargin := max(0, (m.Width-popupWidth)/2)
	topMargin := max(0, (m.Height-popupHeight)/2)

	popupBox := models.ActiveBorderStyle.
		Width(popupWidth).
		Height(popupHeight).
		Padding(1).
		Render(content.String())

	// Add top margin using newlines
	centeredPopup := strings.Repeat("\n", topMargin) + popupBox

	// Add left margin using spaces
	lines := strings.Split(centeredPopup, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = strings.Repeat(" ", leftMargin) + line
		}
	}

	return strings.Join(lines, "\n")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

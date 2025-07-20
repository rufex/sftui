package ui

import (
	"encoding/json"
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
		// Show structured config fields for reconciliation_texts
		if template.Category == "reconciliation_texts" {
			details = append(details, r.renderReconciliationTextConfig(template, m, maxWidth))
		} else {
			// Fallback to JSON display for other categories
			configJSON, _ := json.MarshalIndent(template.Config, "  ", "  ")
			configLines := strings.Split(string(configJSON), "\n")
			for _, line := range configLines {
				configLine := "  " + line
				if maxWidth > 0 {
					configLine = r.TruncateText(configLine, maxWidth)
				}
				details = append(details, configLine)
			}
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

func (r *Renderer) renderReconciliationTextConfig(template models.Template, m *models.Model, maxWidth int) string {
	var lines []string

	// Only show reconciliation_type field for now
	if reconciliationType, exists := template.Config["reconciliation_type"]; exists {
		line := fmt.Sprintf("  reconciliation_type: %v", reconciliationType)
		if maxWidth > 0 {
			line = r.TruncateText(line, maxWidth)
		}

		// Highlight if this field is selected and we're in Details section
		if m.CurrentSection == models.DetailsSection && m.SelectedDetailField == 0 {
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

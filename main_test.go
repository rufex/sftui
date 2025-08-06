package main

import (
	"strings"
	"testing"

	"github.com/rufex/sftui/internal/models"
	"github.com/rufex/sftui/internal/navigation"
	"github.com/rufex/sftui/internal/template"
	"github.com/rufex/sftui/internal/ui"
)

// Tests for template manager functionality
func TestGetCategoryPrefix(t *testing.T) {
	manager := template.NewManager()

	tests := []struct {
		category string
		expected string
	}{
		{"account_templates", "AT"},
		{"export_files", "EF"},
		{"reconciliation_texts", "RT"},
		{"shared_parts", "SP"},
		{"unknown", "?"},
	}

	for _, test := range tests {
		result := manager.GetCategoryPrefix(test.category)
		if result != test.expected {
			t.Errorf("GetCategoryPrefix(%s) = %s, expected %s", test.category, result, test.expected)
		}
	}
}

func TestGetCategoryDisplayName(t *testing.T) {
	manager := template.NewManager()

	tests := []struct {
		category string
		expected string
	}{
		{"account_templates", "Account Template"},
		{"export_files", "Export File"},
		{"reconciliation_texts", "Reconciliation Text"},
		{"shared_parts", "Shared Part"},
		{"unknown", "unknown"},
	}

	for _, test := range tests {
		result := manager.GetCategoryDisplayName(test.category)
		if result != test.expected {
			t.Errorf("GetCategoryDisplayName(%s) = %s, expected %s", test.category, result, test.expected)
		}
	}
}

func TestLoadTemplates(t *testing.T) {
	manager := template.NewManager()
	templates := manager.LoadTemplates()

	if len(templates) != 12 {
		t.Errorf("Expected 12 templates from fixtures, got %d", len(templates))
	}

	// Check that all categories are represented
	categories := make(map[string]int)
	for _, template := range templates {
		categories[template.Category]++
	}

	expectedCategories := map[string]int{
		"account_templates":    3,
		"export_files":         3,
		"reconciliation_texts": 3,
		"shared_parts":         3,
	}

	for category, expectedCount := range expectedCategories {
		if categories[category] != expectedCount {
			t.Errorf("Expected %d templates in category %s, got %d", expectedCount, category, categories[category])
		}
	}
}

func TestFuzzyMatch(t *testing.T) {
	manager := template.NewManager()

	tests := []struct {
		query    string
		target   string
		expected bool
	}{
		{"", "anything", true},           // Empty query matches everything
		{"abc", "abc", true},             // Exact match
		{"abc", "axbxc", true},           // Fuzzy match with gaps
		{"abc", "aabbcc", true},          // Multiple occurrences
		{"abc", "xyz", false},            // No match
		{"abc", "acb", false},            // Wrong order
		{"Account", "account_1", true},   // Case insensitive
		{"EXPORT", "export_files", true}, // Case insensitive
	}

	for _, test := range tests {
		result := manager.FuzzyMatch(test.query, test.target)
		if result != test.expected {
			t.Errorf("FuzzyMatch(%q, %q) = %v, expected %v", test.query, test.target, result, test.expected)
		}
	}
}

func TestFilterTemplates(t *testing.T) {
	manager := template.NewManager()
	templates := manager.LoadTemplates()

	tests := []struct {
		query    string
		expected int
	}{
		{"", 12},           // Empty query shows all
		{"account", 3},     // Match category
		{"export", 3},      // Match category
		{"1", 4},           // Match name suffix
		{"nonexistent", 0}, // No matches
	}

	for _, test := range tests {
		result := manager.FilterTemplates(templates, test.query)
		if len(result) != test.expected {
			t.Errorf("FilterTemplates with query %q returned %d results, expected %d", test.query, len(result), test.expected)
		}
	}
}

// Tests for navigation handler functionality
func TestNavigationHandler(t *testing.T) {
	handler := navigation.NewHandler()
	m := &models.Model{
		CurrentSection: models.FirmSection,
	}

	// Test next section
	handler.NextSection(m)
	if m.CurrentSection != models.HostSection {
		t.Errorf("Expected section to be HostSection after NextSection, got %v", m.CurrentSection)
	}

	// Test previous section
	handler.PrevSection(m)
	if m.CurrentSection != models.FirmSection {
		t.Errorf("Expected section to be FirmSection after PrevSection, got %v", m.CurrentSection)
	}
}

func TestVerticalNavigation(t *testing.T) {
	handler := navigation.NewHandler()

	tests := []struct {
		name     string
		from     models.Section
		expected models.Section
		action   string
	}{
		{"Up from Templates to Firm", models.TemplatesSection, models.FirmSection, "up"},
		{"Up from Details to Host", models.DetailsSection, models.HostSection, "up"},
		{"Up from Output to Templates", models.OutputSection, models.TemplatesSection, "up"},
		{"Down from Firm to Templates", models.FirmSection, models.TemplatesSection, "down"},
		{"Down from Host to Details", models.HostSection, models.DetailsSection, "down"},
		{"Down from Templates to Output", models.TemplatesSection, models.OutputSection, "down"},
		{"Down from Details to Output", models.DetailsSection, models.OutputSection, "down"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := &models.Model{CurrentSection: test.from}

			if test.action == "up" {
				handler.HandleVerticalUp(m)
			} else {
				handler.HandleVerticalDown(m)
			}

			if m.CurrentSection != test.expected {
				t.Errorf("Expected section %v, got %v", test.expected, m.CurrentSection)
			}
		})
	}
}

func TestTemplateNavigation(t *testing.T) {
	handler := navigation.NewHandler()
	m := &models.Model{
		CurrentSection:    models.TemplatesSection,
		FilteredTemplates: []int{0, 1, 2, 3, 4},
		SelectedTemplate:  0,
		TemplatesOffset:   0,
		Height:            24,
	}

	// Test down navigation
	handler.HandleTemplateNavigation(m, "down")
	if m.SelectedTemplate != 1 {
		t.Errorf("Expected SelectedTemplate to be 1 after down navigation, got %d", m.SelectedTemplate)
	}

	// Test up navigation
	handler.HandleTemplateNavigation(m, "up")
	if m.SelectedTemplate != 0 {
		t.Errorf("Expected SelectedTemplate to be 0 after up navigation, got %d", m.SelectedTemplate)
	}

	// Test wrapping
	handler.HandleTemplateNavigation(m, "up")
	if m.SelectedTemplate != 4 {
		t.Errorf("Expected SelectedTemplate to wrap to 4 after up navigation from 0, got %d", m.SelectedTemplate)
	}
}

// Tests for UI renderer functionality
func TestUIRenderer(t *testing.T) {
	renderer := ui.NewRenderer()
	m := &models.Model{
		Templates: []models.Template{
			{Name: "test1", Category: "account_templates"},
			{Name: "test2", Category: "export_files"},
		},
		FilteredTemplates: []int{0, 1},
		SelectedTemplate:  0,
		SelectedTemplates: make(map[int]bool),
	}

	// Test templates view
	view := renderer.TemplatesView(m)
	if !strings.Contains(view, "test1") {
		t.Errorf("Templates view should contain 'test1'")
	}
	if !strings.Contains(view, "[AT]") {
		t.Errorf("Templates view should contain category prefix [AT]")
	}

	// Test details view
	detailsView := renderer.DetailsView(m)
	if !strings.Contains(detailsView, "Name: test1") {
		t.Errorf("Details view should contain 'Name: test1'")
	}
}

func TestTruncateText(t *testing.T) {
	renderer := ui.NewRenderer()

	tests := []struct {
		text     string
		maxWidth int
		expected string
	}{
		{"short", 20, "short"},
		{"this is a very long text that should be truncated", 20, "this is a ver..."},
		{"test", 10, "test"},
		{"", 10, ""},
	}

	for _, test := range tests {
		result := renderer.TruncateText(test.text, test.maxWidth)
		if result != test.expected {
			t.Errorf("TruncateText(%q, %d) = %q, expected %q", test.text, test.maxWidth, result, test.expected)
		}
	}
}

func TestActionPopup(t *testing.T) {
	renderer := ui.NewRenderer()
	m := &models.Model{
		SelectedTemplates: map[int]bool{0: true, 1: true},
		SelectedAction:    0,
		Width:             80,
		Height:            24,
	}

	popup := renderer.ActionPopupView(m)
	if !strings.Contains(popup, "2 templates selected") {
		t.Errorf("Action popup should show '2 templates selected'")
	}
	if !strings.Contains(popup, "> create") {
		t.Errorf("Action popup should highlight 'create' action")
	}
}

func TestHelpView(t *testing.T) {
	renderer := ui.NewRenderer()
	m := &models.Model{
		Width:  80,
		Height: 24,
	}

	help := renderer.HelpView(m)
	if !strings.Contains(help, "Key Bindings") {
		t.Errorf("Help view should contain 'Key Bindings'")
	}
	if !strings.Contains(help, "TAB") {
		t.Errorf("Help view should contain 'TAB' key binding")
	}
}

// Tests for config manager functionality
func TestConfigManager(t *testing.T) {
	configManager := template.NewConfigManager()
	firm, host, output := configManager.LoadSilverfinConfig()

	// Since we're using fixtures, we should get some values
	if firm == "" {
		t.Errorf("Expected firm to be set from fixtures")
	}
	if host == "" {
		t.Errorf("Expected host to be set from fixtures")
	}
	if output == "Error getting home directory" || output == "Error parsing Silverfin config" {
		t.Errorf("Expected successful config loading, got error: %s", output)
	}
}

func TestLoadFirmOptions(t *testing.T) {
	configManager := template.NewConfigManager()
	firmOptions, err := configManager.LoadFirmOptions()

	if err != nil {
		t.Errorf("Expected no error loading firm options, got %v", err)
		return
	}

	if len(firmOptions) == 0 {
		t.Logf("No firm options loaded - this might be expected if config is not found")
		t.Skip("No firm options available - skipping test")
	}

	// Check that we have both firms and partners
	hasFirm := false
	hasPartner := false

	for _, option := range firmOptions {
		if option.Type == "firm" {
			hasFirm = true
		}
		if option.Type == "partner" {
			hasPartner = true
		}

		if option.ID == "" || option.Name == "" {
			t.Errorf("Expected firm option to have ID and Name, got ID='%s', Name='%s'", option.ID, option.Name)
		}
	}

	if !hasFirm {
		t.Errorf("Expected at least one firm option")
	}

	if !hasPartner {
		t.Errorf("Expected at least one partner option")
	}
}

func TestFirmPopupView(t *testing.T) {
	renderer := ui.NewRenderer()
	m := &models.Model{
		Width:  80,
		Height: 24,
	}

	// Test with no firm options
	m.FirmOptions = []models.FirmOption{}
	view := renderer.FirmPopupView(m)

	if !strings.Contains(view, "No firms or partners available") {
		t.Errorf("Expected message about no firms available")
	}

	// Test with firm options
	m.FirmOptions = []models.FirmOption{
		{ID: "1001", Name: "Test Firm", Type: "firm"},
		{ID: "25", Name: "Test Partner", Type: "partner"},
	}
	m.SelectedFirm = 0

	view = renderer.FirmPopupView(m)

	if !strings.Contains(view, "Select Default Firm/Partner") {
		t.Errorf("Expected popup title")
	}

	if !strings.Contains(view, "Test Firm (1001)") {
		t.Errorf("Expected firm name with ID in popup")
	}

	if !strings.Contains(view, "Test Partner (25)") {
		t.Errorf("Expected partner name with ID in popup")
	}

	if !strings.Contains(view, "Use ↑/↓ or k/j to navigate") {
		t.Errorf("Expected navigation instructions")
	}

	if !strings.Contains(view, "Press ENTER to select, ESC to cancel") {
		t.Errorf("Expected action instructions")
	}
}

func TestSetHost(t *testing.T) {
	configManager := template.NewConfigManager()

	// Test setting a new host
	newHost := "https://new-test-host.com"
	err := configManager.SetHost(newHost)

	if err != nil {
		t.Errorf("Expected no error setting host, got %v", err)
	}

	// Verify the host was set by loading config again
	_, host, _ := configManager.LoadSilverfinConfig()

	if host != newHost {
		t.Errorf("Expected host to be '%s', got '%s'", newHost, host)
	}
}

func TestHostPopupViewWithTextInput(t *testing.T) {
	renderer := ui.NewRenderer()
	m := &models.Model{
		Width:  80,
		Height: 24,
	}
	m.HostTextInput.SetValue("https://test.com")

	view := renderer.HostPopupView(m)

	// Check that textinput is rendered
	if !strings.Contains(view, "Host: ") {
		t.Errorf("Expected view to contain 'Host: ' label")
	}

	// Check that the textinput component is used (it will have its own styling)
	if !strings.Contains(view, "https://test.com") {
		t.Errorf("Expected view to contain the host value 'https://test.com'")
	}

	// Check updated instructions for textinput
	if !strings.Contains(view, "Use ←/→ to move cursor, Ctrl+A/E for start/end") {
		t.Errorf("Expected updated instructions for textinput controls")
	}

	if !strings.Contains(view, "Press ENTER to save, ESC to cancel") {
		t.Errorf("Expected standard save/cancel instructions")
	}
}

func TestDetailsNavigation(t *testing.T) {
	manager := template.NewManager()
	templates := manager.LoadTemplates()

	// Find a reconciliation_text template for testing
	reconciliationIndex := -1
	for i, template := range templates {
		if template.Category == "reconciliation_texts" {
			reconciliationIndex = i
			break
		}
	}

	if reconciliationIndex == -1 {
		t.Skip("No reconciliation_texts templates available for testing")
	}

	// Create a model with the reconciliation template
	m := &models.Model{
		Templates:         templates,
		CurrentSection:    models.DetailsSection,
		SelectedTemplate:  0,
		FilteredTemplates: []int{reconciliationIndex},
	}

	// Test initial state
	if m.SelectedDetailField != 0 {
		t.Errorf("Expected initial SelectedDetailField to be 0, got %d", m.SelectedDetailField)
	}

	// Test that navigation works - now includes text parts
	navHandler := navigation.NewHandler()
	template := m.Templates[reconciliationIndex]

	// Count total fields (config fields + text parts)
	totalFields := 1 // reconciliation_type
	if textPartsInterface, exists := template.Config["text_parts"]; exists {
		if textParts, ok := textPartsInterface.(map[string]interface{}); ok {
			totalFields += len(textParts)
		}
	}

	navHandler.HandleDetailsNavigation(m, "down")
	// Should move to next field
	if m.SelectedDetailField != 1 {
		t.Errorf("Expected SelectedDetailField to be 1 after down navigation, got %d", m.SelectedDetailField)
	}

	// Navigate back up
	navHandler.HandleDetailsNavigation(m, "up")
	if m.SelectedDetailField != 0 {
		t.Errorf("Expected SelectedDetailField to be 0 after up navigation, got %d", m.SelectedDetailField)
	}
}

func TestDetailsNavigationWithoutReconciliationType(t *testing.T) {
	// Create a template without reconciliation_type
	templates := []models.Template{
		{
			Name:     "Test Template",
			Category: "account_templates",
			Config:   map[string]interface{}{},
		},
	}

	m := &models.Model{
		Templates:         templates,
		FilteredTemplates: []int{0},
		CurrentSection:    models.DetailsSection,
	}

	navHandler := navigation.NewHandler()

	// Should not navigate if no config fields available
	initialField := m.SelectedDetailField
	navHandler.HandleDetailsNavigation(m, "down")
	if m.SelectedDetailField != initialField {
		t.Errorf("Expected SelectedDetailField to remain unchanged for template without config fields")
	}
}

func TestDetailsHighlighting(t *testing.T) {
	manager := template.NewManager()
	templates := manager.LoadTemplates()

	// Find a reconciliation_text template
	reconciliationIndex := -1
	for i, template := range templates {
		if template.Category == "reconciliation_texts" {
			reconciliationIndex = i
			break
		}
	}

	if reconciliationIndex == -1 {
		t.Skip("No reconciliation_texts templates available for testing")
	}

	// Setup for testing highlighting
	m := &models.Model{
		Templates:           templates,
		CurrentSection:      models.DetailsSection,
		FilteredTemplates:   []int{reconciliationIndex},
		SelectedTemplate:    0,
		SelectedDetailField: 0,
	}

	// Render details view
	renderer := ui.NewRenderer()
	view := renderer.DetailsViewWithHeightAndWidth(m, 10, 50)

	// Should contain reconciliation_type field
	if !strings.Contains(view, "reconciliation_type") {
		t.Errorf("Expected details view to contain reconciliation_type field")
	}

	// Test that highlighting is not applied when not in Details section
	m.CurrentSection = models.TemplatesSection
	viewNoHighlight := renderer.DetailsViewWithHeightAndWidth(m, 10, 50)

	// Both views should contain the field, but highlighting should be different
	if !strings.Contains(viewNoHighlight, "reconciliation_type") {
		t.Errorf("Expected details view to contain reconciliation_type field even when not active")
	}
}

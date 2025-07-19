package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rufex/sftui/internal/models"
	"github.com/rufex/sftui/internal/navigation"
	"github.com/rufex/sftui/internal/template"
	"github.com/rufex/sftui/internal/ui"
)

func TestGetCategoryPrefix(t *testing.T) {
	manager := template.NewManager()

	tests := []struct {
		category string
		expected string
	}{
		{"account_templates", "A"},
		{"export_files", "E"},
		{"reconciliation_texts", "R"},
		{"shared_parts", "S"},
		{"unknown", "?"},
	}

	for _, test := range tests {
		result := manager.GetCategoryPrefix(test.category)
		if result != test.expected {
			t.Errorf("GetCategoryPrefix(%s) = %s, expected %s", test.category, result, test.expected)
		}
	}
}

func TestInitialModel(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	if m.CurrentSection != models.TemplatesSection {
		t.Errorf("Expected initial section to be TemplatesSection, got %v", m.CurrentSection)
	}

	if m.SelectedTemplate != 0 {
		t.Errorf("Expected initial selectedTemplate to be 0, got %d", m.SelectedTemplate)
	}

	if m.ShowHelp != false {
		t.Errorf("Expected ShowHelp to be false initially")
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
	if !strings.Contains(view, "[A]") {
		t.Errorf("Templates view should contain category prefix [A]")
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

func TestTemplateSelection(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	if len(m.SelectedTemplates) != 0 {
		t.Errorf("Expected no templates selected initially, got %d", len(m.SelectedTemplates))
	}

	// Simulate template selection
	if len(m.FilteredTemplates) > 0 {
		actualIndex := m.FilteredTemplates[0]
		m.SelectedTemplates[actualIndex] = true

		if len(m.SelectedTemplates) != 1 {
			t.Errorf("Expected 1 template selected, got %d", len(m.SelectedTemplates))
		}

		if !m.SelectedTemplates[actualIndex] {
			t.Errorf("Expected template at index %d to be selected", actualIndex)
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

// Test firm selection functionality
func TestFirmPopupTrigger(t *testing.T) {
	app := newApp()
	m := app.initialModel()
	
	// Set current section to Firm
	m.CurrentSection = models.FirmSection
	
	// Simulate Enter key press
	key := tea.KeyMsg{Type: tea.KeyEnter}
	
	app.model = m
	_, _ = app.Update(key)
	
	if !app.model.ShowFirmPopup {
		t.Errorf("Expected ShowFirmPopup to be true after Enter in Firm section")
	}
	
	if app.model.SelectedFirm != 0 {
		t.Errorf("Expected SelectedFirm to be reset to 0, got %d", app.model.SelectedFirm)
	}
	
	if app.model.Output != "Select a firm or partner" {
		t.Errorf("Expected output message 'Select a firm or partner', got %s", app.model.Output)
	}
}

func TestFirmPopupNavigation(t *testing.T) {
	app := newApp()
	m := app.initialModel()
	
	// Ensure we have firm options loaded
	if len(m.FirmOptions) == 0 {
		t.Skip("No firm options available for testing")
	}
	
	// Set popup state
	m.ShowFirmPopup = true
	m.SelectedFirm = 0
	
	tests := []struct {
		keyType  tea.KeyType
		keyStr   string
		expected int
	}{
		{tea.KeyDown, "down", 1},
		{tea.KeyRunes, "j", 2 % len(m.FirmOptions)},
		{tea.KeyUp, "up", (2 - 1 + len(m.FirmOptions)) % len(m.FirmOptions)},
		{tea.KeyRunes, "k", (1 - 1 + len(m.FirmOptions)) % len(m.FirmOptions)},
	}
	
	app.model = m
	for _, test := range tests {
		var key tea.KeyMsg
		if test.keyType == tea.KeyRunes {
			key = tea.KeyMsg{Type: test.keyType, Runes: []rune{rune(test.keyStr[0])}}
		} else {
			key = tea.KeyMsg{Type: test.keyType}
		}
		_, _ = app.Update(key)
		
		if app.model.SelectedFirm != test.expected {
			t.Errorf("After key %s, expected SelectedFirm %d, got %d", test.keyStr, test.expected, app.model.SelectedFirm)
		}
	}
}

func TestFirmPopupEscape(t *testing.T) {
	app := newApp()
	m := app.initialModel()
	
	// Set popup state
	m.ShowFirmPopup = true
	m.SelectedFirm = 2
	
	// Simulate Escape key
	key := tea.KeyMsg{Type: tea.KeyEscape}
	
	app.model = m
	_, _ = app.Update(key)
	
	if app.model.ShowFirmPopup {
		t.Errorf("Expected ShowFirmPopup to be false after Escape")
	}
	
	if app.model.SelectedFirm != 0 {
		t.Errorf("Expected SelectedFirm to be reset to 0, got %d", app.model.SelectedFirm)
	}
	
	if app.model.Output != "Firm selection cancelled" {
		t.Errorf("Expected output 'Firm selection cancelled', got %s", app.model.Output)
	}
}

func TestFirmPopupSelection(t *testing.T) {
	app := newApp()
	m := app.initialModel()
	
	// Ensure we have firm options loaded
	if len(m.FirmOptions) == 0 {
		t.Skip("No firm options available for testing")
	}
	
	// Set popup state
	m.ShowFirmPopup = true
	m.SelectedFirm = 0
	selectedOption := m.FirmOptions[0]
	
	// Simulate Enter key for selection
	key := tea.KeyMsg{Type: tea.KeyEnter}
	
	app.model = m
	_, _ = app.Update(key)
	
	if app.model.ShowFirmPopup {
		t.Errorf("Expected ShowFirmPopup to be false after selection")
	}
	
	if app.model.SelectedFirm != 0 {
		t.Errorf("Expected SelectedFirm to be reset to 0, got %d", app.model.SelectedFirm)
	}
	
	if !strings.Contains(app.model.Firm, selectedOption.Name) {
		t.Errorf("Expected firm to contain %s, got %s", selectedOption.Name, app.model.Firm)
	}
	
	if !strings.Contains(app.model.Output, "Default firm set to") {
		t.Errorf("Expected success message in output, got %s", app.model.Output)
	}
}

func TestLoadFirmOptions(t *testing.T) {
	configManager := template.NewConfigManager()
	firmOptions, err := configManager.LoadFirmOptions()
	
	if err != nil {
		t.Errorf("Expected no error loading firm options, got %v", err)
	}
	
	if len(firmOptions) == 0 {
		t.Errorf("Expected at least some firm options from fixtures")
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
	app := newApp()
	m := app.initialModel()
	renderer := ui.NewRenderer()
	
	// Set some basic dimensions
	m.Width = 80
	m.Height = 24
	
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

func TestFirmPopupIntegration(t *testing.T) {
	app := newApp()
	m := app.initialModel()
	
	key := tea.KeyMsg{Type: tea.KeyEnter}
	
	// Test that Enter in firm popup selects a firm (if firm options are available)
	if len(m.FirmOptions) > 0 {
		m.CurrentSection = models.FirmSection
		m.ShowFirmPopup = true
		m.SelectedFirm = 0
		
		app.model = m
		_, _ = app.Update(key)
		
		// Should close popup after selection
		if app.model.ShowFirmPopup {
			t.Errorf("Expected firm popup to close after selection")
		}
		
		// Should show success message
		if !strings.Contains(app.model.Output, "Default firm set to") {
			t.Errorf("Expected success message, got: %s", app.model.Output)
		}
	}
	
	// Test that Enter in other sections doesn't trigger firm popup
	m.ShowFirmPopup = false
	m.CurrentSection = models.TemplatesSection
	
	app.model = m
	_, _ = app.Update(key)
	
	if app.model.ShowFirmPopup {
		t.Errorf("Expected firm popup not to trigger from Templates section")
	}
}

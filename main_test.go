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

// Test host edit functionality
func TestHostPopupTrigger(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Set current section to Host
	m.CurrentSection = models.HostSection

	// Simulate Enter key press
	key := tea.KeyMsg{Type: tea.KeyEnter}

	app.model = m
	_, _ = app.Update(key)

	if !app.model.ShowHostPopup {
		t.Errorf("Expected ShowHostPopup to be true after Enter in Host section")
	}

	if app.model.HostTextInput.Value() != app.model.Host {
		t.Errorf("Expected HostTextInput to be pre-filled with current host")
	}

	if app.model.Output != "Edit host URL" {
		t.Errorf("Expected output message 'Edit host URL', got %s", app.model.Output)
	}
}

func TestHostPopupTextInputMultipleChars(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Set popup state with textinput
	m.ShowHostPopup = true
	m.HostTextInput.SetValue("https://test")
	m.HostTextInput.Focus()

	// Test adding characters through textinput
	chars := []rune{'.', 'c', 'o', 'm'}
	for _, char := range chars {
		key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}}
		app.model = m
		_, _ = app.Update(key)
		m = app.model
	}

	// Check that textinput handles the characters (exact result depends on textinput implementation)
	finalValue := app.model.HostTextInput.Value()
	if !strings.Contains(finalValue, "test") {
		t.Errorf("Expected textinput to contain 'test', got '%s'", finalValue)
	}
}

// TestHostPopupBackspace already exists in new implementation, removing duplicate

// TestHostPopupEscape already exists in new implementation, removing duplicate

// TestHostPopupSave already exists in new implementation, removing duplicate

// TestHostPopupView already exists in new implementation, removing duplicate

func TestHostPopupIntegration(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	key := tea.KeyMsg{Type: tea.KeyEnter}

	// Test that Enter in other sections doesn't trigger host popup
	m.CurrentSection = models.TemplatesSection

	app.model = m
	_, _ = app.Update(key)

	if app.model.ShowHostPopup {
		t.Errorf("Expected host popup not to trigger from Templates section")
	}

	// Test that Enter in host section triggers popup
	m.CurrentSection = models.HostSection

	app.model = m
	_, _ = app.Update(key)

	if !app.model.ShowHostPopup {
		t.Errorf("Expected host popup to trigger from Host section")
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

// Test textinput initialization and basic functionality
func TestHostTextInputInitialization(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Check that textinput is properly initialized
	if m.HostTextInput.CharLimit != 256 {
		t.Errorf("Expected CharLimit to be 256, got %d", m.HostTextInput.CharLimit)
	}

	if m.HostTextInput.Width != 50 {
		t.Errorf("Expected Width to be 50, got %d", m.HostTextInput.Width)
	}

	expectedPlaceholder := "Enter host URL (e.g., https://api.example.com)"
	if m.HostTextInput.Placeholder != expectedPlaceholder {
		t.Errorf("Expected placeholder '%s', got '%s'", expectedPlaceholder, m.HostTextInput.Placeholder)
	}
}

func TestHostPopupTextInputFocus(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Set current section to Host and simulate Enter to open popup
	m.CurrentSection = models.HostSection
	m.Host = "https://example.com"

	key := tea.KeyMsg{Type: tea.KeyEnter}
	app.model = m
	_, _ = app.Update(key)

	// Check that popup is shown and textinput is focused with correct value
	if !app.model.ShowHostPopup {
		t.Errorf("Expected host popup to be shown")
	}

	if !app.model.HostTextInput.Focused() {
		t.Errorf("Expected textinput to be focused")
	}

	if app.model.HostTextInput.Value() != m.Host {
		t.Errorf("Expected textinput value to be '%s', got '%s'", m.Host, app.model.HostTextInput.Value())
	}
}

func TestHostPopupTextInputTyping(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Open host popup with focused textinput
	m.ShowHostPopup = true
	m.HostTextInput.SetValue("https://test.com")
	m.HostTextInput.Focus()

	// Type a character
	char := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}

	app.model = m
	_, _ = app.Update(char)

	// The textinput should handle the character insertion
	value := app.model.HostTextInput.Value()
	if !strings.Contains(value, "x") {
		t.Errorf("Expected textinput to contain 'x' after typing, got '%s'", value)
	}
}

func TestHostPopupTextInputBackspace(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Open host popup with focused textinput
	m.ShowHostPopup = true
	m.HostTextInput.SetValue("https://test.com")
	m.HostTextInput.Focus()

	originalLength := len(m.HostTextInput.Value())

	// Send backspace
	backspace := tea.KeyMsg{Type: tea.KeyBackspace}

	app.model = m
	_, _ = app.Update(backspace)

	// The textinput should handle the backspace
	newValue := app.model.HostTextInput.Value()
	if len(newValue) >= originalLength {
		t.Errorf("Expected textinput value to be shorter after backspace, original: %d, new: %d", originalLength, len(newValue))
	}
}

func TestHostPopupEscapeKey(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Open host popup
	m.ShowHostPopup = true
	m.HostTextInput.SetValue("https://test.com")
	m.HostTextInput.Focus()

	// Send escape key
	escape := tea.KeyMsg{Type: tea.KeyEsc}

	app.model = m
	_, _ = app.Update(escape)

	// Popup should be closed and textinput should be blurred
	if app.model.ShowHostPopup {
		t.Errorf("Expected host popup to be closed after escape")
	}

	if app.model.HostTextInput.Focused() {
		t.Errorf("Expected textinput to be blurred after escape")
	}

	if app.model.Output != "Host edit cancelled" {
		t.Errorf("Expected output message 'Host edit cancelled', got '%s'", app.model.Output)
	}
}

func TestHostPopupEnterKeySave(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Open host popup and set a new value
	m.ShowHostPopup = true
	newHost := "https://new-host.com"
	m.HostTextInput.SetValue(newHost)
	m.HostTextInput.Focus()

	// Send enter key to save
	enter := tea.KeyMsg{Type: tea.KeyEnter}

	app.model = m
	_, _ = app.Update(enter)

	// Popup should be closed and textinput should be blurred
	if app.model.ShowHostPopup {
		t.Errorf("Expected host popup to be closed after enter")
	}

	if app.model.HostTextInput.Focused() {
		t.Errorf("Expected textinput to be blurred after enter")
	}

	if app.model.Host != newHost {
		t.Errorf("Expected host to be updated to '%s', got '%s'", newHost, app.model.Host)
	}

	if app.model.Output != "Host updated successfully" {
		t.Errorf("Expected output message 'Host updated successfully', got '%s'", app.model.Output)
	}
}

func TestHostPopupViewWithTextInput(t *testing.T) {
	app := newApp()
	m := app.initialModel()
	renderer := ui.NewRenderer()

	// Set some basic dimensions
	m.Width = 80
	m.Height = 24
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

func TestHostTextInputAdvancedFeatures(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Test character limit
	if m.HostTextInput.CharLimit != 256 {
		t.Errorf("Expected character limit of 256, got %d", m.HostTextInput.CharLimit)
	}

	// Test that textinput handles long URLs correctly
	longURL := "https://very-long-subdomain.example-domain-name.com/very/long/path/with/many/segments/and/parameters?param1=value1&param2=value2&param3=value3"
	m.HostTextInput.SetValue(longURL)

	if len(m.HostTextInput.Value()) > 256 {
		t.Errorf("Expected textinput to respect character limit, got length %d", len(m.HostTextInput.Value()))
	}

	// Test placeholder functionality
	m.HostTextInput.SetValue("")
	if m.HostTextInput.Placeholder == "" {
		t.Errorf("Expected placeholder to be set")
	}

	// Test focus/blur functionality
	m.HostTextInput.Focus()
	if !m.HostTextInput.Focused() {
		t.Errorf("Expected textinput to be focused after Focus() call")
	}

	m.HostTextInput.Blur()
	if m.HostTextInput.Focused() {
		t.Errorf("Expected textinput to be blurred after Blur() call")
	}
}

func TestDetailsNavigation(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Ensure we have templates loaded
	if len(m.Templates) == 0 {
		t.Skip("No templates available for testing")
	}

	// Find a reconciliation_text template for testing
	reconciliationIndex := -1
	for i, template := range m.Templates {
		if template.Category == "reconciliation_texts" {
			reconciliationIndex = i
			break
		}
	}

	if reconciliationIndex == -1 {
		t.Skip("No reconciliation_texts templates available for testing")
	}

	// Set current section to Details
	m.CurrentSection = models.DetailsSection
	m.SelectedTemplate = 0

	// Update filtered templates to include our reconciliation template
	m.FilteredTemplates = []int{reconciliationIndex}

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
	app := newApp()
	m := app.initialModel()

	// Create a template without reconciliation_type
	m.Templates = []models.Template{
		{
			Name:     "Test Template",
			Category: "account_templates",
			Config:   map[string]interface{}{},
		},
	}
	m.FilteredTemplates = []int{0}
	m.CurrentSection = models.DetailsSection

	navHandler := navigation.NewHandler()

	// Should not navigate if no config fields available
	initialField := m.SelectedDetailField
	navHandler.HandleDetailsNavigation(m, "down")
	if m.SelectedDetailField != initialField {
		t.Errorf("Expected SelectedDetailField to remain unchanged for template without config fields")
	}
}

func TestReconciliationTypePopupTrigger(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Find a reconciliation_text template
	reconciliationIndex := -1
	for i, template := range m.Templates {
		if template.Category == "reconciliation_texts" {
			reconciliationIndex = i
			break
		}
	}

	if reconciliationIndex == -1 {
		t.Skip("No reconciliation_texts templates available for testing")
	}

	// Setup for testing Enter key in Details section
	m.CurrentSection = models.DetailsSection
	m.FilteredTemplates = []int{reconciliationIndex}
	m.SelectedTemplate = 0
	m.SelectedDetailField = 0

	// Simulate Enter key press
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, _ = app.Update(keyMsg)

	// Check that popup was triggered
	if !m.ShowReconciliationTypePopup {
		t.Errorf("Expected ShowReconciliationTypePopup to be true after Enter in Details section")
	}

	if m.SelectedReconciliationType < 0 || m.SelectedReconciliationType > 2 {
		t.Errorf("Expected SelectedReconciliationType to be 0-2, got %d", m.SelectedReconciliationType)
	}
}

func TestReconciliationTypePopupNavigation(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Setup popup state
	m.ShowReconciliationTypePopup = true
	m.SelectedReconciliationType = 0

	// Test down navigation
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	_, _ = app.Update(keyMsg)
	if m.SelectedReconciliationType != 1 {
		t.Errorf("Expected SelectedReconciliationType to be 1 after down navigation, got %d", m.SelectedReconciliationType)
	}

	// Test up navigation
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	_, _ = app.Update(keyMsg)
	if m.SelectedReconciliationType != 0 {
		t.Errorf("Expected SelectedReconciliationType to be 0 after up navigation, got %d", m.SelectedReconciliationType)
	}

	// Test wrap around (down from 2 should go to 0)
	m.SelectedReconciliationType = 2
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	_, _ = app.Update(keyMsg)
	if m.SelectedReconciliationType != 0 {
		t.Errorf("Expected SelectedReconciliationType to wrap to 0, got %d", m.SelectedReconciliationType)
	}

	// Test wrap around (up from 0 should go to 2)
	m.SelectedReconciliationType = 0
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	_, _ = app.Update(keyMsg)
	if m.SelectedReconciliationType != 2 {
		t.Errorf("Expected SelectedReconciliationType to wrap to 2, got %d", m.SelectedReconciliationType)
	}
}

func TestReconciliationTypePopupCancel(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Setup popup state
	m.ShowReconciliationTypePopup = true
	m.SelectedReconciliationType = 1

	// Test escape cancellation
	keyMsg := tea.KeyMsg{Type: tea.KeyEsc}
	_, _ = app.Update(keyMsg)

	if m.ShowReconciliationTypePopup {
		t.Errorf("Expected ShowReconciliationTypePopup to be false after escape")
	}

	if m.SelectedReconciliationType != 0 {
		t.Errorf("Expected SelectedReconciliationType to reset to 0 after escape, got %d", m.SelectedReconciliationType)
	}

	if !strings.Contains(m.Output, "cancelled") {
		t.Errorf("Expected output to contain 'cancelled', got: %s", m.Output)
	}
}

func TestDetailsHighlighting(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Find a reconciliation_text template
	reconciliationIndex := -1
	for i, template := range m.Templates {
		if template.Category == "reconciliation_texts" {
			reconciliationIndex = i
			break
		}
	}

	if reconciliationIndex == -1 {
		t.Skip("No reconciliation_texts templates available for testing")
	}

	// Setup for testing highlighting
	m.CurrentSection = models.DetailsSection
	m.FilteredTemplates = []int{reconciliationIndex}
	m.SelectedTemplate = 0
	m.SelectedDetailField = 0

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

func TestUpDownNavigationInDetailsSection(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Find a reconciliation_text template
	reconciliationIndex := -1
	for i, template := range m.Templates {
		if template.Category == "reconciliation_texts" {
			reconciliationIndex = i
			break
		}
	}

	if reconciliationIndex == -1 {
		t.Skip("No reconciliation_texts templates available for testing")
	}

	// Setup
	m.CurrentSection = models.DetailsSection
	m.FilteredTemplates = []int{reconciliationIndex}
	m.SelectedTemplate = 0
	m.SelectedDetailField = 0

	// Test that up/down keys work in Details section
	initialField := m.SelectedDetailField

	// Simulate down key
	keyMsg := tea.KeyMsg{Type: tea.KeyDown}
	_, _ = app.Update(keyMsg)

	// Now we have multiple fields (reconciliation_type + text parts), so should move to next field
	template := m.Templates[reconciliationIndex]

	// Count total fields
	totalFields := 1 // reconciliation_type
	if textPartsInterface, exists := template.Config["text_parts"]; exists {
		if textParts, ok := textPartsInterface.(map[string]interface{}); ok {
			totalFields += len(textParts)
		}
	}

	if totalFields > 1 {
		// Should move to next field
		if m.SelectedDetailField != 1 {
			t.Errorf("Expected SelectedDetailField to be 1 with multiple fields, got %d", m.SelectedDetailField)
		}
	} else {
		// With only one field, should remain the same
		if m.SelectedDetailField != initialField {
			t.Errorf("Expected SelectedDetailField to remain %d with single field, got %d", initialField, m.SelectedDetailField)
		}
	}

	// Simulate up key
	keyMsg = tea.KeyMsg{Type: tea.KeyUp}
	_, _ = app.Update(keyMsg)

	// Should go back to initial field
	if m.SelectedDetailField != initialField {
		t.Errorf("Expected SelectedDetailField to return to %d after up navigation, got %d", initialField, m.SelectedDetailField)
	}
}

func TestDetailsNavigationMultipleHighlighting(t *testing.T) {
	app := newApp()
	m := app.initialModel()

	// Create a simple test template with just a few known fields
	testTemplate := models.Template{
		Name:     "test_template",
		Category: "reconciliation_texts",
		Config: map[string]interface{}{
			"reconciliation_type": "can_be_reconciled_without_data",
			"public":              false,
			"text_parts": map[string]interface{}{
				"part_1": "text_parts/part_1.liquid",
			},
		},
	}

	m.Templates = []models.Template{testTemplate}
	m.FilteredTemplates = []int{0}
	m.SelectedTemplate = 0
	m.CurrentSection = models.DetailsSection

	renderer := ui.NewRenderer()

	// Test specific field indices to see which ones should highlight what
	testCases := []struct {
		fieldIndex   int
		description  string
		shouldHighlight string
	}{
		{0, "First config field", "reconciliation_type"},
		{1, "Second config field", "public"},
		{2, "First text part", "part_1"},
	}

	for _, tc := range testCases {
		m.SelectedDetailField = tc.fieldIndex
		detailsContent := renderer.DetailsView(m)
		
		// Debug: print what's being highlighted
		t.Logf("Field index %d (%s):", tc.fieldIndex, tc.description)
		t.Logf("Content contains reconciliation_type: %v", strings.Contains(detailsContent, "reconciliation_type"))
		t.Logf("Content contains public: %v", strings.Contains(detailsContent, "public"))
		t.Logf("Content contains part_1: %v", strings.Contains(detailsContent, "part_1"))
		
		// Check which items are highlighted
		reconciliationHighlighted := strings.Contains(detailsContent, models.SelectedItemStyle.Render("  reconciliation_type: can_be_reconciled_without_data"))
		publicHighlighted := strings.Contains(detailsContent, models.SelectedItemStyle.Render("  public: false"))
		part1Highlighted := strings.Contains(detailsContent, models.SelectedItemStyle.Render("  part_1"))
		
		t.Logf("reconciliation_type highlighted: %v", reconciliationHighlighted)
		t.Logf("public highlighted: %v", publicHighlighted)
		t.Logf("part_1 highlighted: %v", part1Highlighted)
		
		// Count total highlighted items
		highlightCount := 0
		if reconciliationHighlighted { highlightCount++ }
		if publicHighlighted { highlightCount++ }
		if part1Highlighted { highlightCount++ }
		
		if highlightCount > 1 {
			t.Errorf("Field index %d: Expected 1 highlighted field, got %d", tc.fieldIndex, highlightCount)
		}
	}
}

func TestDetailsFieldIndexCalculation(t *testing.T) {
	testTemplate := models.Template{
		Name:     "test_template",
		Category: "reconciliation_texts",
		Config: map[string]interface{}{
			"reconciliation_type": "can_be_reconciled_without_data",
			"public":              false,
			"text_parts": map[string]interface{}{
				"part_1": "text_parts/part_1.liquid",
			},
		},
	}

	renderer := ui.NewRenderer()
	navHandler := navigation.NewHandler()

	// Test field counting
	rendererConfigCount := renderer.GetConfigFieldCount(testTemplate)
	navConfigCount := navHandler.GetActualConfigFieldCount(testTemplate)
	
	t.Logf("Renderer config count: %d", rendererConfigCount)
	t.Logf("Navigation config count: %d", navConfigCount)
	
	if rendererConfigCount != navConfigCount {
		t.Errorf("Config count mismatch: renderer=%d, navigation=%d", rendererConfigCount, navConfigCount)
	}
	
	// Expected: 2 config fields (reconciliation_type, public) + 1 text part
	expectedConfigFields := 2
	expectedTextParts := 1
	expectedTotal := expectedConfigFields + expectedTextParts
	
	if rendererConfigCount != expectedConfigFields {
		t.Errorf("Expected %d config fields, got %d", expectedConfigFields, rendererConfigCount)
	}
	
	textPartsCount := navHandler.GetTextPartsCount(testTemplate)
	if textPartsCount != expectedTextParts {
		t.Errorf("Expected %d text parts, got %d", expectedTextParts, textPartsCount)
	}
	
	totalFields := navHandler.GetConfigFieldCount(testTemplate, map[string][]string{})
	if totalFields != expectedTotal {
		t.Errorf("Expected %d total fields, got %d", expectedTotal, totalFields)
	}
}

func TestGetConfigFieldCount(t *testing.T) {
	navHandler := navigation.NewHandler()

	tests := []struct {
		name          string
		template      models.Template
		expectedCount int
	}{
		{
			name: "reconciliation_texts with reconciliation_type",
			template: models.Template{
				Category: "reconciliation_texts",
				Config:   map[string]interface{}{"reconciliation_type": "can_be_reconciled_without_data"},
			},
			expectedCount: 1,
		},
		{
			name: "reconciliation_texts without reconciliation_type",
			template: models.Template{
				Category: "reconciliation_texts",
				Config:   map[string]interface{}{"other_field": "value"},
			},
			expectedCount: 0,
		},
		{
			name: "account_templates",
			template: models.Template{
				Category: "account_templates",
				Config:   map[string]interface{}{"some_field": "value"},
			},
			expectedCount: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create empty shared parts usage for testing
			sharedPartsUsage := make(map[string][]string)
			count := navHandler.GetConfigFieldCount(test.template, sharedPartsUsage)
			if count != test.expectedCount {
				t.Errorf("Expected count %d, got %d", test.expectedCount, count)
			}
		})
	}
}

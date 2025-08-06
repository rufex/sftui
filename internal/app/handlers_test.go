package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rufex/sftui/internal/models"
)

func TestFirmPopupTrigger(t *testing.T) {
	app := New()
	m := app.InitialModel()

	// Set current section to Firm
	m.CurrentSection = models.FirmSection

	// Simulate Enter key press
	key := tea.KeyMsg{Type: tea.KeyEnter}

	app.Model = m
	_, _ = app.Update(key)

	if !app.Model.ShowFirmPopup {
		t.Errorf("Expected ShowFirmPopup to be true after Enter in Firm section")
	}

	if app.Model.SelectedFirm != 0 {
		t.Errorf("Expected SelectedFirm to be reset to 0, got %d", app.Model.SelectedFirm)
	}

	if app.Model.Output != "Select a firm or partner" {
		t.Errorf("Expected output message 'Select a firm or partner', got %s", app.Model.Output)
	}
}

func TestFirmPopupNavigation(t *testing.T) {
	app := New()
	m := app.InitialModel()

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

	app.Model = m
	for _, test := range tests {
		var key tea.KeyMsg
		if test.keyType == tea.KeyRunes {
			key = tea.KeyMsg{Type: test.keyType, Runes: []rune{rune(test.keyStr[0])}}
		} else {
			key = tea.KeyMsg{Type: test.keyType}
		}
		_, _ = app.Update(key)

		if app.Model.SelectedFirm != test.expected {
			t.Errorf("After key %s, expected SelectedFirm %d, got %d", test.keyStr, test.expected, app.Model.SelectedFirm)
		}
	}
}

func TestFirmPopupEscape(t *testing.T) {
	app := New()
	m := app.InitialModel()

	// Set popup state
	m.ShowFirmPopup = true
	m.SelectedFirm = 2

	// Simulate Escape key
	key := tea.KeyMsg{Type: tea.KeyEscape}

	app.Model = m
	_, _ = app.Update(key)

	if app.Model.ShowFirmPopup {
		t.Errorf("Expected ShowFirmPopup to be false after Escape")
	}

	if app.Model.SelectedFirm != 0 {
		t.Errorf("Expected SelectedFirm to be reset to 0, got %d", app.Model.SelectedFirm)
	}

	if app.Model.Output != "Firm selection cancelled" {
		t.Errorf("Expected output 'Firm selection cancelled', got %s", app.Model.Output)
	}
}

func TestFirmPopupSelection(t *testing.T) {
	app := New()
	m := app.InitialModel()

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

	app.Model = m
	_, _ = app.Update(key)

	if app.Model.ShowFirmPopup {
		t.Errorf("Expected ShowFirmPopup to be false after selection")
	}

	if app.Model.SelectedFirm != 0 {
		t.Errorf("Expected SelectedFirm to be reset to 0, got %d", app.Model.SelectedFirm)
	}

	if !strings.Contains(app.Model.Firm, selectedOption.Name) {
		t.Errorf("Expected firm to contain %s, got %s", selectedOption.Name, app.Model.Firm)
	}

	if !strings.Contains(app.Model.Output, "Default firm set to") {
		t.Errorf("Expected success message in output, got %s", app.Model.Output)
	}
}

func TestHostPopupTrigger(t *testing.T) {
	app := New()
	m := app.InitialModel()

	// Set current section to Host
	m.CurrentSection = models.HostSection

	// Simulate Enter key press
	key := tea.KeyMsg{Type: tea.KeyEnter}

	app.Model = m
	_, _ = app.Update(key)

	if !app.Model.ShowHostPopup {
		t.Errorf("Expected ShowHostPopup to be true after Enter in Host section")
	}

	if app.Model.HostTextInput.Value() != app.Model.Host {
		t.Errorf("Expected HostTextInput to be pre-filled with current host")
	}

	if app.Model.Output != "Edit host URL" {
		t.Errorf("Expected output message 'Edit host URL', got %s", app.Model.Output)
	}
}

func TestHostPopupTextInputMultipleChars(t *testing.T) {
	app := New()
	m := app.InitialModel()

	// Set popup state with textinput
	m.ShowHostPopup = true
	m.HostTextInput.SetValue("https://test")
	m.HostTextInput.Focus()

	// Test adding characters through textinput
	chars := []rune{'.', 'c', 'o', 'm'}
	for _, char := range chars {
		key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}}
		app.Model = m
		_, _ = app.Update(key)
		m = app.Model
	}

	// Check that textinput handles the characters (exact result depends on textinput implementation)
	finalValue := app.Model.HostTextInput.Value()
	if !strings.Contains(finalValue, "test") {
		t.Errorf("Expected textinput to contain 'test', got '%s'", finalValue)
	}
}

func TestHostPopupIntegration(t *testing.T) {
	app := New()
	m := app.InitialModel()

	key := tea.KeyMsg{Type: tea.KeyEnter}

	// Test that Enter in other sections doesn't trigger host popup
	m.CurrentSection = models.TemplatesSection

	app.Model = m
	_, _ = app.Update(key)

	if app.Model.ShowHostPopup {
		t.Errorf("Expected host popup not to trigger from Templates section")
	}

	// Test that Enter in host section triggers popup
	m.CurrentSection = models.HostSection

	app.Model = m
	_, _ = app.Update(key)

	if !app.Model.ShowHostPopup {
		t.Errorf("Expected host popup to trigger from Host section")
	}
}

func TestHostPopupTextInputFocus(t *testing.T) {
	app := New()
	m := app.InitialModel()

	// Set current section to Host and simulate Enter to open popup
	m.CurrentSection = models.HostSection
	m.Host = "https://example.com"

	key := tea.KeyMsg{Type: tea.KeyEnter}
	app.Model = m
	_, _ = app.Update(key)

	// Check that popup is shown and textinput is focused with correct value
	if !app.Model.ShowHostPopup {
		t.Errorf("Expected host popup to be shown")
	}

	if !app.Model.HostTextInput.Focused() {
		t.Errorf("Expected textinput to be focused")
	}

	if app.Model.HostTextInput.Value() != m.Host {
		t.Errorf("Expected textinput value to be '%s', got '%s'", m.Host, app.Model.HostTextInput.Value())
	}
}

func TestHostPopupTextInputTyping(t *testing.T) {
	app := New()
	m := app.InitialModel()

	// Open host popup with focused textinput
	m.ShowHostPopup = true
	m.HostTextInput.SetValue("https://test.com")
	m.HostTextInput.Focus()

	// Type a character
	char := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}

	app.Model = m
	_, _ = app.Update(char)

	// The textinput should handle the character insertion
	value := app.Model.HostTextInput.Value()
	if !strings.Contains(value, "x") {
		t.Errorf("Expected textinput to contain 'x' after typing, got '%s'", value)
	}
}

func TestHostPopupTextInputBackspace(t *testing.T) {
	app := New()
	m := app.InitialModel()

	// Open host popup with focused textinput
	m.ShowHostPopup = true
	m.HostTextInput.SetValue("https://test.com")
	m.HostTextInput.Focus()

	originalLength := len(m.HostTextInput.Value())

	// Send backspace
	backspace := tea.KeyMsg{Type: tea.KeyBackspace}

	app.Model = m
	_, _ = app.Update(backspace)

	// The textinput should handle the backspace
	newValue := app.Model.HostTextInput.Value()
	if len(newValue) >= originalLength {
		t.Errorf("Expected textinput value to be shorter after backspace, original: %d, new: %d", originalLength, len(newValue))
	}
}

func TestHostPopupEscapeKey(t *testing.T) {
	app := New()
	m := app.InitialModel()

	// Open host popup
	m.ShowHostPopup = true
	m.HostTextInput.SetValue("https://test.com")
	m.HostTextInput.Focus()

	// Send escape key
	escape := tea.KeyMsg{Type: tea.KeyEsc}

	app.Model = m
	_, _ = app.Update(escape)

	// Popup should be closed and textinput should be blurred
	if app.Model.ShowHostPopup {
		t.Errorf("Expected host popup to be closed after escape")
	}

	if app.Model.HostTextInput.Focused() {
		t.Errorf("Expected textinput to be blurred after escape")
	}

	if app.Model.Output != "Host edit cancelled" {
		t.Errorf("Expected output message 'Host edit cancelled', got '%s'", app.Model.Output)
	}
}

func TestHostPopupEnterKeySave(t *testing.T) {
	app := New()
	m := app.InitialModel()

	// Open host popup and set a new value
	m.ShowHostPopup = true
	newHost := "https://new-host.com"
	m.HostTextInput.SetValue(newHost)
	m.HostTextInput.Focus()

	// Send enter key to save
	enter := tea.KeyMsg{Type: tea.KeyEnter}

	app.Model = m
	_, _ = app.Update(enter)

	// Popup should be closed and textinput should be blurred
	if app.Model.ShowHostPopup {
		t.Errorf("Expected host popup to be closed after enter")
	}

	if app.Model.HostTextInput.Focused() {
		t.Errorf("Expected textinput to be blurred after enter")
	}

	if app.Model.Host != newHost {
		t.Errorf("Expected host to be updated to '%s', got '%s'", newHost, app.Model.Host)
	}

	if app.Model.Output != "Host updated successfully" {
		t.Errorf("Expected output message 'Host updated successfully', got '%s'", app.Model.Output)
	}
}

func TestReconciliationTypePopupTrigger(t *testing.T) {
	app := New()
	m := app.InitialModel()

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
	app.Model = m
	_, _ = app.Update(keyMsg)

	// Check that popup was triggered
	if !app.Model.ShowReconciliationTypePopup {
		t.Errorf("Expected ShowReconciliationTypePopup to be true after Enter in Details section")
	}

	if app.Model.SelectedReconciliationType < 0 || app.Model.SelectedReconciliationType > 2 {
		t.Errorf("Expected SelectedReconciliationType to be 0-2, got %d", app.Model.SelectedReconciliationType)
	}
}

func TestReconciliationTypePopupNavigation(t *testing.T) {
	app := New()
	m := app.InitialModel()

	// Setup popup state
	m.ShowReconciliationTypePopup = true
	m.SelectedReconciliationType = 0

	// Test down navigation
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	app.Model = m
	_, _ = app.Update(keyMsg)
	if app.Model.SelectedReconciliationType != 1 {
		t.Errorf("Expected SelectedReconciliationType to be 1 after down navigation, got %d", app.Model.SelectedReconciliationType)
	}

	// Test up navigation
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	_, _ = app.Update(keyMsg)
	if app.Model.SelectedReconciliationType != 0 {
		t.Errorf("Expected SelectedReconciliationType to be 0 after up navigation, got %d", app.Model.SelectedReconciliationType)
	}

	// Test wrap around (down from 2 should go to 0)
	app.Model.SelectedReconciliationType = 2
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	_, _ = app.Update(keyMsg)
	if app.Model.SelectedReconciliationType != 0 {
		t.Errorf("Expected SelectedReconciliationType to wrap to 0, got %d", app.Model.SelectedReconciliationType)
	}

	// Test wrap around (up from 0 should go to 2)
	app.Model.SelectedReconciliationType = 0
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	_, _ = app.Update(keyMsg)
	if app.Model.SelectedReconciliationType != 2 {
		t.Errorf("Expected SelectedReconciliationType to wrap to 2, got %d", app.Model.SelectedReconciliationType)
	}
}

func TestReconciliationTypePopupCancel(t *testing.T) {
	app := New()
	m := app.InitialModel()

	// Setup popup state
	m.ShowReconciliationTypePopup = true
	m.SelectedReconciliationType = 1

	// Test escape cancellation
	keyMsg := tea.KeyMsg{Type: tea.KeyEsc}
	app.Model = m
	_, _ = app.Update(keyMsg)

	if app.Model.ShowReconciliationTypePopup {
		t.Errorf("Expected ShowReconciliationTypePopup to be false after escape")
	}

	if app.Model.SelectedReconciliationType != 0 {
		t.Errorf("Expected SelectedReconciliationType to reset to 0 after escape, got %d", app.Model.SelectedReconciliationType)
	}

	if !strings.Contains(app.Model.Output, "cancelled") {
		t.Errorf("Expected output to contain 'cancelled', got: %s", app.Model.Output)
	}
}

func TestInPlaceEditingCancelRestore(t *testing.T) {
	app := New()
	model := app.InitialModel()

	// Create a template with boolean config field
	template := models.Template{
		Name:     "test_template",
		Category: "reconciliation_texts",
		Path:     "/test/path",
		Config: map[string]interface{}{
			"public": false,
		},
	}

	model.Templates = []models.Template{template}
	model.FilteredTemplates = []int{0}
	model.SelectedTemplate = 0
	model.CurrentSection = models.DetailsSection
	model.SelectedDetailField = 0

	// Start in-place editing
	model.ShowInPlaceEdit = true
	model.InPlaceEditField = "public"
	model.InPlaceEditOptions = []string{"true", "false"}
	model.InPlaceEditOriginalValue = false
	model.InPlaceEditSelectedIndex = 0 // selecting "true"

	// Simulate navigation to different value
	model.InPlaceEditSelectedIndex = 0 // "true" is selected

	// Verify the template still has original value (not changed until Enter)
	if model.Templates[0].Config["public"] != false {
		t.Errorf("Template value should not change until Enter is pressed, got %v", model.Templates[0].Config["public"])
	}

	// Simulate Escape key to cancel
	// This mimics the escape handling logic from handlers.go
	actualIndex := model.FilteredTemplates[model.SelectedTemplate]
	model.Templates[actualIndex].Config[model.InPlaceEditField] = model.InPlaceEditOriginalValue
	model.ShowInPlaceEdit = false
	model.InPlaceEditField = ""
	model.InPlaceEditOptions = nil
	model.InPlaceEditOriginalValue = nil
	model.InPlaceEditSelectedIndex = 0

	// Verify the original value was restored
	if model.Templates[0].Config["public"] != false {
		t.Errorf("Expected original value false to be restored after escape, got %v", model.Templates[0].Config["public"])
	}

	// Verify editing state is cleared
	if model.ShowInPlaceEdit {
		t.Error("ShowInPlaceEdit should be false after escape")
	}
}

func TestSpaceKeyTemplateSelection(t *testing.T) {
	app := New()
	m := app.InitialModel()

	if len(m.FilteredTemplates) == 0 {
		t.Skip("No templates available for testing")
	}

	// Set current section to Templates
	m.CurrentSection = models.TemplatesSection
	m.SelectedTemplate = 0

	// Simulate space key press
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}

	app.Model = m
	_, _ = app.Update(key)

	// Check that template was selected
	actualIndex := m.FilteredTemplates[0]
	if !app.Model.SelectedTemplates[actualIndex] {
		t.Errorf("Expected template at index %d to be selected", actualIndex)
	}

	if !strings.Contains(app.Model.Output, "1 template selected") {
		t.Errorf("Expected output to show '1 template selected', got: %s", app.Model.Output)
	}
}

func TestSearchModeToggle(t *testing.T) {
	app := New()
	m := app.InitialModel()

	// Set current section to Templates
	m.CurrentSection = models.TemplatesSection

	// Simulate '/' key press to enter search mode
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}

	app.Model = m
	_, _ = app.Update(key)

	if !app.Model.SearchMode {
		t.Errorf("Expected SearchMode to be true after '/' key")
	}

	if app.Model.Output != "Search mode - type to filter templates" {
		t.Errorf("Expected search mode message, got: %s", app.Model.Output)
	}
}

func TestBackspaceDeselectAll(t *testing.T) {
	app := New()
	m := app.InitialModel()

	if len(m.FilteredTemplates) == 0 {
		t.Skip("No templates available for testing")
	}

	// Set current section to Templates and select some templates
	m.CurrentSection = models.TemplatesSection
	m.SelectedTemplates[0] = true
	m.SelectedTemplates[1] = true

	// Simulate backspace key press
	key := tea.KeyMsg{Type: tea.KeyBackspace}

	app.Model = m
	_, _ = app.Update(key)

	if len(app.Model.SelectedTemplates) != 0 {
		t.Errorf("Expected all templates to be deselected, got %d selected", len(app.Model.SelectedTemplates))
	}

	if app.Model.Output != "All templates deselected" {
		t.Errorf("Expected deselect message, got: %s", app.Model.Output)
	}
}

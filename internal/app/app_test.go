package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rufex/sftui/internal/models"
)

func TestInitialModel(t *testing.T) {
	application := New()
	m := application.InitialModel()

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

func TestTemplateSelection(t *testing.T) {
	app := New()
	m := app.InitialModel()

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

func TestHostTextInputInitialization(t *testing.T) {
	app := New()
	m := app.InitialModel()

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

func TestHostTextInputAdvancedFeatures(t *testing.T) {
	app := New()
	m := app.InitialModel()

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

func TestBuildSharedPartsMapping(t *testing.T) {
	app := New()
	m := app.InitialModel()

	// Check that shared parts mapping is built
	if m.SharedPartsUsage == nil {
		t.Errorf("Expected SharedPartsUsage to be initialized")
	}

	// Find a shared part template to test mapping
	hasSharedPart := false
	for _, template := range m.Templates {
		if template.Category == "shared_parts" {
			hasSharedPart = true
			break
		}
	}

	if hasSharedPart && len(m.SharedPartsUsage) == 0 {
		t.Errorf("Expected SharedPartsUsage to contain mappings when shared parts exist")
	}
}

func TestAppInit(t *testing.T) {
	app := New()
	cmd := app.Init()

	// Init should return nil command
	if cmd != nil {
		t.Errorf("Expected Init() to return nil, got %v", cmd)
	}
}

func TestAppView(t *testing.T) {
	app := New()
	m := app.InitialModel()
	app.Model = m

	// Set basic dimensions for rendering
	m.Width = 80
	m.Height = 24

	view := app.View()

	// Should contain basic UI elements
	if !strings.Contains(view, "Templates") {
		t.Errorf("Expected view to contain 'Templates'")
	}

	if !strings.Contains(view, "Details") {
		t.Errorf("Expected view to contain 'Details'")
	}
}

func TestAppViewHelp(t *testing.T) {
	app := New()
	m := app.InitialModel()
	m.ShowHelp = true
	m.Width = 80
	m.Height = 24
	app.Model = m

	view := app.View()

	if !strings.Contains(view, "Key Bindings") {
		t.Errorf("Help view should contain 'Key Bindings'")
	}
}

func TestAppViewTooSmall(t *testing.T) {
	app := New()
	m := app.InitialModel()
	m.Width = 80
	m.Height = 5 // Too small
	app.Model = m

	view := app.View()

	if !strings.Contains(view, "Terminal too small") {
		t.Errorf("Expected 'Terminal too small' message for small terminal")
	}
}

func TestAppWindowSizeHandling(t *testing.T) {
	app := New()
	m := app.InitialModel()
	app.Model = m

	// Test window size message
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	_, _ = app.Update(msg)

	if app.Model.Width != 100 {
		t.Errorf("Expected width to be 100, got %d", app.Model.Width)
	}

	if app.Model.Height != 30 {
		t.Errorf("Expected height to be 30, got %d", app.Model.Height)
	}
}

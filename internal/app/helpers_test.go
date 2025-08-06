package app

import (
	"testing"

	"github.com/rufex/sftui/internal/models"
)

func TestGetConfigFieldCount(t *testing.T) {
	app := New()

	tests := []struct {
		name          string
		template      models.Template
		expectedCount int
	}{
		{
			name: "template with public field",
			template: models.Template{
				Config: map[string]interface{}{
					"public": true,
				},
			},
			expectedCount: 1,
		},
		{
			name: "template with reconciliation_type field",
			template: models.Template{
				Config: map[string]interface{}{
					"reconciliation_type": "can_be_reconciled_without_data",
				},
			},
			expectedCount: 1,
		},
		{
			name: "template with multiple fields",
			template: models.Template{
				Config: map[string]interface{}{
					"public":                         true,
					"reconciliation_type":            "can_be_reconciled_without_data",
					"virtual_account_number":         "123",
					"allow_duplicate_reconciliation": false,
					"is_active":                      true,
				},
			},
			expectedCount: 5,
		},
		{
			name: "template with no config fields",
			template: models.Template{
				Config: map[string]interface{}{
					"other_field": "value",
				},
			},
			expectedCount: 0,
		},
		{
			name: "template with all possible config fields",
			template: models.Template{
				Config: map[string]interface{}{
					"public":                         true,
					"reconciliation_type":            "can_be_reconciled_without_data",
					"virtual_account_number":         "123",
					"allow_duplicate_reconciliation": false,
					"is_active":                      true,
					"use_full_width":                 false,
					"downloadable_as_docx":           true,
					"encoding":                       "UTF-8",
					"published":                      true,
					"hide_code":                      false,
					"externally_managed":             false,
				},
			},
			expectedCount: 11,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			count := app.GetConfigFieldCount(test.template)
			if count != test.expectedCount {
				t.Errorf("Expected count %d, got %d", test.expectedCount, count)
			}
		})
	}
}

func TestGetSelectedConfigField(t *testing.T) {
	app := New()

	template := models.Template{
		Config: map[string]interface{}{
			"public":              true,
			"reconciliation_type": "can_be_reconciled_without_data",
			"is_active":           true,
		},
	}

	tests := []struct {
		fieldIndex    int
		expectedField string
	}{
		{0, "public"},
		{1, "reconciliation_type"},
		{2, "is_active"},
	}

	for _, test := range tests {
		field := app.GetSelectedConfigField(template, test.fieldIndex)
		if field != test.expectedField {
			t.Errorf("Expected field '%s' at index %d, got '%s'", test.expectedField, test.fieldIndex, field)
		}
	}

	// Test out of bounds
	field := app.GetSelectedConfigField(template, 10)
	if field != "" {
		t.Errorf("Expected empty string for out of bounds index, got '%s'", field)
	}
}

func TestGetFieldEditOptions(t *testing.T) {
	app := New()

	tests := []struct {
		fieldName       string
		expectedOptions []string
	}{
		{"public", []string{"true", "false"}},
		{"is_active", []string{"true", "false"}},
		{"use_full_width", []string{"true", "false"}},
		{"downloadable_as_docx", []string{"true", "false"}},
		{"published", []string{"true", "false"}},
		{"hide_code", []string{"true", "false"}},
		{"externally_managed", []string{"true", "false"}},
		{"allow_duplicate_reconciliation", []string{"true", "false"}},
		{"reconciliation_type", []string{"can_be_reconciled_without_data", "reconciliation_not_necessary", "only_reconciled_with_data"}},
		{"encoding", []string{"UTF-8", "ISO-8859-1", "Windows-1252"}},
		{"unknown_field", nil},
	}

	for _, test := range tests {
		t.Run(test.fieldName, func(t *testing.T) {
			options := app.GetFieldEditOptions(test.fieldName, nil)

			if test.expectedOptions == nil {
				if options != nil {
					t.Errorf("Expected nil options for field '%s', got %v", test.fieldName, options)
				}
				return
			}

			if len(options) != len(test.expectedOptions) {
				t.Errorf("Expected %d options for field '%s', got %d", len(test.expectedOptions), test.fieldName, len(options))
				return
			}

			for i, expected := range test.expectedOptions {
				if options[i] != expected {
					t.Errorf("Expected option '%s' at index %d for field '%s', got '%s'", expected, i, test.fieldName, options[i])
				}
			}
		})
	}
}

func TestIsFieldEditable(t *testing.T) {
	app := New()

	editableFields := []string{
		"public",
		"is_active",
		"use_full_width",
		"downloadable_as_docx",
		"published",
		"hide_code",
		"externally_managed",
		"allow_duplicate_reconciliation",
		"reconciliation_type",
		"encoding",
	}

	nonEditableFields := []string{
		"name",
		"category",
		"path",
		"text_parts",
		"unknown_field",
		"virtual_account_number",
	}

	for _, field := range editableFields {
		if !app.IsFieldEditable(field) {
			t.Errorf("Field '%s' should be editable", field)
		}
	}

	for _, field := range nonEditableFields {
		if app.IsFieldEditable(field) {
			t.Errorf("Field '%s' should not be editable", field)
		}
	}
}

func TestUpdateConfigField(t *testing.T) {
	app := New()

	// Test with a valid template path from fixtures
	manager := app.templateManager
	templates := manager.LoadTemplates()

	if len(templates) == 0 {
		t.Skip("No templates available for testing")
	}

	// Use the first template's path for testing
	templatePath := templates[0].Path

	// Test boolean field conversion
	err := app.updateConfigField(templatePath, "public", "true")
	if err != nil {
		t.Errorf("Expected no error updating boolean field, got %v", err)
	}

	// Test string field (no conversion)
	err = app.updateConfigField(templatePath, "reconciliation_type", "can_be_reconciled_without_data")
	if err != nil {
		t.Errorf("Expected no error updating string field, got %v", err)
	}

	// Test encoding field (string, no conversion)
	err = app.updateConfigField(templatePath, "encoding", "UTF-8")
	if err != nil {
		t.Errorf("Expected no error updating encoding field, got %v", err)
	}
}

package app

import (
	"github.com/rufex/sftui/internal/models"
)

func (a *App) GetConfigFieldCount(template models.Template) int {
	count := 0

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

	for _, key := range configKeys {
		if _, exists := template.Config[key]; exists {
			count++
		}
	}

	return count
}

func (a *App) GetSelectedConfigField(template models.Template, fieldIndex int) string {
	currentIndex := 0

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

	for _, key := range configKeys {
		if _, exists := template.Config[key]; exists {
			if currentIndex == fieldIndex {
				return key
			}
			currentIndex++
		}
	}

	return ""
}

func (a *App) GetFieldEditOptions(fieldName string, currentValue interface{}) []string {
	switch fieldName {
	case "public", "is_active", "use_full_width", "downloadable_as_docx",
		"published", "hide_code", "externally_managed", "allow_duplicate_reconciliation":
		return []string{"true", "false"}
	case "reconciliation_type":
		return []string{"can_be_reconciled_without_data", "reconciliation_not_necessary", "only_reconciled_with_data"}
	case "encoding":
		return []string{"UTF-8", "ISO-8859-1", "Windows-1252"}
	default:
		return nil
	}
}

func (a *App) IsFieldEditable(fieldName string) bool {
	return a.GetFieldEditOptions(fieldName, nil) != nil
}

func (a *App) updateConfigField(templatePath, fieldName, newValue string) error {
	var value interface{}
	switch fieldName {
	case "public", "is_active", "use_full_width", "downloadable_as_docx",
		"published", "hide_code", "externally_managed", "allow_duplicate_reconciliation":
		value = newValue == "true"
	default:
		value = newValue
	}

	return a.configManager.UpdateConfigField(templatePath, fieldName, value)
}

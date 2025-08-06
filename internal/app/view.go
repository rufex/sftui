package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/rufex/sftui/internal/models"
)

func (a *App) View() string {
	if a.Model.ShowHelp {
		return a.uiRenderer.HelpView(a.Model)
	}

	if a.Model.ShowActionPopup {
		return a.uiRenderer.ActionPopupView(a.Model)
	}

	if a.Model.ShowFirmPopup {
		return a.uiRenderer.FirmPopupView(a.Model)
	}

	if a.Model.ShowHostPopup {
		return a.uiRenderer.HostPopupView(a.Model)
	}

	if a.Model.ShowReconciliationTypePopup {
		return a.uiRenderer.ReconciliationTypePopupView(a.Model)
	}

	if a.Model.ShowTextPartPopup {
		return a.uiRenderer.TextPartPopupView(a.Model)
	}

	if a.Model.Height < 10 {
		return "Terminal too small"
	}

	var searchBar string
	searchBarHeight := 0
	if a.Model.SearchMode {
		searchBarHeight = 3
		searchStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("6")).
			Padding(0, 1)

		searchContent := fmt.Sprintf("Search: %s_", a.Model.SearchQuery)
		searchBar = searchStyle.Width(a.Model.Width - 4).Render(searchContent)
	}

	topContentHeight := 1
	outputContentHeight := 1
	statusHeight := 1

	reservedLines := (4 * 2) + 4 + statusHeight + searchBarHeight
	availableContentHeight := a.Model.Height - reservedLines
	if availableContentHeight < 1 {
		availableContentHeight = 1
	}

	halfWidth := (a.Model.Width - 6) / 2
	fullWidth := a.Model.Width - 4

	firmBox := a.uiRenderer.RenderSection(a.Model, models.FirmSection, "Firm", a.uiRenderer.FirmView(a.Model), halfWidth, topContentHeight)
	hostBox := a.uiRenderer.RenderSection(a.Model, models.HostSection, "Host", a.uiRenderer.HostView(a.Model), halfWidth, topContentHeight)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, firmBox, hostBox)

	selectedCount := len(a.Model.SelectedTemplates)
	var templatesTitle string

	templateCount := len(a.Model.FilteredTemplates)
	totalCount := len(a.Model.Templates)

	if a.Model.SearchMode && a.Model.SearchQuery != "" {
		if selectedCount > 0 {
			templatesTitle = fmt.Sprintf("Templates (%d/%d) - %d selected", templateCount, totalCount, selectedCount)
		} else {
			templatesTitle = fmt.Sprintf("Templates (%d/%d)", templateCount, totalCount)
		}
	} else {
		if selectedCount > 0 {
			templatesTitle = fmt.Sprintf("Templates (%d) - %d selected", totalCount, selectedCount)
		} else {
			templatesTitle = fmt.Sprintf("Templates (%d)", totalCount)
		}
	}
	templatesContent := a.uiRenderer.TemplatesViewWithHeightAndWidth(a.Model, availableContentHeight, halfWidth)
	detailsContent := a.uiRenderer.DetailsViewWithHeightAndWidth(a.Model, availableContentHeight, halfWidth)

	templatesBox := a.uiRenderer.RenderSection(a.Model, models.TemplatesSection, templatesTitle, templatesContent, halfWidth, availableContentHeight)
	detailsBox := a.uiRenderer.RenderSection(a.Model, models.DetailsSection, "Details", detailsContent, halfWidth, availableContentHeight)
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, templatesBox, detailsBox)

	outputBox := a.uiRenderer.RenderSection(a.Model, models.OutputSection, "Output", a.uiRenderer.OutputView(a.Model), fullWidth, outputContentHeight)

	statusBar := a.uiRenderer.StatusBarView(a.Model)

	if a.Model.SearchMode {
		return lipgloss.JoinVertical(lipgloss.Left, searchBar, topRow, mainRow, outputBox, statusBar)
	} else {
		return lipgloss.JoinVertical(lipgloss.Left, topRow, mainRow, outputBox, statusBar)
	}
}

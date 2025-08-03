package models

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type Section int

const (
	FirmSection Section = iota
	HostSection
	TemplatesSection
	DetailsSection
	OutputSection
)

type Template struct {
	Name     string
	Path     string
	Type     string
	Category string
	Config   map[string]interface{}
}

type SilverfinConfig struct {
	DefaultFirmIDs map[string]string            `json:"defaultFirmIDs"`
	Host           string                       `json:"host"`
	Firms          map[string]map[string]string `json:",inline"`
}

type FirmOption struct {
	ID   string
	Name string
	Type string // "firm" or "partner"
}

type Model struct {
	CurrentSection              Section
	Templates                   []Template
	SelectedTemplate            int
	TemplatesOffset             int
	SelectedTemplates           map[int]bool
	SearchMode                  bool
	SearchQuery                 string
	FilteredTemplates           []int
	ShowActionPopup             bool
	SelectedAction              int
	ShowFirmPopup               bool
	SelectedFirm                int
	FirmOptions                 []FirmOption
	ShowHostPopup               bool
	HostTextInput               textinput.Model
	Firm                        string
	Host                        string
	ShowHelp                    bool
	Output                      string
	Width                       int
	Height                      int
	SelectedDetailField         int
	ShowReconciliationTypePopup bool
	SelectedReconciliationType  int
	SelectedTextPart            int
	ShowTextPartPopup           bool
	TextPartNameInput           textinput.Model
	TextPartPathInput           textinput.Model
	TextPartEditMode            string              // "name" or "path"
	SharedPartsUsage            map[string][]string // maps template handle to shared part names
}

var (
	ActiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("6")) // Cyan

	InactiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")) // Gray

	SelectedItemStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("6")). // Cyan (same as active border)
				Foreground(lipgloss.Color("15")) // White

	CategoryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")) // Gray

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1)
)

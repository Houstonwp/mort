package tui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Accent      lipgloss.Color
	Subtle      lipgloss.Color
	Header      lipgloss.Style
	Subtitle    lipgloss.Style
	ListBox     lipgloss.Style
	SearchPrompt lipgloss.Style
	HelperText  lipgloss.Style
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	Panel       lipgloss.Style
	RatesPanel  lipgloss.Style
	Label       lipgloss.Style
	Value       lipgloss.Style
	Section     lipgloss.Style
}

func NewTheme() Theme {
	accent := lipgloss.Color("99")
	subtle := lipgloss.Color("245")

	return Theme{
		Accent:   accent,
		Subtle:   subtle,
		Header:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")),
		Subtitle: lipgloss.NewStyle().Foreground(subtle),
		ListBox:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1),
		SearchPrompt: lipgloss.NewStyle().Foreground(accent).PaddingLeft(1),
		HelperText:   lipgloss.NewStyle().Foreground(subtle),
		TabActive:    lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(accent).Bold(true).Padding(0, 2),
		TabInactive:  lipgloss.NewStyle().Foreground(subtle).Padding(0, 2),
		Panel:        lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(1, 2),
		RatesPanel:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(0, 0),
		Label:        lipgloss.NewStyle().Foreground(accent).Bold(true),
		Value:        lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		Section:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Underline(true),
	}
}

package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"

	"mort/internal/tuiapp"
)

type detailView struct {
	detail      *tuiapp.TableDetail
	index       int
	tab         int
	rates       table.Model
	viewport    viewport.Model
	textContent string
}

func newDetailView() detailView {
	rateColumns := []table.Column{
		{Title: "Age", Width: 8},
		{Title: "Dur", Width: 6},
		{Title: "Rate", Width: 20},
	}
	rt := table.New(
		table.WithColumns(rateColumns),
		table.WithRows(nil),
		table.WithFocused(true),
	)
	rt.KeyMap = table.DefaultKeyMap()
	tableStyles := table.DefaultStyles()
	tableStyles.Header = tableStyles.Header.Background(accentColor).Foreground(lipgloss.Color("0")).Bold(true)
	tableStyles.Cell = tableStyles.Cell.Foreground(lipgloss.Color("252")).Padding(0, 0)
	tableStyles.Selected = tableStyles.Selected.Foreground(lipgloss.Color("15")).Background(lipgloss.Color("236"))
	rt.SetStyles(tableStyles)

	tv := viewport.New(0, 0)
	km := viewport.DefaultKeyMap()
	km.Up = key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up"))
	km.Down = key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down"))
	km.PageUp = key.NewBinding(key.WithKeys("pgup", "b"), key.WithHelp("PgUp/b", "page up"))
	km.PageDown = key.NewBinding(key.WithKeys("pgdown", "f"), key.WithHelp("PgDn/f", "page down"))
	km.HalfPageUp = key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("Ctrl+U", "½ page up"))
	km.HalfPageDown = key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("Ctrl+D", "½ page down"))
	tv.KeyMap = km

	return detailView{
		rates:    rt,
		viewport: tv,
	}
}

func (dv *detailView) SetDetail(detail *tuiapp.TableDetail) {
	dv.detail = detail
	dv.index = 0
	dv.tab = tabClassification
	dv.resetViewport()
	dv.refreshRatesTable()
}

func (dv *detailView) Update(msg tea.Msg) tea.Cmd {
	if dv.detail == nil {
		return nil
	}
	if dv.tab == tabRates {
		var cmd tea.Cmd
		dv.rates, cmd = dv.rates.Update(msg)
		return cmd
	}
	var cmd tea.Cmd
	dv.viewport, cmd = dv.viewport.Update(msg)
	return cmd
}

func (dv *detailView) View(width, height int) string {
	if dv.detail == nil {
		return "Loading table detail…"
	}
	contentWidth := max(20, width-4)
	title := headerStyle.Render(dv.detail.Classification.TableName)
	subtitle := helperTextStyle.Render(fmt.Sprintf("%s • Table %d of %d • Version %s",
		dv.detail.Classification.ProviderName, dv.index+1, len(dv.detail.Tables), dv.detail.Version))
	info := lipgloss.NewStyle().Width(contentWidth).Render(
		lipgloss.JoinVertical(lipgloss.Left, title, subtitle),
	)

	tabs := lipgloss.NewStyle().Width(contentWidth).Render(renderTabs(dv.tab))
	footer := helperTextStyle.Width(contentWidth).Render("←/→ table • Tab switch panel • j/k scroll rates • esc back")

	bodyHeight := availableBodyHeight(height, info, tabs, footer)
	bodyPanel := dv.renderBodyPanel(contentWidth, bodyHeight)

	content := lipgloss.JoinVertical(lipgloss.Left, info, tabs, bodyPanel, footer)
	return lipgloss.NewStyle().
		Width(max(1, width)).
		Height(max(1, height)).
		Render(content)
}

func renderTabs(active int) string {
	var segments []string
	for i, label := range tabLabels {
		style := tabInactiveStyle
		if i == active {
			style = tabActiveStyle
		}
		segments = append(segments, style.Render(label))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, segments...)
}

func availableBodyHeight(total int, header, tabs, footer string) int {
	if total <= 0 {
		total = lipgloss.Height(header) + lipgloss.Height(tabs) + lipgloss.Height(footer) + 12
	}
	used := lipgloss.Height(header) + lipgloss.Height(tabs) + lipgloss.Height(footer) + 2
	remaining := total - used
	if remaining < 3 {
		remaining = 3
	}
	return remaining
}

func (dv *detailView) renderBodyPanel(width, bodyHeight int) string {
	switch dv.tab {
	case tabRates:
		return dv.renderRatesPanel(width, bodyHeight)
	case tabMetadata:
		return dv.renderScrollablePanel(width, bodyHeight, renderMetadata(dv.detail, dv.index, width-6))
	default:
		return dv.renderScrollablePanel(width, bodyHeight, renderClassification(dv.detail, width-6))
	}
}

func (dv *detailView) renderScrollablePanel(width, bodyHeight int, content string) string {
	innerWidth := max(10, width-4)
	viewHeight := max(1, bodyHeight-2)
	if dv.viewport.Width != innerWidth {
		dv.viewport.Width = innerWidth
	}
	if dv.viewport.Height != viewHeight {
		dv.viewport.Height = viewHeight
	}
	wrapped := wordwrap.String(content, innerWidth)
	if wrapped != dv.textContent {
		dv.textContent = wrapped
		dv.viewport.SetContent(wrapped)
		dv.viewport.GotoTop()
	}
	return panelStyle.Width(width).Height(bodyHeight).Render(dv.viewport.View())
}

func (dv *detailView) renderRatesPanel(width, bodyHeight int) string {
	innerWidth := max(10, width-4)
	dv.rates.SetWidth(innerWidth)
	dv.rates.SetHeight(max(1, bodyHeight-2))
	return ratesPanelStyle.Width(width).Height(bodyHeight).Render(dv.rates.View())
}

func (dv *detailView) refreshRatesTable() {
	if dv.detail == nil || len(dv.detail.Tables) == 0 {
		dv.rates.SetRows(nil)
		dv.rates.Blur()
		return
	}
	if dv.index >= len(dv.detail.Tables) {
		dv.index = len(dv.detail.Tables) - 1
	}
	tableData := dv.detail.Tables[dv.index]
	rows := make([]table.Row, 0, len(tableData.Rates))
	for _, rate := range tableData.Rates {
		duration := "-"
		if rate.Duration != nil {
			duration = fmt.Sprintf("%d", *rate.Duration)
		}
		value := "N/A"
		if rate.Rate != nil {
			value = fmt.Sprintf("%.6f", *rate.Rate)
		}
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", rate.Age),
			duration,
			value,
		})
	}
	dv.rates.SetRows(rows)
	if dv.tab == tabRates {
		dv.rates.Focus()
	} else {
		dv.rates.Blur()
	}
}

func (dv *detailView) NextTable() {
	if dv.detail == nil {
		return
	}
	if dv.index < len(dv.detail.Tables)-1 {
		dv.index++
		dv.refreshRatesTable()
		dv.resetViewport()
	}
}

func (dv *detailView) PrevTable() {
	if dv.detail == nil {
		return
	}
	if dv.index > 0 {
		dv.index--
		dv.refreshRatesTable()
		dv.resetViewport()
	}
}

func (dv *detailView) NextTab() {
	if dv.detail == nil {
		return
	}
	dv.tab = (dv.tab + 1) % len(tabLabels)
	dv.refreshRatesTable()
	dv.resetViewport()
}

func (dv *detailView) PrevTab() {
	if dv.detail == nil {
		return
	}
	dv.tab--
	if dv.tab < 0 {
		dv.tab = len(tabLabels) - 1
	}
	dv.refreshRatesTable()
	dv.resetViewport()
}

func (dv *detailView) SetTab(idx int) {
	if dv.detail == nil {
		return
	}
	if idx >= 0 && idx < len(tabLabels) {
		dv.tab = idx
		dv.refreshRatesTable()
		dv.resetViewport()
	}
}

func (dv *detailView) resetViewport() {
	dv.textContent = ""
	dv.viewport.SetYOffset(0)
}

func renderClassification(detail *tuiapp.TableDetail, width int) string {
	if detail == nil {
		return "No classification data."
	}
	c := detail.Classification
	var b strings.Builder
	b.WriteString(sectionTitleStyle.Render("Overview"))
	b.WriteString("\n\n")
	appendKV(&b, "Table Identity", c.TableIdentity, width)
	appendKV(&b, "Provider Domain", c.ProviderDomain, width)
	appendKV(&b, "Provider Name", c.ProviderName, width)
	appendKV(&b, "Reference", c.TableReference, width)
	appendKV(&b, "Content Type", fmt.Sprintf("%s (%s)", c.ContentType.Label, c.ContentType.Code), width)
	appendKV(&b, "Description", c.TableDescription, width)
	if c.Comments != "" {
		appendKV(&b, "Comments", c.Comments, width)
	}
	if len(c.Keywords) > 0 {
		appendKV(&b, "Keywords", strings.Join(c.Keywords, ", "), width)
	}
	return strings.TrimSpace(b.String())
}

func renderMetadata(detail *tuiapp.TableDetail, tableIdx int, width int) string {
	if detail == nil || tableIdx >= len(detail.Tables) {
		return "No metadata."
	}
	table := detail.Tables[tableIdx]
	if table.Metadata == nil {
		return "No metadata attached to this table."
	}
	meta := table.Metadata
	var b strings.Builder
	b.WriteString(sectionTitleStyle.Render("Table Metadata"))
	b.WriteString("\n\n")
	appendKV(&b, "Scaling Factor", meta.ScalingFactor, width)
	appendKV(&b, "Data Type", fmt.Sprintf("%s (%s)", meta.DataType.Label, meta.DataType.Code), width)
	appendKV(&b, "Nation", fmt.Sprintf("%s (%s)", meta.Nation.Label, meta.Nation.Code), width)
	appendKV(&b, "Description", meta.TableDescription, width)
	if len(meta.Axes) > 0 {
		b.WriteString("\n")
		b.WriteString(sectionTitleStyle.Render("Axes"))
		b.WriteString("\n\n")
		for _, axis := range meta.Axes {
			appendBullet(&b, fmt.Sprintf("%s (%s)", axis.AxisName, axis.ID), width)
			appendBullet(&b, fmt.Sprintf("%s–%s step %s (%s)",
				axis.MinValue, axis.MaxValue, axis.Increment, axis.ScaleType.Label), width)
			b.WriteString("\n")
		}
	}
	return strings.TrimSpace(b.String())
}

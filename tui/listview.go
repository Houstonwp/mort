package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mort/internal/tuiapp"
)

type listView struct {
	model       list.Model
	totalWidth  int
	totalHeight int
}

func newListView() listView {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetSpacing(0)
	delegate.Styles.NormalTitle = headerStyle.Copy().Faint(true)
	delegate.Styles.SelectedTitle = headerStyle.Copy()
	delegate.Styles.NormalDesc = helperTextStyle
	delegate.Styles.SelectedDesc = helperTextStyle.Copy().Foreground(lipgloss.Color("252"))

	l := list.New(nil, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowFilter(true)
	l.SetShowHelp(true)
	l.SetShowPagination(true)
	l.SetStatusBarItemName("table", "tables")
	l.Paginator.Type = paginator.Arabic
	l.Paginator.ArabicFormat = "%d/%d"
	l.Styles.Title = headerStyle
	l.Styles.NoItems = helperTextStyle
	l.Styles.PaginationStyle = helperTextStyle
	l.Styles.HelpStyle = helperTextStyle

	lv := listView{
		model:       l,
		totalWidth:  80,
		totalHeight: 24,
	}
	lv.applySize()
	return lv
}

func (lv *listView) SetItems(items []tuiapp.TableSummary) {
	listItems := make([]list.Item, len(items))
	for i, summary := range items {
		listItems[i] = tableItem{summary: summary}
	}
	lv.model.SetItems(listItems)
	if len(listItems) > 0 {
		lv.model.Select(0)
	}
}

func (lv *listView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	lv.model, cmd = lv.model.Update(msg)
	return cmd
}

func (lv *listView) SetSize(totalWidth, totalHeight int) {
	if totalWidth > 0 {
		lv.totalWidth = totalWidth
	}
	if totalHeight > 0 {
		lv.totalHeight = totalHeight
	}
	lv.applySize()
}

func (lv listView) View(totalWidth, totalHeight int) string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Width(lv.panelWidth()).Render("Mortality Tables"),
		panelStyle.
			Width(lv.panelWidth()).
			Height(lv.panelHeight()).
			Render(lv.model.View()),
	)
	return lipgloss.NewStyle().
		Width(totalWidth).
		Height(totalHeight).
		Render(content)
}

func (lv *listView) applySize() {
	panelWidth := lv.panelWidth()
	panelHeight := lv.panelHeight()
	listWidth := max(10, panelWidth-2)
	listHeight := max(5, panelHeight-2)
	lv.model.SetSize(listWidth, listHeight)
}

func (lv listView) panelWidth() int {
	return max(20, lv.totalWidth-2)
}

func (lv listView) panelHeight() int {
	return max(6, lv.totalHeight-3)
}

func (lv listView) SelectedSummary() (tuiapp.TableSummary, bool) {
	item, ok := lv.model.SelectedItem().(tableItem)
	if !ok {
		return tuiapp.TableSummary{}, false
	}
	return item.summary, true
}

func (lv listView) Filtering() bool {
	return lv.model.SettingFilter()
}

func (lv listView) Model() list.Model {
	return lv.model
}

type tableItem struct {
	summary tuiapp.TableSummary
}

func (t tableItem) Title() string {
	id := t.summary.TableIdentity
	if id == "" {
		id = t.summary.Identifier
	}
	return fmt.Sprintf("%s. %s", strings.ToUpper(id), t.summary.Name)
}

func (t tableItem) Description() string {
	if t.summary.Summary == "" {
		return ""
	}
	return truncate(t.summary.Summary, 90)
}

func (t tableItem) FilterValue() string {
	fields := []string{
		t.summary.Name,
		t.summary.Summary,
	}
	return strings.ToLower(strings.Join(fields, " "))
}

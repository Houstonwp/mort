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
	model list.Model
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

	return listView{model: l}
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

func (lv listView) View(totalWidth, totalHeight int) string {
	panelWidth := max(20, totalWidth-2)
	panelHeight := max(6, totalHeight-3)
	listWidth := max(10, panelWidth-2)
	listHeight := max(5, panelHeight-2)

	lv.model.SetSize(listWidth, listHeight)

	body := panelStyle.Width(panelWidth).Height(panelHeight).Render(lv.model.View())
	title := headerStyle.Width(panelWidth).Render("Mortality Tables")
	content := lipgloss.JoinVertical(lipgloss.Left, title, body)
	return lipgloss.NewStyle().
		Width(totalWidth).
		Height(totalHeight).
		Render(content)
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

func (t tableItem) FilterValue() string { return t.summary.Name }

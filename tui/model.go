package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"

	"mort/internal/tuiapp"
)

const summariesChunkSize = 150

const (
	tabClassification = iota
	tabMetadata
	tabRates
)

var tabLabels = []string{"Classification", "Metadata", "Rates"}

var (
	accentColor       = lipgloss.Color("99")
	subtleColor       = lipgloss.Color("245")
	warningColor      = lipgloss.Color("203")
	panelStyle        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accentColor).Padding(0, 1)
	ratesPanelStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accentColor).Padding(0, 1)
	warningPanelStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(warningColor).Padding(0, 1)
	helperTextStyle   = lipgloss.NewStyle().Foreground(subtleColor)
	tabActiveStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(accentColor).Bold(true).Padding(0, 2)
	tabInactiveStyle  = lipgloss.NewStyle().Foreground(subtleColor).Padding(0, 2)
	headerStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	valueStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	labelStyle        = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	sectionTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Underline(true)
	warningTextStyle  = lipgloss.NewStyle().Foreground(warningColor).Bold(true)
)

const maxWarningMessages = 50
const mortLogo = ` __  __  ___  ____ _____
|  \/  |/ _ \|  _ \_   _|
| |\/| | | | | |_) || |
| |  | | |_| |  _ < | |
|_|  |_|\___/|_| \_\|_|`

type state int

const (
	stateLoading state = iota
	stateList
	stateDetail
)

// Model represents the Bubble Tea program state.
type Model struct {
	state       state
	jsonDir     string
	width       int
	height      int
	list        list.Model
	summaries   []tuiapp.TableSummary
	filtered    []tuiapp.TableSummary
	warnings    []string
	detail      *tuiapp.TableDetail
	detailIndex int
	detailTab   int
	ratesTable  table.Model
	textView    viewport.Model
	textContent string
	err         error
}

// NewModel initializes a TUI model for the given JSON directory.
func NewModel(jsonDir string) Model {
	if jsonDir == "" {
		jsonDir = filepath.Join(".", "json")
	}

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

	model := Model{
		state:      stateLoading,
		jsonDir:    jsonDir,
		list:       l,
		width:      80,
		height:     24,
		ratesTable: rt,
		textView:   tv,
	}
	model.resizeRatesTable(model.detailInnerWidth())
	return model
}

// Init kicks off table loading.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadSummariesCmd(m.jsonDir, nil, 0),
		tea.WindowSize(),
	)
}

// Update handles Bubble Tea messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetSize(m.detailInnerWidth(), max(5, msg.Height-8))
		m.resizeRatesTable(m.detailInnerWidth())
	case summariesChunkMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if len(msg.warn) > 0 {
			m.addWarnings(msg.warn)
		}
		if len(msg.chunk) > 0 {
			m.summaries = append(m.summaries, msg.chunk...)
		}
		if msg.done {
			sortSummaries(m.summaries)
			m.applyFilter("")
		}
		if !msg.done {
			return m, loadSummariesCmd(m.jsonDir, msg.files, msg.next)
		}
		if m.state == stateLoading {
			m.state = stateList
		}
	case detailLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.detail = msg.detail
		m.detailIndex = 0
		m.detailTab = tabClassification
		m.state = stateDetail
		m.resizeRatesTable(m.detailInnerWidth())
		m.refreshRatesTable()
		m.resetTextView()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.state == stateDetail {
				m.state = stateList
				m.detail = nil
				m.ratesTable.Blur()
				return m, nil
			}
		case "tab":
			if m.state == stateDetail && m.detail != nil {
				m.detailTab = (m.detailTab + 1) % len(tabLabels)
				m.refreshRatesTable()
				m.resetTextView()
				return m, nil
			}
		case "shift+tab":
			if m.state == stateDetail && m.detail != nil {
				m.detailTab--
				if m.detailTab < 0 {
					m.detailTab = len(tabLabels) - 1
				}
				m.refreshRatesTable()
				m.resetTextView()
				return m, nil
			}
		case "1", "2", "3":
			if m.state == stateDetail && m.detail != nil {
				idx := int(msg.Runes[0] - '1')
				if idx >= 0 && idx < len(tabLabels) {
					m.detailTab = idx
					m.refreshRatesTable()
					m.resetTextView()
					return m, nil
				}
			}
			if m.state == stateList && len(m.filtered) > 0 {
				if item, ok := m.list.SelectedItem().(tableItem); ok {
					return m, loadDetailCmd(item.summary.FilePath)
				}
			}
		case "left", "h":
			if m.state == stateDetail && m.detail != nil && m.detailIndex > 0 {
				m.detailIndex--
				m.refreshRatesTable()
				m.resetTextView()
				return m, nil
			}
		case "right", "l":
			if m.state == stateDetail && m.detail != nil && m.detailIndex < len(m.detail.Tables)-1 {
				m.detailIndex++
				m.refreshRatesTable()
				m.resetTextView()
				return m, nil
			}
		}

		if m.state == stateDetail && m.detail != nil && m.detailTab == tabRates {
			var cmd tea.Cmd
			m.ratesTable, cmd = m.ratesTable.Update(msg)
			return m, cmd
		} else if m.state == stateDetail && m.detail != nil && m.detailTab != tabRates {
			var cmd tea.Cmd
			m.textView, cmd = m.textView.Update(msg)
			return m, cmd
		}
	}

	if m.state == stateList {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the UI.
func (m Model) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("error: %v\n", m.err))
	}
	switch m.state {
	case stateLoading:
		logo := lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render(mortLogo)
		subtitle := helperTextStyle.Render("loading mortality tables…")
		content := lipgloss.JoinVertical(lipgloss.Center, logo, subtitle)
		return lipgloss.Place(
			max(1, m.width),
			max(1, m.height),
			lipgloss.Center,
			lipgloss.Center,
			content,
		)
	case stateDetail:
		return m.detailView()
	default:
		return m.listView()
	}
}

func (m Model) listView() string {
	contentWidth := m.detailContentWidth()
	availableHeight := max(6, m.height-4)
	warning := m.warningView(contentWidth)
	listHeight := max(5, availableHeight)
	m.list.SetSize(m.detailInnerWidth(), listHeight)
	bodySections := []string{m.list.View()}
	if warning != "" {
		bodySections = append(bodySections, warning)
	}
	body := panelStyle.Width(contentWidth).Render(
		lipgloss.JoinVertical(lipgloss.Left, bodySections...),
	)

	title := headerStyle.Width(contentWidth).Render("Mortality Tables")
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		body,
	)
	return lipgloss.NewStyle().
		Width(max(1, m.width)).
		Height(max(1, m.height)).
		Render(content)
}

func (m Model) detailView() string {
	if m.detail == nil {
		return "Loading table detail…"
	}
	title := headerStyle.Render(m.detail.Classification.TableName)
	subtitle := helperTextStyle.Render(fmt.Sprintf("%s • Table %d of %d • Version %s",
		m.detail.Classification.ProviderName, m.detailIndex+1, len(m.detail.Tables), m.detail.Version))
	contentWidth := m.detailContentWidth()
	info := lipgloss.NewStyle().Width(contentWidth).Render(
		lipgloss.JoinVertical(lipgloss.Left, title, subtitle),
	)

	tabs := lipgloss.NewStyle().Width(contentWidth).Render(renderTabs(m.detailTab))

	footer := helperTextStyle.Width(contentWidth).Render("←/→ table • Tab switch panel • j/k scroll rates • esc back")

	bodyHeight := m.availableBodyHeight(info, tabs, footer)
	bodyPanel := m.renderBodyPanel(contentWidth, bodyHeight)

	warningLine := m.warningView(contentWidth)
	var content string
	if warningLine != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, info, tabs, bodyPanel, warningLine, footer)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left, info, tabs, bodyPanel, footer)
	}
	return lipgloss.NewStyle().
		Width(max(1, m.width)).
		Height(max(1, m.height)).
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

func (m *Model) addWarnings(entries []string) {
	if len(entries) == 0 {
		return
	}
	m.warnings = append(m.warnings, entries...)
	if len(m.warnings) > maxWarningMessages {
		m.warnings = append([]string(nil), m.warnings[len(m.warnings)-maxWarningMessages:]...)
	}
}

func (m Model) warningView(width int) string {
	if len(m.warnings) == 0 {
		return ""
	}
	previewCount := min(len(m.warnings), 3)
	start := len(m.warnings) - previewCount
	preview := append([]string(nil), m.warnings[start:]...)
	if len(m.warnings) > previewCount {
		preview = append(preview, fmt.Sprintf("+%d more warning(s)", len(m.warnings)-previewCount))
	}
	panelWidth := max(10, width-4)
	text := warningTextStyle.Render(strings.Join(preview, " • "))
	return warningPanelStyle.Width(panelWidth).Render(text)
}

func (m *Model) applyFilter(query string) {
	if len(m.summaries) == 0 {
		m.filtered = nil
		m.list.SetItems(nil)
		return
	}
	if strings.TrimSpace(query) == "" {
		m.filtered = append([]tuiapp.TableSummary(nil), m.summaries...)
	} else {
		filtered := tuiapp.FilterSummaries(m.summaries, query)
		m.filtered = filtered
	}
	items := make([]list.Item, len(m.filtered))
	for i, summary := range m.filtered {
		items[i] = tableItem{summary: summary}
	}
	m.list.SetItems(items)
	if len(items) > 0 {
		m.list.Select(0)
	}
}

func (m *Model) refreshRatesTable() {
	if m.detail == nil || len(m.detail.Tables) == 0 {
		m.ratesTable.SetRows(nil)
		m.ratesTable.Blur()
		return
	}
	if m.detailIndex >= len(m.detail.Tables) {
		m.detailIndex = len(m.detail.Tables) - 1
	}
	tableData := m.detail.Tables[m.detailIndex]
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
	m.ratesTable.SetRows(rows)
	if m.detailTab == tabRates {
		m.ratesTable.Focus()
	} else {
		m.ratesTable.Blur()
	}
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

type detailLoadedMsg struct {
	detail *tuiapp.TableDetail
	err    error
}

func loadDetailCmd(path string) tea.Cmd {
	return func() tea.Msg {
		detail, err := tuiapp.LoadTableDetail(path)
		return detailLoadedMsg{detail: detail, err: err}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *Model) resizeRatesTable(width int) {
	if width <= 0 {
		width = 40
	}
	rateWidth := max(10, width-16)
	columns := []table.Column{
		{Title: "Age", Width: 8},
		{Title: "Dur", Width: 6},
		{Title: "Rate", Width: rateWidth},
	}
	m.ratesTable.SetColumns(columns)
}

func limitLines(s string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[:maxLines], "\n")
}

func appendKV(b *strings.Builder, label, value string, width int) {
	if value == "" {
		value = "—"
	}
	wrapWidth := max(10, width-lipgloss.Width(label)-4)
	lines := strings.Split(wordwrap.String(value, wrapWidth), "\n")
	b.WriteString(labelStyle.Render(label + ": "))
	b.WriteString(valueStyle.Render(lines[0]))
	b.WriteString("\n")
	indent := strings.Repeat(" ", lipgloss.Width(label)+2)
	for _, line := range lines[1:] {
		b.WriteString(indent)
		b.WriteString(valueStyle.Render(line))
		b.WriteString("\n")
	}
}

func appendBullet(b *strings.Builder, text string, width int) {
	if width <= 0 {
		width = 40
	}
	lines := strings.Split(wordwrap.String(text, max(10, width-4)), "\n")
	for i, line := range lines {
		prefix := "• "
		if i > 0 {
			prefix = "  "
		}
		b.WriteString(prefix)
		b.WriteString(valueStyle.Render(line))
		b.WriteString("\n")
	}
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-1]) + "…"
}

func sortSummaries(list []tuiapp.TableSummary) {
	sort.SliceStable(list, func(i, j int) bool {
		ai, aHas := parseIdentity(list[i].TableIdentity)
		bi, bHas := parseIdentity(list[j].TableIdentity)
		switch {
		case aHas && bHas && ai != bi:
			return ai < bi
		case aHas && !bHas:
			return true
		case !aHas && bHas:
			return false
		default:
			if list[i].TableIdentity != list[j].TableIdentity {
				return list[i].TableIdentity < list[j].TableIdentity
			}
			return list[i].Name < list[j].Name
		}
	})
}

func parseIdentity(id string) (int, bool) {
	val, err := strconv.Atoi(id)
	if err != nil {
		return 0, false
	}
	return val, true
}

func (m *Model) renderScrollablePanel(width, bodyHeight int, content string) string {
	innerWidth := max(10, width-4)
	viewHeight := max(1, bodyHeight-2)
	if m.textView.Width != innerWidth {
		m.textView.Width = innerWidth
	}
	if m.textView.Height != viewHeight {
		m.textView.Height = viewHeight
	}
	wrapped := wordwrap.String(content, innerWidth)
	if wrapped != m.textContent {
		m.textContent = wrapped
		m.textView.SetContent(wrapped)
		m.textView.GotoTop()
	}
	return panelStyle.Width(width).Height(bodyHeight).Render(m.textView.View())
}

func (m *Model) renderRatesPanel(width, bodyHeight int) string {
	innerWidth := max(10, width-4)
	m.ratesTable.SetWidth(innerWidth)
	m.ratesTable.SetHeight(max(1, bodyHeight-2))
	return ratesPanelStyle.Width(width).Height(bodyHeight).Render(m.ratesTable.View())
}

func (m *Model) resetTextView() {
	m.textContent = ""
	m.textView.SetYOffset(0)
}

type summariesChunkMsg struct {
	chunk []tuiapp.TableSummary
	files []string
	next  int
	done  bool
	warn  []string
	err   error
}

func loadSummariesCmd(dir string, files []string, start int) tea.Cmd {
	return func() tea.Msg {
		var names []string
		if files == nil {
			entries, err := os.ReadDir(dir)
			if err != nil {
				return summariesChunkMsg{err: fmt.Errorf("read json dir: %w", err)}
			}
			for _, entry := range entries {
				if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
					continue
				}
				names = append(names, entry.Name())
			}
			sort.Strings(names)
		} else {
			names = files
		}

		if len(names) == 0 {
			return summariesChunkMsg{files: names, done: true}
		}

		if start >= len(names) {
			return summariesChunkMsg{files: names, next: len(names), done: true}
		}

		end := start + summariesChunkSize
		if end > len(names) {
			end = len(names)
		}

		chunk := make([]tuiapp.TableSummary, 0, end-start)
		warnings := make([]string, 0)
		for _, name := range names[start:end] {
			path := filepath.Join(dir, name)
			summary, err := tuiapp.LoadTableSummary(path)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("%s: %v", name, err))
				continue
			}
			chunk = append(chunk, *summary)
		}

		done := end >= len(names)
		return summariesChunkMsg{
			chunk: chunk,
			files: names,
			next:  end,
			done:  done,
			warn:  warnings,
		}
	}
}
func (m Model) detailContentWidth() int {
	return max(20, m.width-4)
}

func (m Model) detailInnerWidth() int {
	return max(10, m.detailContentWidth()-2)
}

func (m Model) availableBodyHeight(header, tabs, footer string) int {
	total := m.height
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

func (m *Model) renderBodyPanel(width, bodyHeight int) string {
	switch m.detailTab {
	case tabRates:
		return m.renderRatesPanel(width, bodyHeight)
	case tabMetadata:
		return m.renderScrollablePanel(width, bodyHeight, renderMetadata(m.detail, m.detailIndex, width-6))
	default:
		return m.renderScrollablePanel(width, bodyHeight, renderClassification(m.detail, width-6))
	}
}

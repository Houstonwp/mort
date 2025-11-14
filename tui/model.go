package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

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
	panelStyle        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accentColor).Padding(0, 1)
	ratesPanelStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accentColor).Padding(0, 1)
	helperTextStyle   = lipgloss.NewStyle().Foreground(subtleColor)
	tabActiveStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(accentColor).Bold(true).Padding(0, 2)
	tabInactiveStyle  = lipgloss.NewStyle().Foreground(subtleColor).Padding(0, 2)
	headerStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	valueStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	labelStyle        = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	sectionTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Underline(true)
)

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
	state     state
	jsonDir   string
	width     int
	height    int
	list      listView
	summaries []tuiapp.TableSummary
	detail    detailView
	err       error
}

// NewModel initializes a TUI model for the given JSON directory.
func NewModel(jsonDir string) Model {
	if jsonDir == "" {
		jsonDir = filepath.Join(".", "json")
	}

	model := Model{
		state:   stateLoading,
		jsonDir: jsonDir,
		width:   80,
		height:  24,
		list:    newListView(),
		detail:  newDetailView(),
	}
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
	case summariesChunkMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if len(msg.chunk) > 0 {
			m.summaries = append(m.summaries, msg.chunk...)
		}
		if !msg.done {
			return m, loadSummariesCmd(m.jsonDir, msg.files, msg.next)
		}
		sortSummaries(m.summaries)
		m.list.SetItems(m.summaries)
		if m.state == stateLoading {
			m.state = stateList
		}
	case detailLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.detail.SetDetail(msg.detail)
		m.state = stateDetail
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.state == stateDetail {
				m.state = stateList
				return m, nil
			}
		case "tab":
			if m.state == stateDetail {
				m.detail.NextTab()
				return m, nil
			}
		case "shift+tab":
			if m.state == stateDetail {
				m.detail.PrevTab()
				return m, nil
			}
		case "enter":
			if m.state == stateList && !m.list.Filtering() {
				if summary, ok := m.list.SelectedSummary(); ok {
					return m, loadDetailCmd(summary.FilePath)
				}
			}
		case "1", "2", "3":
			if m.state == stateDetail {
				idx := int(msg.Runes[0] - '1')
				m.detail.SetTab(idx)
				return m, nil
			}
		case "left", "h":
			if m.state == stateDetail {
				m.detail.PrevTable()
				return m, nil
			}
		case "right", "l":
			if m.state == stateDetail {
				m.detail.NextTable()
				return m, nil
			}
		}
	}

	if m.state == stateDetail {
		return m, m.detail.Update(msg)
	}
	if m.state == stateList {
		return m, m.list.Update(msg)
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
	return m.list.View(m.width, m.height)
}

func (m Model) detailView() string {
	return m.detail.View(m.detailContentWidth(), m.height)
}

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

type summariesChunkMsg struct {
	chunk []tuiapp.TableSummary
	files []string
	next  int
	done  bool
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
		for _, name := range names[start:end] {
			path := filepath.Join(dir, name)
			summary, err := tuiapp.LoadTableSummary(path)
			if err != nil {
				return summariesChunkMsg{err: err}
			}
			chunk = append(chunk, *summary)
		}

		done := end >= len(names)
		return summariesChunkMsg{
			chunk: chunk,
			files: names,
			next:  end,
			done:  done,
		}
	}
}
func (m Model) detailContentWidth() int {
	return max(20, m.width-4)
}

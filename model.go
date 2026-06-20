package main

import (
	"bytes"
	"dmarc-tui/internal/dmarc"
	"dmarc-tui/internal/ipinfo"
	"net"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type filterMode int

const (
	filterAll filterMode = iota
	filterPass
	filterReview
	filterFail
)

type sortMode int

const (
	sortVolume sortMode = iota
	sortIP
)

type detailTabID int

const (
	tabEvaluation detailTabID = iota
	tabIdentifiers
	tabPolicy
	tabRawXML
	tabIPInfo
)

// ── Model ─────────────────────────────────────────────────────────────────────

type model struct {
	feedbacks   []*dmarc.Feedback
	fileIdx     int
	filtered    []dmarc.Record
	filter      filterMode
	sort        sortMode
	searching   bool
	inspecting  bool
	detailTab   detailTabID
	helpVisible bool
	list        list.Model
	textInput   textinput.Model
	help        help.Model
	detailVP    viewport.Model
	helpVP      viewport.Model
	ipCache     map[string]ipinfo.CacheEntry
	width       int
	height      int
	ready       bool
}

// minTermWidth / minTermHeight are the smallest dimensions the app can render
// meaningfully. Below these the table math overflows and text truncates badly;
// View() shows a "resize terminal" message instead.
const (
	minTermWidth  = 100
	minTermHeight = 20
)

func (m model) contentWidth() int {
	return m.width
}

// listGeometry returns the shared geometry values used by both rebuildContent
// (to size the list/delegate) and renderRecordsTable (to draw borders/headers).
func (m model) listGeometry() (tableInnerW, wSrc, wProv, listH int) {
	W := m.contentWidth()
	tableInnerW = W - 5
	contentW := tableInnerW - 3

	wProv = max(contentW*provWidthPct/100, 12)
	if wProv > 40 {
		wProv = 40 // cap: domain names beyond 40 chars are truncated anyway
	}
	fixedW := colWidthVol + colWidthAuth + wProv + classBadgeW + 2*4
	wSrc = max(contentW-fixedW, 14)
	if wSrc > 40 {
		wSrc = 40 // cap: longest realistic IP (IPv6) fits in 40
	}
	// Fixed rows outside the list: topBar(1) + blank(1) + statCards(5) +
	// blank(1) + filterBar(1) + blank(1) = 10 above the table; then table
	// chrome adds topBorder(1) + header(1) + blank(1) = 3 more = 13 total.
	// Post-list chrome: bottomBorder(1) + trailingBlank(1) + statusBar(1) = 3.
	// The two vpH constants below have compensating offsets (11 vs 5 instead
	// of 13 vs 3) that cancel correctly: listH = m.height - 16 either way.
	vpH := max(m.height-11, 3)
	listH = max(vpH-5, 1)
	return
}

func newModel(feedbacks []*dmarc.Feedback) model {
	d := recordDelegate{}
	l := list.New([]list.Item{}, d, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	ti := textinput.New()
	ti.Prompt = ""
	ti.TextStyle = styleWhite
	ti.Cursor.Style = styleWhiteBold

	h := help.New()
	h.Styles.ShortKey = styleWhiteBold
	h.Styles.ShortDesc = styleMuted
	h.Styles.ShortSeparator = styleMuted
	h.Styles.Ellipsis = styleMuted

	m := model{
		feedbacks: feedbacks,
		list:      l,
		textInput: ti,
		help:      h,
		detailVP:  viewport.New(0, 0),
		helpVP:    viewport.New(0, 0),
		ipCache:   make(map[string]ipinfo.CacheEntry),
	}
	m.loadFile(0)
	return m
}

func (m *model) loadFile(idx int) {
	m.fileIdx = idx
	m.applyFilterSort()
	m.list.Select(0)
}

func (m *model) applyFilterSort() tea.Cmd {
	q := strings.ToLower(strings.TrimSpace(m.textInput.Value()))

	var out []dmarc.Record
	for _, r := range m.feedbacks[m.fileIdx].Records {
		switch m.filter {
		case filterPass:
			if dmarc.Classify(r) != dmarc.ClassPass {
				continue
			}
		case filterReview:
			if dmarc.Classify(r) != dmarc.ClassNeedsReview {
				continue
			}
		case filterFail:
			if dmarc.Classify(r) != dmarc.ClassFail {
				continue
			}
		}
		if q != "" && !dmarc.RecordMatches(r, q) {
			continue
		}
		out = append(out, r)
	}

	switch m.sort {
	case sortVolume:
		sort.SliceStable(out, func(i, j int) bool {
			return out[i].Row.Count > out[j].Row.Count
		})
	case sortIP:
		sort.SliceStable(out, func(i, j int) bool {
			a := net.ParseIP(out[i].Row.SourceIP)
			b := net.ParseIP(out[j].Row.SourceIP)
			if a == nil || b == nil {
				return out[i].Row.SourceIP < out[j].Row.SourceIP
			}
			// Compare numerically (byte-wise on the parsed address) — comparing
			// a.String()/b.String() lexicographically would put "203.0.113.10"
			// before "203.0.113.2", since '1' < '2' as characters.
			return bytes.Compare(a, b) < 0
		})
	}

	m.filtered = out

	items := make([]list.Item, len(m.filtered))
	for i, r := range m.filtered {
		items[i] = recordItem{record: r}
	}
	cmd := m.list.SetItems(items)
	if n := len(m.filtered); n > 0 && m.list.Index() >= n {
		m.list.Select(n - 1)
	}
	return cmd
}

// maybeStartIPFetch dispatches a fetchIPInfo command if the IPInfo tab is
// active and the current record's IP isn't already in the cache. It marks the
// entry as in-flight (done=false) before returning so subsequent calls for the
// same IP are no-ops while the request is pending.
func (m *model) maybeStartIPFetch() tea.Cmd {
	if m.detailTab != tabIPInfo || !m.inspecting || len(m.filtered) == 0 {
		return nil
	}
	ip := m.filtered[m.list.Index()].Row.SourceIP
	if _, ok := m.ipCache[ip]; ok {
		return nil
	}
	m.ipCache[ip] = ipinfo.CacheEntry{}
	return ipinfo.Fetch(ip)
}

// ── Bubble Tea interface ──────────────────────────────────────────────────────

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.rebuildContent()

	case ipinfo.Msg:
		m.ipCache[msg.IP] = ipinfo.CacheEntry{Done: true, Info: msg.Info, Err: msg.Err}
		// Refresh the viewport only when the result is for the currently visible record.
		if m.inspecting && m.detailTab == tabIPInfo && len(m.filtered) > 0 &&
			m.filtered[m.list.Index()].Row.SourceIP == msg.IP {
			m.refreshDetailVP()
		}

	case tea.KeyMsg:
		// The help overlay is modal — it swallows every key except the ones
		// that close it, regardless of what screen it was opened over.
		if m.helpVisible {
			switch msg.String() {
			case "esc", "q", "?":
				m.helpVisible = false
			case "up", "k":
				m.helpVP.LineUp(1)
			case "down", "j":
				m.helpVP.LineDown(1)
			case "pgup":
				m.helpVP.HalfViewUp()
			case "pgdown":
				m.helpVP.HalfViewDown()
			}
			return m, nil
		}

		if m.inspecting {
			var fetchCmd tea.Cmd
			switch msg.String() {
			case "esc", "q":
				m.inspecting = false
			case "?":
				m.helpVisible = true
				m.helpVP.GotoTop()
				m.refreshHelpVP()
			case "tab", "right":
				if m.detailTab < tabIPInfo {
					m.detailTab++
				} else {
					m.detailTab = tabEvaluation
				}
				m.detailVP.GotoTop()
				m.refreshDetailVP()
				fetchCmd = m.maybeStartIPFetch()
			case "left":
				if m.detailTab > tabEvaluation {
					m.detailTab--
				} else {
					m.detailTab = tabIPInfo
				}
				m.detailVP.GotoTop()
				m.refreshDetailVP()
				fetchCmd = m.maybeStartIPFetch()
			case "1", "2", "3", "4", "5":
				m.detailTab = detailTabID(msg.String()[0] - '1')
				m.detailVP.GotoTop()
				m.refreshDetailVP()
				fetchCmd = m.maybeStartIPFetch()
			case "up", "k":
				if m.list.Index() > 0 {
					m.list.Select(m.list.Index() - 1)
					m.detailVP.GotoTop()
					m.refreshDetailVP()
					fetchCmd = m.maybeStartIPFetch()
				}
			case "down", "j":
				if m.list.Index() < len(m.list.Items())-1 {
					m.list.Select(m.list.Index() + 1)
					m.detailVP.GotoTop()
					m.refreshDetailVP()
					fetchCmd = m.maybeStartIPFetch()
				}
			case "r":
				// Retry an IPInfo fetch — clears the cache entry so
				// maybeStartIPFetch treats it as unseen and re-dispatches.
				if m.detailTab == tabIPInfo && len(m.filtered) > 0 {
					ip := m.filtered[m.list.Index()].Row.SourceIP
					delete(m.ipCache, ip)
					m.refreshDetailVP()
					fetchCmd = m.maybeStartIPFetch()
				}
			case "pgup":
				m.detailVP.HalfViewUp()
			case "pgdown":
				m.detailVP.HalfViewDown()
			}
			return m, fetchCmd
		}

		if m.searching {
			switch msg.Type {
			case tea.KeyEsc:
				m.searching = false
				m.textInput.Blur()
				m.textInput.SetValue("")
				cmds = append(cmds, m.applyFilterSort())
				m.list.Select(0)
			case tea.KeyEnter:
				m.searching = false
				m.textInput.Blur()
			default:
				prev := m.textInput.Value()
				var tiCmd tea.Cmd
				m.textInput, tiCmd = m.textInput.Update(msg)
				cmds = append(cmds, tiCmd)
				if m.textInput.Value() != prev {
					cmds = append(cmds, m.applyFilterSort())
					m.list.Select(0)
				}
			}
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if len(m.filtered) > 0 {
				m.inspecting = true
				m.detailTab = tabEvaluation
				m.detailVP.GotoTop()
				m.refreshDetailVP()
			}
		case "j", "down":
			if m.list.Index() < len(m.list.Items())-1 {
				m.list.Select(m.list.Index() + 1)
			}
		case "k", "up":
			if m.list.Index() > 0 {
				m.list.Select(m.list.Index() - 1)
			}
		case "g":
			m.list.Select(0)
		case "G":
			if n := len(m.list.Items()); n > 0 {
				m.list.Select(n - 1)
			}
		case "/":
			m.searching = true
			cmds = append(cmds, m.textInput.Focus())
		case "?":
			m.helpVisible = true
			m.helpVP.GotoTop()
			m.refreshHelpVP()
		case "f":
			m.filter = (m.filter + 1) % 4
			m.list.Select(0)
			cmds = append(cmds, m.applyFilterSort())
		case "s":
			m.sort = (m.sort + 1) % 2
			cmds = append(cmds, m.applyFilterSort())
		case "]", "right":
			if m.fileIdx < len(m.feedbacks)-1 {
				m.loadFile(m.fileIdx + 1)
				m.rebuildContent()
			}
		case "[", "left":
			if m.fileIdx > 0 {
				m.loadFile(m.fileIdx - 1)
				m.rebuildContent()
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) rebuildContent() {
	if m.width == 0 || m.height == 0 {
		return
	}
	tableInnerW, wSrc, wProv, listH := m.listGeometry()

	fb := m.feedbacks[m.fileIdx]
	var totalMsgs int
	for _, r := range fb.Records {
		totalMsgs += r.Row.Count
	}

	d := recordDelegate{
		wSrc:        wSrc,
		wProv:       wProv,
		tableInnerW: tableInnerW,
		totalMsgs:   totalMsgs,
	}
	m.list.SetDelegate(d)
	m.list.SetSize(tableInnerW-1, listH)

	W := m.contentWidth()
	// 3 fixed rows outside the detail/help viewport:
	//   header(1) + blank(1) + statusBar(1) = 3
	vpH := max(m.height-3, 1)
	m.detailVP.Width = W
	m.detailVP.Height = vpH
	m.helpVP.Width = W
	m.helpVP.Height = vpH
	if m.inspecting && len(m.filtered) > 0 {
		m.refreshDetailVP()
	}
	if m.helpVisible {
		m.refreshHelpVP()
	}
}

func (m *model) refreshDetailVP() {
	if m.list.Index() >= len(m.filtered) {
		return
	}
	r := m.filtered[m.list.Index()]
	W := m.contentWidth()
	content := padViewportContent(m.renderDetailBody(r, W), W, m.detailVP.Height, colAppBg)
	m.detailVP.SetContent(content)
}

func (m *model) refreshHelpVP() {
	W := m.contentWidth()
	content := padViewportContent(m.renderHelpBody(W), W, m.helpVP.Height, colAppBg)
	m.helpVP.SetContent(content)
}

func (m model) passFailCounts() (pass, fail int) {
	for _, r := range m.feedbacks[m.fileIdx].Records {
		if dmarc.Classify(r) == dmarc.ClassFail {
			fail++
		} else {
			pass++
		}
	}
	return
}

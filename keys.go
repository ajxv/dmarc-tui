package main

import (
	"github.com/charmbracelet/bubbles/key"
)

// ── List view key bindings ────────────────────────────────────────────────────

type listKeyMap struct {
	Up, Down, Top, Bottom key.Binding
	Enter                 key.Binding
	Filter, Sort, Search  key.Binding
	PrevFile, NextFile    key.Binding
	Help, Quit            key.Binding
}

func (k listKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Filter, k.Sort, k.Search, k.Help, k.Quit}
}
func (k listKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Top, k.Bottom, k.Enter},
		{k.Filter, k.Sort, k.Search},
		{k.PrevFile, k.NextFile, k.Help, k.Quit},
	}
}

var listKeys = listKeyMap{
	Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Top:      key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "top")),
	Bottom:   key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),
	Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
	Filter:   key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter")),
	Sort:     key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "sort")),
	Search:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	PrevFile: key.NewBinding(key.WithKeys("[", "left"), key.WithHelp("[/←", "prev")),
	NextFile: key.NewBinding(key.WithKeys("]", "right"), key.WithHelp("]/→", "next")),
	Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

// ── Detail view key bindings ──────────────────────────────────────────────────

type detailKeyMap struct {
	Back, Tab, Jump        key.Binding
	PrevRecord, NextRecord key.Binding
	ScrollUp, ScrollDown   key.Binding
	Retry                  key.Binding
	Help, Quit             key.Binding
}

func (k detailKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Back, k.Tab, k.PrevRecord, k.NextRecord, k.ScrollUp, k.Retry, k.Help, k.Quit}
}
func (k detailKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Back, k.Tab, k.Jump},
		{k.PrevRecord, k.NextRecord, k.ScrollUp, k.ScrollDown},
		{k.Retry, k.Help, k.Quit},
	}
}

var detailKeys = detailKeyMap{
	Back:       key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Tab:        key.NewBinding(key.WithKeys("tab", "left", "right"), key.WithHelp("tab · ←→", "section")),
	Jump:       key.NewBinding(key.WithKeys("1", "2", "3", "4", "5"), key.WithHelp("1-5", "jump")),
	PrevRecord: key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "prev")),
	NextRecord: key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "next")),
	ScrollUp:   key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "scroll ↑")),
	ScrollDown: key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "scroll ↓")),
	Retry:      key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
	Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:       key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
}

// ── Help overlay key bindings ─────────────────────────────────────────────────

type helpKeyMap struct {
	ScrollUp, ScrollDown key.Binding
	Close                key.Binding
	Quit                 key.Binding
}

func (k helpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ScrollUp, k.ScrollDown, k.Close, k.Quit}
}
func (k helpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.ScrollUp, k.ScrollDown, k.Close, k.Quit}}
}

var helpKeys = helpKeyMap{
	ScrollUp:   key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	ScrollDown: key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Close:      key.NewBinding(key.WithKeys("esc", "?"), key.WithHelp("esc/?", "close")),
	Quit:       key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
}

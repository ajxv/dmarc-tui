package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── Dracula palette ───────────────────────────────────────────────────────────
// Every hex literal in the project lives here.
// No other file should contain a raw "#rrggbb" string — reference these vars.

var (
	// Backgrounds
	colAppBg       = lipgloss.Color("#282a36") // app background (Dracula bg, used everywhere)
	colSelectionBg = lipgloss.Color("#44475a") // selected-row highlight (current line)
	colBadgeBg     = lipgloss.Color("#21222c") // domain-via badge background (recessed)

	// Named foreground palette
	colBrand    = lipgloss.Color("#bd93f9") // brand purple (logo)
	colRed      = lipgloss.Color("#ff5555") // red
	colGreen    = lipgloss.Color("#50fa7b") // green
	colYellow   = lipgloss.Color("#f1fa8c") // yellow
	colAccent   = lipgloss.Color("#82aaff") // blue accent / active tab
	colCyan     = lipgloss.Color("#8be9fd") // cyan
	colFg       = lipgloss.Color("#f8f8f2") // primary text
	colMuted    = lipgloss.Color("#6272a4") // muted / comment
	colDimMuted = lipgloss.Color("#3a3c4e") // dim borders

	// Evaluation-specific: brighter variants for the detail result table.
	// Distinct from the softer badge palette to carry more weight on a plain
	// background, making the pass/fail verdict read clearly at a glance.
	colPassBright = lipgloss.Color("#50fa7b") // bright green
	colFailBright = lipgloss.Color("#ff5555") // bright red
)

// ── Column-layout constants ───────────────────────────────────────────────────
// Shared between recordDelegate.Render (model.go) and renderRecordsTable
// (view.go). Both must agree or headers and rows silently misalign.

const (
	colWidthVol  = 16
	colWidthAuth = 28
	classBadgeW  = 12 // fixed visual width for all class badges, incl. "[ ]" — sized to fit "review"
	volBarWidth  = 8  // character width of the volume bar glyph in each record row

	// Layout ratios (integer %; used in integer arithmetic so documented here to
	// make the truncation visible rather than burying 28 and 14 in call sites).
	barsWidthPct = 28 // % of stat-card area for the pass-rate bars panel
	provWidthPct = 14 // % of table content width for the authenticated-domain column
)

// ── Background helpers ────────────────────────────────────────────────────────

// bgCodeCache memoizes the ANSI background escape lipgloss emits for a color.
var bgCodeCache = map[lipgloss.Color]string{}

// bgEscape returns the exact ANSI background-color escape sequence lipgloss
// itself emits for bg. We can't derive this from the hex string with simple
// int parsing — lipgloss round-trips the color through go-colorful's
// float64 RGB representation, which can shift a channel by ±1 from the literal
// hex value (e.g. #1f2335 → 31;35;52 instead of 31;35;53). Two backgrounds
// that are "the same" color but encoded via different paths render as a
// visible seam, so we always source the escape from lipgloss to guarantee a
// byte-for-byte match with every other `Background(bg)` in the view.
func bgEscape(bg lipgloss.Color) string {
	if code, ok := bgCodeCache[bg]; ok {
		return code
	}
	rendered := lipgloss.NewStyle().Background(bg).Render("X")
	code := rendered
	if i := strings.IndexByte(rendered, 'X'); i > 0 {
		code = rendered[:i]
	}
	bgCodeCache[bg] = code
	return code
}

// fillLine pads s to width and paints the full row with bg.
func fillLine(s string, width int, bg lipgloss.Color) string {
	bgCode := bgEscape(bg)
	padded := padRight(s, width)
	padded = strings.ReplaceAll(padded, "\x1b[0m", "\x1b[0m"+bgCode)
	padded = strings.ReplaceAll(padded, "\x1b[m", "\x1b[m"+bgCode)
	padded = strings.ReplaceAll(padded, "\x1b[49m", "\x1b[49m"+bgCode)
	return bgCode + padded + "\x1b[0m"
}

// padViewportContent ensures the content has at least h lines, each filled to
// width W with bg, so the viewport never exposes the raw terminal background
// in its empty-row area.
func padViewportContent(content string, W, h int, bg lipgloss.Color) string {
	lines := strings.Split(content, "\n")
	for len(lines) < h {
		lines = append(lines, fillLine("", W, bg))
	}
	return strings.Join(lines, "\n")
}

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	styleRed      = lipgloss.NewStyle().Foreground(colRed)
	styleGreen    = lipgloss.NewStyle().Foreground(colGreen)
	styleYellow   = lipgloss.NewStyle().Foreground(colYellow)
	styleAccent   = lipgloss.NewStyle().Foreground(colAccent)
	styleCyan     = lipgloss.NewStyle().Foreground(colCyan)
	styleWhite    = lipgloss.NewStyle().Foreground(colFg)
	styleMuted    = lipgloss.NewStyle().Foreground(colMuted)
	styleDimMuted = lipgloss.NewStyle().Foreground(colDimMuted)
	styleBorder   = lipgloss.NewStyle().Foreground(colDimMuted)

	styleRedBold    = lipgloss.NewStyle().Foreground(colRed).Bold(true)
	styleGreenBold  = lipgloss.NewStyle().Foreground(colGreen).Bold(true)
	styleYellowBold = lipgloss.NewStyle().Foreground(colYellow).Bold(true)
	styleAccentBold = lipgloss.NewStyle().Foreground(colAccent).Bold(true)
	styleWhiteBold  = lipgloss.NewStyle().Foreground(colFg).Bold(true)
	styleCyanBold   = lipgloss.NewStyle().Foreground(colCyan).Bold(true)
)

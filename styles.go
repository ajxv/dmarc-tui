package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── Tokyo Night palette ───────────────────────────────────────────────────────
// Every hex literal in the project lives here.
// No other file should contain a raw "#rrggbb" string — reference these vars.

var (
	// Backgrounds
	colAppBg       = lipgloss.Color("#1a1b26") // app background (used everywhere)
	colSelectionBg = lipgloss.Color("#1e2030") // selected-row highlight
	colBadgeBg     = lipgloss.Color("#1e2335") // domain-via badge background

	// Named foreground palette
	colBrand    = lipgloss.Color("#7c6af7") // brand purple (logo)
	colRed      = lipgloss.Color("#f7768e")
	colGreen    = lipgloss.Color("#9ece6a")
	colYellow   = lipgloss.Color("#e0af68")
	colAccent   = lipgloss.Color("#7aa2f7") // blue accent / active tab
	colCyan     = lipgloss.Color("#7dcfff")
	colFg       = lipgloss.Color("#c0caf5") // primary text
	colMuted    = lipgloss.Color("#565f89")
	colDimMuted = lipgloss.Color("#3b4261")

	// Evaluation-specific: brighter variants for the detail result table.
	// Distinct from the softer badge palette to carry more weight on a plain
	// background, making the pass/fail verdict read clearly at a glance.
	colPassBright = lipgloss.Color("#4ade80")
	colFailBright = lipgloss.Color("#f87171")
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

package ui

import "github.com/charmbracelet/lipgloss"

// ── Palette ────────────────────────────────────────────────────────────────
// Clean minimal palette inspired by 1Password / Bitwarden.
var (
	// Neutrals
	colBackground = lipgloss.Color("#0f1117")
	colSurface    = lipgloss.Color("#1a1d27")
	colBorder     = lipgloss.Color("#2e3245")
	colBorderFocus = lipgloss.Color("#4a7cf0")

	// Text
	colText        = lipgloss.Color("#e8eaf0")
	colTextMuted   = lipgloss.Color("#6b7280")
	colTextSubtle  = lipgloss.Color("#374151")

	// Accent
	colAccent      = lipgloss.Color("#4a7cf0")
	colAccentDim   = lipgloss.Color("#2d4d9a")

	// Semantic
	colSuccess = lipgloss.Color("#34d399")
	colWarning = lipgloss.Color("#fbbf24")
	colDanger  = lipgloss.Color("#f87171")
	colInfo    = lipgloss.Color("#60a5fa")
)

// ── Base Styles ────────────────────────────────────────────────────────────

var styleBase = lipgloss.NewStyle().
	Background(colBackground).
	Foreground(colText)

var stylePanel = lipgloss.NewStyle().
	Background(colSurface).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colBorder).
	Padding(0, 1)

var stylePanelFocus = stylePanel.
	BorderForeground(colBorderFocus)

// ── Logo / Header ──────────────────────────────────────────────────────────

var styleLogo = lipgloss.NewStyle().
	Foreground(colAccent).
	Bold(true).
	PaddingLeft(1)

var styleTagline = lipgloss.NewStyle().
	Foreground(colTextMuted).
	PaddingLeft(1)

// ── Lists ──────────────────────────────────────────────────────────────────

var styleListItem = lipgloss.NewStyle().
	Foreground(colText).
	PaddingLeft(2)

var styleListItemSelected = lipgloss.NewStyle().
	Background(colAccentDim).
	Foreground(colText).
	Bold(true).
	PaddingLeft(1)

var styleListItemIcon = lipgloss.NewStyle().
	Foreground(colAccent)

var styleCount = lipgloss.NewStyle().
	Foreground(colTextMuted).
	PaddingLeft(1)

// ── Input Fields ───────────────────────────────────────────────────────────

var styleInputLabel = lipgloss.NewStyle().
	Foreground(colTextMuted).
	PaddingLeft(2)

var styleInputBox = lipgloss.NewStyle().
	Background(colSurface).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colBorderFocus).
	Foreground(colText).
	Padding(0, 1).
	Width(40)

var styleInputPlaceholder = lipgloss.NewStyle().
	Foreground(colTextSubtle)

var styleCursor = lipgloss.NewStyle().
	Background(colAccent).
	Foreground(colBackground)

// ── Status Bar ─────────────────────────────────────────────────────────────

var styleStatusBar = lipgloss.NewStyle().
	Background(colSurface).
	Foreground(colTextMuted).
	PaddingLeft(1).
	PaddingRight(1)

var styleStatusSuccess = styleStatusBar.
	Foreground(colSuccess)

var styleStatusError = styleStatusBar.
	Foreground(colDanger)

var styleStatusInfo = styleStatusBar.
	Foreground(colInfo)

// ── Keybind Help ───────────────────────────────────────────────────────────

var styleKey = lipgloss.NewStyle().
	Background(colBorder).
	Foreground(colText).
	Padding(0, 1).
	Margin(0, 1, 0, 0)

var styleKeyDesc = lipgloss.NewStyle().
	Foreground(colTextMuted)

// ── Value Display ──────────────────────────────────────────────────────────

var styleMasked = lipgloss.NewStyle().
	Foreground(colTextMuted).
	Italic(true)

var styleRevealed = lipgloss.NewStyle().
	Foreground(colSuccess)

// ── Badges / Tags ──────────────────────────────────────────────────────────

var styleBadgeStrong = lipgloss.NewStyle().
	Background(colSuccess).
	Foreground(colBackground).
	Padding(0, 1).
	Bold(true)

var styleBadgeMedium = lipgloss.NewStyle().
	Background(colWarning).
	Foreground(colBackground).
	Padding(0, 1).
	Bold(true)

var styleBadgeWeak = lipgloss.NewStyle().
	Background(colDanger).
	Foreground(colBackground).
	Padding(0, 1).
	Bold(true)

// ── Dividers ───────────────────────────────────────────────────────────────

var styleDivider = lipgloss.NewStyle().
	Foreground(colBorder)

// ── Section Titles ─────────────────────────────────────────────────────────

var styleSectionTitle = lipgloss.NewStyle().
	Foreground(colAccent).
	Bold(true).
	PaddingLeft(1).
	MarginBottom(1)

// ── Generator specific ─────────────────────────────────────────────────────

var styleGenPassword = lipgloss.NewStyle().
	Background(colSurface).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colAccent).
	Foreground(colText).
	Padding(0, 2).
	Bold(true)

var styleCheckboxOn = lipgloss.NewStyle().
	Foreground(colAccent).
	Bold(true)

var styleCheckboxOff = lipgloss.NewStyle().
	Foreground(colTextMuted)

// ── Unlock Screen ──────────────────────────────────────────────────────────

var styleUnlockBox = lipgloss.NewStyle().
	Background(colSurface).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colBorderFocus).
	Padding(1, 3).
	Width(48)

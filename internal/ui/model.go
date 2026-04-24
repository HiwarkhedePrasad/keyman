// Package ui implements the Keyman terminal user interface using Bubble Tea.
package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/keyman/keyman/internal/generator"
	"github.com/keyman/keyman/internal/storage"
)

// ── Modes ─────────────────────────────────────────────────────────────────

type mode int

const (
	modeUnlock        mode = iota // master password prompt on launch
	modeVaultList                 // folder/vault selection screen
	modeCreateVault               // new vault name + password
	modeKeyList                   // key list inside a vault
	modeAddKey                    // enter key name
	modeAddValue                  // enter key value
	modeEditKey                   // edit existing key value
	modeSearch                    // search/filter keys
	modeGenerator                 // password generator screen
	modeDeleteConfirm             // confirm delete key
	modeDeleteVaultConfirm        // confirm delete vault
	modeHelp                      // full help overlay
)

// ── Messages ──────────────────────────────────────────────────────────────

type statusClearMsg struct{}
type tickMsg time.Time

// ── Model ─────────────────────────────────────────────────────────────────

// Model is the root Bubble Tea model for Keyman.
type Model struct {
	mode mode

	// Layout
	width  int
	height int

	// Vaults
	vaults   []string
	vaultSel int

	// Current vault session
	vaultName     string
	vaultPassword string
	vault         *storage.Vault

	// Key list
	keys        []string // sorted display list (filtered or all)
	allKeys     []string // full unfiltered list
	keySel      int
	revealed    map[string]bool // keys whose values are shown

	// Input
	input   string
	tempKey string // key name staged during addKey→addValue flow

	// Create vault flow
	newVaultName string
	newVaultPass string
	newVaultStep int // 0=name, 1=password

	// Search
	searchQuery string

	// Generator
	genOpts   generator.Options
	genResult string
	genCursor int // which option is selected (0-3: length,upper,digits,symbols)

	// Status bar
	status      string
	statusStyle lipgloss.Style
	statusTimer *time.Timer

	// Password visibility toggle
	showPassword bool

	// Delete confirmation target
	deleteTarget string
}

// New returns an initialised Model.
func New() *Model {
	if err := storage.EnsureDataDir(); err != nil {
		panic(err)
	}

	vaults, _ := storage.ListVaults()
	sort.Strings(vaults)

	return &Model{
		mode:        modeVaultList,
		vaults:      vaults,
		revealed:    make(map[string]bool),
		genOpts:     generator.DefaultOptions(),
		statusStyle: styleStatusBar,
	}
}

// ── Init ──────────────────────────────────────────────────────────────────

func (m *Model) Init() tea.Cmd {
	return nil
}

// ── Update ────────────────────────────────────────────────────────────────

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case statusClearMsg:
		m.status = ""
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch m.mode {

	// ── Vault list ─────────────────────────────────────────────────────
	case modeVaultList:
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.vaultSel > 0 {
				m.vaultSel--
			}
		case "down", "j":
			if m.vaultSel < len(m.vaults)-1 {
				m.vaultSel++
			}
		case "enter":
			if len(m.vaults) > 0 {
				m.vaultName = m.vaults[m.vaultSel]
				m.input = ""
				m.showPassword = false
				m.mode = modeUnlock
			}
		case "n":
			m.newVaultName = ""
			m.newVaultPass = ""
			m.newVaultStep = 0
			m.input = ""
			m.mode = modeCreateVault
		case "D":
			if len(m.vaults) > 0 {
				m.deleteTarget = m.vaults[m.vaultSel]
				m.mode = modeDeleteVaultConfirm
			}
		case "?":
			m.mode = modeHelp
		}

	// ── Unlock ─────────────────────────────────────────────────────────
	case modeUnlock:
		switch key {
		case "esc":
			m.mode = modeVaultList
			m.input = ""
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			vault, err := storage.LoadVault(m.vaultName, m.input)
			if err != nil {
				m.setStatus("Wrong password or corrupted vault", styleStatusError)
				m.input = ""
				return m, nil
			}
			m.vault = vault
			m.vaultPassword = m.input
			m.input = ""
			m.revealed = make(map[string]bool)
			m.refreshKeys("")
			m.mode = modeKeyList
		case "ctrl+h", "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case "tab":
			m.showPassword = !m.showPassword
		default:
			if isPrintable(key) {
				m.input += key
			}
		}

	// ── Create vault ───────────────────────────────────────────────────
	case modeCreateVault:
		switch key {
		case "esc":
			m.mode = modeVaultList
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.newVaultStep == 0 {
				if m.input == "" {
					m.setStatus("Vault name cannot be empty", styleStatusError)
					return m, nil
				}
				if storage.VaultExists(m.input) {
					m.setStatus("A vault with that name already exists", styleStatusError)
					return m, nil
				}
				m.newVaultName = m.input
				m.input = ""
				m.newVaultStep = 1
			} else {
				if m.input == "" {
					m.setStatus("Password cannot be empty", styleStatusError)
					return m, nil
				}
				m.newVaultPass = m.input
				m.input = ""
				if err := storage.CreateVault(m.newVaultName, m.newVaultPass); err != nil {
					m.setStatus("Failed to create vault: "+err.Error(), styleStatusError)
					return m, nil
				}
				m.vaults, _ = storage.ListVaults()
				sort.Strings(m.vaults)
				m.vaultSel = 0
				m.mode = modeVaultList
				m.setStatus("Vault \""+m.newVaultName+"\" created", styleStatusSuccess)
			}
		case "ctrl+h", "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case "tab":
			m.showPassword = !m.showPassword
		default:
			if isPrintable(key) {
				m.input += key
			}
		}

	// ── Key list ───────────────────────────────────────────────────────
	case modeKeyList:
		switch key {
		case "ctrl+c":
			return m, tea.Quit
		case "b", "esc":
			m.mode = modeVaultList
			m.vaultName = ""
			m.vaultPassword = ""
			m.vault = nil
		case "up", "k":
			if m.keySel > 0 {
				m.keySel--
			}
		case "down", "j":
			if m.keySel < len(m.keys)-1 {
				m.keySel++
			}
		case "enter", "c":
			if len(m.keys) > 0 {
				k := m.keys[m.keySel]
				v := m.vault.Entries[k]
				if err := clipboard.WriteAll(v); err != nil {
					m.setStatus("Failed to copy to clipboard", styleStatusError)
				} else {
					m.setStatus("Copied \""+k+"\" to clipboard", styleStatusSuccess)
				}
			}
		case "v":
			if len(m.keys) > 0 {
				k := m.keys[m.keySel]
				m.revealed[k] = !m.revealed[k]
			}
		case "a":
			m.input = ""
			m.tempKey = ""
			m.mode = modeAddKey
		case "e":
			if len(m.keys) > 0 {
				k := m.keys[m.keySel]
				m.tempKey = k
				m.input = m.vault.Entries[k]
				m.mode = modeEditKey
			}
		case "d":
			if len(m.keys) > 0 {
				m.deleteTarget = m.keys[m.keySel]
				m.mode = modeDeleteConfirm
			}
		case "/":
			m.searchQuery = ""
			m.input = ""
			m.mode = modeSearch
		case "g":
			m.genResult, _ = generator.Generate(m.genOpts)
			m.mode = modeGenerator
		case "?":
			m.mode = modeHelp
		}

	// ── Add key ────────────────────────────────────────────────────────
	case modeAddKey:
		switch key {
		case "esc":
			m.mode = modeKeyList
		case "enter":
			if m.input == "" {
				m.setStatus("Key name cannot be empty", styleStatusError)
				return m, nil
			}
			if _, exists := m.vault.Entries[m.input]; exists {
				m.setStatus("Key \""+m.input+"\" already exists — press e to edit", styleStatusError)
				return m, nil
			}
			m.tempKey = m.input
			m.input = ""
			m.mode = modeAddValue
		case "ctrl+h", "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if isPrintable(key) {
				m.input += key
			}
		}

	// ── Add value ──────────────────────────────────────────────────────
	case modeAddValue:
		switch key {
		case "esc":
			m.mode = modeKeyList
		case "enter":
			m.vault.Entries[m.tempKey] = m.input
			if err := m.saveVault(); err != nil {
				m.setStatus("Save failed: "+err.Error(), styleStatusError)
				return m, nil
			}
			m.refreshKeys(m.searchQuery)
			m.input = ""
			m.mode = modeKeyList
			m.setStatus("Added \""+m.tempKey+"\"", styleStatusSuccess)
		case "ctrl+h", "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case "ctrl+g":
			// Quick generate into input
			pw, err := generator.Generate(m.genOpts)
			if err == nil {
				m.input = pw
			}
		default:
			if isPrintable(key) {
				m.input += key
			}
		}

	// ── Edit key value ─────────────────────────────────────────────────
	case modeEditKey:
		switch key {
		case "esc":
			m.mode = modeKeyList
		case "enter":
			m.vault.Entries[m.tempKey] = m.input
			if err := m.saveVault(); err != nil {
				m.setStatus("Save failed: "+err.Error(), styleStatusError)
				return m, nil
			}
			m.refreshKeys(m.searchQuery)
			m.input = ""
			m.mode = modeKeyList
			m.setStatus("Updated \""+m.tempKey+"\"", styleStatusSuccess)
		case "ctrl+h", "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if isPrintable(key) {
				m.input += key
			}
		}

	// ── Search ─────────────────────────────────────────────────────────
	case modeSearch:
		switch key {
		case "esc", "enter":
			m.searchQuery = m.input
			m.mode = modeKeyList
			m.refreshKeys(m.searchQuery)
			m.keySel = 0
		case "ctrl+h", "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
				m.searchQuery = m.input
				m.refreshKeys(m.searchQuery)
				m.keySel = 0
			}
		default:
			if isPrintable(key) {
				m.input += key
				m.searchQuery = m.input
				m.refreshKeys(m.searchQuery)
				m.keySel = 0
			}
		}

	// ── Generator ──────────────────────────────────────────────────────
	case modeGenerator:
		switch key {
		case "esc", "b":
			m.mode = modeKeyList
		case "enter", "c":
			if err := clipboard.WriteAll(m.genResult); err != nil {
				m.setStatus("Failed to copy", styleStatusError)
			} else {
				m.setStatus("Password copied to clipboard", styleStatusSuccess)
			}
			m.mode = modeKeyList
		case "r":
			m.genResult, _ = generator.Generate(m.genOpts)
		case "u":
			m.genOpts.UseUpper = !m.genOpts.UseUpper
			m.genResult, _ = generator.Generate(m.genOpts)
		case "d":
			m.genOpts.UseDigits = !m.genOpts.UseDigits
			m.genResult, _ = generator.Generate(m.genOpts)
		case "s":
			m.genOpts.UseSymbols = !m.genOpts.UseSymbols
			m.genResult, _ = generator.Generate(m.genOpts)
		case "+", "=":
			if m.genOpts.Length < 64 {
				m.genOpts.Length++
				m.genResult, _ = generator.Generate(m.genOpts)
			}
		case "-":
			if m.genOpts.Length > 8 {
				m.genOpts.Length--
				m.genResult, _ = generator.Generate(m.genOpts)
			}
		}

	// ── Delete key confirm ──────────────────────────────────────────────
	case modeDeleteConfirm:
		switch key {
		case "y", "Y", "enter":
			delete(m.vault.Entries, m.deleteTarget)
			if err := m.saveVault(); err != nil {
				m.setStatus("Save failed: "+err.Error(), styleStatusError)
			} else {
				m.setStatus("Deleted \""+m.deleteTarget+"\"", styleStatusInfo)
			}
			m.refreshKeys(m.searchQuery)
			if m.keySel >= len(m.keys) && m.keySel > 0 {
				m.keySel--
			}
			m.mode = modeKeyList
		case "n", "N", "esc":
			m.mode = modeKeyList
		}

	// ── Delete vault confirm ────────────────────────────────────────────
	case modeDeleteVaultConfirm:
		switch key {
		case "y", "Y", "enter":
			name := m.deleteTarget
			if err := storage.DeleteVault(name); err != nil {
				m.setStatus("Failed to delete vault: "+err.Error(), styleStatusError)
			} else {
				m.setStatus("Deleted vault \""+name+"\"", styleStatusInfo)
			}
			m.vaults, _ = storage.ListVaults()
			sort.Strings(m.vaults)
			if m.vaultSel >= len(m.vaults) && m.vaultSel > 0 {
				m.vaultSel--
			}
			m.mode = modeVaultList
		case "n", "N", "esc":
			m.mode = modeVaultList
		}

	// ── Help ───────────────────────────────────────────────────────────
	case modeHelp:
		m.mode = modeKeyList
		if m.vault == nil {
			m.mode = modeVaultList
		}
	}

	return m, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────

func (m *Model) refreshKeys(query string) {
	m.allKeys = make([]string, 0, len(m.vault.Entries))
	for k := range m.vault.Entries {
		m.allKeys = append(m.allKeys, k)
	}
	sort.Strings(m.allKeys)

	if query == "" {
		m.keys = m.allKeys
		return
	}

	q := strings.ToLower(query)
	filtered := m.keys[:0]
	for _, k := range m.allKeys {
		if strings.Contains(strings.ToLower(k), q) {
			filtered = append(filtered, k)
		}
	}
	m.keys = filtered
}

func (m *Model) saveVault() error {
	return storage.SaveVault(m.vaultName, m.vaultPassword, m.vault)
}

func (m *Model) setStatus(msg string, style lipgloss.Style) {
	m.status = msg
	m.statusStyle = style
	if m.statusTimer != nil {
		m.statusTimer.Stop()
	}
	m.statusTimer = time.AfterFunc(3*time.Second, func() {
		// Status will clear on next render via a flag
	})
}

func isPrintable(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Check if all runes in the string are printable
	// This allows pasting multi-character text
	for _, r := range s {
		if !unicode.IsPrint(r) && r != '\n' && r != '\t' {
			return false
		}
	}
	return true
}

// maskValue replaces a value string with bullet characters.
func maskValue(v string) string {
	if len(v) == 0 {
		return ""
	}
	return strings.Repeat("●", min(len(v), 16))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// checkbox returns a styled checkbox string.
func checkbox(on bool, label string) string {
	if on {
		return styleCheckboxOn.Render("[✓] " + label)
	}
	return styleCheckboxOff.Render("[ ] " + label)
}

// strengthBadge renders a coloured strength badge.
func strengthBadge(pw string) string {
	s := generator.Strength(pw)
	switch s {
	case "Strong":
		return styleBadgeStrong.Render("  " + s + "  ")
	case "Medium":
		return styleBadgeMedium.Render("  " + s + "  ")
	default:
		return styleBadgeWeak.Render("  " + s + "  ")
	}
}

// renderInput renders the text input with cursor appended.
func renderInput(input string, hidden bool) string {
	display := input
	if hidden {
		display = strings.Repeat("*", len(input))
	}
	return styleInputBox.Render(display + styleCursor.Render(" "))
}

// keyHint renders a single key + description hint.
func keyHint(k, desc string) string {
	return styleKey.Render(k) + styleKeyDesc.Render(desc)
}

// hintRow joins hints with spaces.
func hintRow(hints ...string) string {
	return "  " + strings.Join(hints, "  ")
}

// header returns the logo + tagline block.
func header() string {
	logo := styleLogo.Render("⬡  KEYMAN")
	tag := styleTagline.Render("Secure credential manager")
	return lipgloss.JoinVertical(lipgloss.Left, logo, tag)
}

// divider draws a horizontal line of the given width.
func divider(width int) string {
	if width <= 0 {
		width = 60
	}
	return styleDivider.Render(strings.Repeat("─", width))
}

// ── View ──────────────────────────────────────────────────────────────────

func (m *Model) View() string {
	if m.width == 0 {
		return ""
	}

	switch m.mode {
	case modeVaultList:
		return m.viewVaultList()
	case modeUnlock:
		return m.viewUnlock()
	case modeCreateVault:
		return m.viewCreateVault()
	case modeKeyList:
		return m.viewKeyList()
	case modeAddKey:
		return m.viewAddKey()
	case modeAddValue:
		return m.viewAddValue()
	case modeEditKey:
		return m.viewEditKey()
	case modeSearch:
		return m.viewSearch()
	case modeGenerator:
		return m.viewGenerator()
	case modeDeleteConfirm:
		return m.viewDeleteConfirm()
	case modeDeleteVaultConfirm:
		return m.viewDeleteVaultConfirm()
	case modeHelp:
		return m.viewHelp()
	}
	return ""
}

// ── Screen: Vault List ────────────────────────────────────────────────────

func (m *Model) viewVaultList() string {
	var b strings.Builder

	b.WriteString(header())
	b.WriteString("\n\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n\n")

	if len(m.vaults) == 0 {
		b.WriteString(styleTextMuted("  No vaults found. Press  n  to create one.\n"))
	} else {
		b.WriteString(styleSectionTitle.Render("  Vaults"))
		b.WriteString("\n")
		for i, v := range m.vaults {
			if i == m.vaultSel {
				icon := styleListItemIcon.Render("⬡")
				label := styleListItemSelected.Render(fmt.Sprintf(" %-30s", v))
				b.WriteString("  " + icon + label + "\n")
			} else {
				icon := styleTextMuted("  ")
				label := styleListItem.Render(fmt.Sprintf("  %-30s", v))
				b.WriteString(icon + label + "\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n")
	b.WriteString(hintRow(
		keyHint("enter", "open"),
		keyHint("n", "new vault"),
		keyHint("D", "delete vault"),
		keyHint("?", "help"),
		keyHint("q", "quit"),
	))
	b.WriteString("\n")

	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(m.statusStyle.Render("  " + m.status))
	}

	return b.String()
}

// ── Screen: Unlock ────────────────────────────────────────────────────────

func (m *Model) viewUnlock() string {
	var b strings.Builder

	b.WriteString(header())
	b.WriteString("\n\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n\n")

	b.WriteString(styleSectionTitle.Render(fmt.Sprintf("  Unlock  \"%s\"", m.vaultName)))
	b.WriteString("\n\n")
	b.WriteString(styleInputLabel.Render("  Master password"))
	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(renderInput(m.input, !m.showPassword))
	b.WriteString("\n")
	b.WriteString(styleTextMuted("  tab to reveal / hide"))
	b.WriteString("\n\n")

	b.WriteString(divider(m.width))
	b.WriteString("\n")
	b.WriteString(hintRow(keyHint("enter", "unlock"), keyHint("tab", "show/hide"), keyHint("esc", "back")))
	b.WriteString("\n")

	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(m.statusStyle.Render("  " + m.status))
	}

	return b.String()
}

// ── Screen: Create Vault ──────────────────────────────────────────────────

func (m *Model) viewCreateVault() string {
	var b strings.Builder

	b.WriteString(header())
	b.WriteString("\n\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleSectionTitle.Render("  Create new vault"))
	b.WriteString("\n\n")

	if m.newVaultStep == 0 {
		b.WriteString(styleInputLabel.Render("  Vault name"))
		b.WriteString("\n  ")
		b.WriteString(renderInput(m.input, false))
	} else {
		b.WriteString(styleTextMuted(fmt.Sprintf("  Vault: %s", m.newVaultName)))
		b.WriteString("\n\n")
		b.WriteString(styleInputLabel.Render("  Master password  (keep this safe — it cannot be recovered)"))
		b.WriteString("\n  ")
		b.WriteString(renderInput(m.input, !m.showPassword))
		b.WriteString("\n")
		b.WriteString(styleTextMuted("  tab to reveal / hide"))
	}

	b.WriteString("\n\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n")
	b.WriteString(hintRow(keyHint("enter", "next"), keyHint("esc", "cancel")))
	b.WriteString("\n")

	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(m.statusStyle.Render("  " + m.status))
	}

	return b.String()
}

// ── Screen: Key List ──────────────────────────────────────────────────────

func (m *Model) viewKeyList() string {
	var b strings.Builder

	b.WriteString(header())
	b.WriteString("\n\n")

	// Vault badge + search indicator
	vaultLabel := styleSectionTitle.Render(fmt.Sprintf("  ⬡ %s", m.vaultName))
	countLabel := styleCount.Render(fmt.Sprintf("%d %s", len(m.keys), pluralise(len(m.keys), "entry", "entries")))
	if m.searchQuery != "" {
		countLabel += styleStatusInfo.Render(fmt.Sprintf("  /  \"%s\"", m.searchQuery))
	}
	b.WriteString(vaultLabel + "  " + countLabel)
	b.WriteString("\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n")

	// Key list body
	if len(m.keys) == 0 {
		if m.searchQuery != "" {
			b.WriteString(styleTextMuted("\n  No keys match your search.\n"))
		} else {
			b.WriteString(styleTextMuted("\n  Vault is empty. Press  a  to add an entry.\n"))
		}
	} else {
		maxVisible := m.height - 12
		if maxVisible < 1 {
			maxVisible = 1
		}
		start := 0
		if m.keySel >= maxVisible {
			start = m.keySel - maxVisible + 1
		}
		end := clamp(start+maxVisible, 0, len(m.keys))

		for i := start; i < end; i++ {
			k := m.keys[i]
			v := m.vault.Entries[k]

			var valueStr string
			if m.revealed[k] {
				valueStr = styleRevealed.Render(v)
			} else {
				valueStr = styleMasked.Render(maskValue(v))
			}

			var row string
			if i == m.keySel {
				nameStr := styleListItemSelected.Render(fmt.Sprintf(" %-28s", k))
				row = "  " + styleListItemIcon.Render("→") + nameStr + "  " + valueStr
			} else {
				nameStr := styleListItem.Render(fmt.Sprintf("  %-28s", k))
				row = nameStr + "  " + valueStr
			}
			b.WriteString(row + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n")
	b.WriteString(hintRow(
		keyHint("enter", "copy"),
		keyHint("v", "reveal"),
		keyHint("a", "add"),
		keyHint("e", "edit"),
		keyHint("d", "delete"),
		keyHint("/", "search"),
		keyHint("g", "generate"),
		keyHint("b", "back"),
	))
	b.WriteString("\n")

	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(m.statusStyle.Render("  " + m.status))
	}

	return b.String()
}

// ── Screen: Add Key ───────────────────────────────────────────────────────

func (m *Model) viewAddKey() string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("\n\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleSectionTitle.Render("  Add new entry"))
	b.WriteString("\n\n")
	b.WriteString(styleInputLabel.Render("  Name  (e.g. GitHub API Key, AWS Secret)"))
	b.WriteString("\n  ")
	b.WriteString(renderInput(m.input, false))
	b.WriteString("\n\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n")
	b.WriteString(hintRow(keyHint("enter", "next"), keyHint("esc", "cancel")))
	if m.status != "" {
		b.WriteString("\n\n")
		b.WriteString(m.statusStyle.Render("  " + m.status))
	}
	return b.String()
}

// ── Screen: Add Value ─────────────────────────────────────────────────────

func (m *Model) viewAddValue() string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("\n\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleSectionTitle.Render("  Add new entry"))
	b.WriteString("\n\n")
	b.WriteString(styleTextMuted(fmt.Sprintf("  Name: %s", m.tempKey)))
	b.WriteString("\n\n")
	b.WriteString(styleInputLabel.Render("  Value / Secret"))
	b.WriteString("\n  ")
	b.WriteString(renderInput(m.input, false))
	b.WriteString("\n")
	b.WriteString(styleTextMuted("  ctrl+g  to auto-generate a strong password"))
	b.WriteString("\n\n")
	if m.input != "" {
		b.WriteString("  Strength: ")
		b.WriteString(strengthBadge(m.input))
		b.WriteString("\n\n")
	}
	b.WriteString(divider(m.width))
	b.WriteString("\n")
	b.WriteString(hintRow(keyHint("enter", "save"), keyHint("ctrl+g", "generate"), keyHint("esc", "cancel")))
	if m.status != "" {
		b.WriteString("\n\n")
		b.WriteString(m.statusStyle.Render("  " + m.status))
	}
	return b.String()
}

// ── Screen: Edit Key ──────────────────────────────────────────────────────

func (m *Model) viewEditKey() string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("\n\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleSectionTitle.Render(fmt.Sprintf("  Edit  \"%s\"", m.tempKey)))
	b.WriteString("\n\n")
	b.WriteString(styleInputLabel.Render("  New value"))
	b.WriteString("\n  ")
	b.WriteString(renderInput(m.input, false))
	b.WriteString("\n\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n")
	b.WriteString(hintRow(keyHint("enter", "save"), keyHint("esc", "cancel")))
	if m.status != "" {
		b.WriteString("\n\n")
		b.WriteString(m.statusStyle.Render("  " + m.status))
	}
	return b.String()
}

// ── Screen: Search ────────────────────────────────────────────────────────

func (m *Model) viewSearch() string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("\n\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleSectionTitle.Render("  Search entries"))
	b.WriteString("\n\n")
	b.WriteString(styleInputLabel.Render("  Filter by name"))
	b.WriteString("\n  ")
	b.WriteString(renderInput(m.input, false))
	b.WriteString("\n")
	matches := len(m.keys)
	total := len(m.allKeys)
	b.WriteString(styleTextMuted(fmt.Sprintf("  %d / %d entries", matches, total)))
	b.WriteString("\n\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n")
	b.WriteString(hintRow(keyHint("enter", "apply"), keyHint("esc", "cancel")))
	return b.String()
}

// ── Screen: Generator ────────────────────────────────────────────────────

func (m *Model) viewGenerator() string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("\n\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleSectionTitle.Render("  Password generator"))
	b.WriteString("\n\n")

	// Generated password display
	b.WriteString("  ")
	b.WriteString(styleGenPassword.Render(m.genResult))
	b.WriteString("  ")
	b.WriteString(strengthBadge(m.genResult))
	b.WriteString("\n\n")

	// Options
	b.WriteString(styleTextMuted("  Options\n\n"))
	b.WriteString(fmt.Sprintf("  Length: %s %d %s\n\n",
		styleKey.Render("-"),
		m.genOpts.Length,
		styleKey.Render("+"),
	))
	b.WriteString("  " + checkbox(m.genOpts.UseUpper, "Uppercase  [u]") + "\n")
	b.WriteString("  " + checkbox(m.genOpts.UseDigits, "Digits     [d]") + "\n")
	b.WriteString("  " + checkbox(m.genOpts.UseSymbols, "Symbols    [s]") + "\n")
	b.WriteString("\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n")
	b.WriteString(hintRow(
		keyHint("enter", "copy & close"),
		keyHint("r", "regenerate"),
		keyHint("+/-", "length"),
		keyHint("esc", "back"),
	))
	b.WriteString("\n")
	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(m.statusStyle.Render("  " + m.status))
	}
	return b.String()
}

// ── Screen: Delete Confirm ────────────────────────────────────────────────

func (m *Model) viewDeleteConfirm() string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("\n\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleStatusError.Render(fmt.Sprintf("  Delete  \"%s\"?", m.deleteTarget)))
	b.WriteString("\n\n")
	b.WriteString(styleTextMuted("  This cannot be undone.\n\n"))
	b.WriteString(divider(m.width))
	b.WriteString("\n")
	b.WriteString(hintRow(keyHint("y", "yes, delete"), keyHint("n / esc", "cancel")))
	return b.String()
}

// ── Screen: Delete Vault Confirm ──────────────────────────────────────────

func (m *Model) viewDeleteVaultConfirm() string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("\n\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleStatusError.Render(fmt.Sprintf("  Permanently delete vault  \"%s\"?", m.deleteTarget)))
	b.WriteString("\n\n")
	b.WriteString(styleTextMuted("  All entries will be lost. This cannot be undone.\n\n"))
	b.WriteString(divider(m.width))
	b.WriteString("\n")
	b.WriteString(hintRow(keyHint("y", "yes, delete"), keyHint("n / esc", "cancel")))
	return b.String()
}

// ── Screen: Help ──────────────────────────────────────────────────────────

func (m *Model) viewHelp() string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("\n\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n\n")
	b.WriteString(styleSectionTitle.Render("  Keyboard shortcuts"))
	b.WriteString("\n\n")

	helpLine := func(keys, desc string) string {
		return fmt.Sprintf("  %-20s %s\n",
			styleKey.Render(keys),
			styleKeyDesc.Render(desc),
		)
	}

	b.WriteString(styleTextMuted("  Vault list\n"))
	b.WriteString(helpLine("enter", "Open vault"))
	b.WriteString(helpLine("n", "New vault"))
	b.WriteString(helpLine("D", "Delete vault"))
	b.WriteString(helpLine("q", "Quit"))
	b.WriteString("\n")

	b.WriteString(styleTextMuted("  Key list\n"))
	b.WriteString(helpLine("enter / c", "Copy value to clipboard"))
	b.WriteString(helpLine("v", "Reveal / hide value"))
	b.WriteString(helpLine("a", "Add new entry"))
	b.WriteString(helpLine("e", "Edit selected entry"))
	b.WriteString(helpLine("d", "Delete selected entry"))
	b.WriteString(helpLine("/", "Search / filter entries"))
	b.WriteString(helpLine("g", "Open password generator"))
	b.WriteString(helpLine("b / esc", "Back to vault list"))
	b.WriteString("\n")

	b.WriteString(styleTextMuted("  Generator\n"))
	b.WriteString(helpLine("r", "Regenerate"))
	b.WriteString(helpLine("+/-", "Increase / decrease length"))
	b.WriteString(helpLine("u / d / s", "Toggle uppercase / digits / symbols"))
	b.WriteString(helpLine("enter", "Copy and close"))
	b.WriteString("\n")

	b.WriteString(styleTextMuted("  General\n"))
	b.WriteString(helpLine("ctrl+g", "Generate password into value field"))
	b.WriteString(helpLine("tab", "Toggle password visibility"))
	b.WriteString(helpLine("?", "This help screen"))

	b.WriteString("\n")
	b.WriteString(divider(m.width))
	b.WriteString("\n")
	b.WriteString(hintRow(keyHint("any key", "close help")))
	return b.String()
}

// ── Small utilities ───────────────────────────────────────────────────────

func styleTextMuted(s string) string {
	return styleListItem.Copy().Foreground(colTextMuted).Render(s)
}

func pluralise(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

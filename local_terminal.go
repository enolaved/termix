package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// session is one local shell in its own pseudo-terminal.
type session struct {
	id   int
	name string
	term *terminal.Terminal
}

func (s *session) running() bool {
	if s == nil || s.term == nil {
		return false
	}
	select {
	case <-s.term.Done():
		return false
	default:
		return true
	}
}

// terminalTheme maps the app's palette onto the terminal emulator's colors.
func (a *app) terminalTheme() *terminal.Theme {
	p := a.palette()
	return &terminal.Theme{
		Foreground:    p.ink,
		Background:    p.terminal,
		Cursor:        p.accent,
		CursorText:    p.onAccent,
		Selection:     p.selected,
		SelectionText: p.ink,
		Palette: [16]ui.Color{
			ui.Hex("#0B0E13"), p.red, p.green, p.amber,
			p.blue, p.purple, p.cyan, p.ink2,
			p.ink3, p.red, p.mint, p.amber,
			p.blue, p.purple, p.cyan, p.ink,
		},
	}
}

// newSession opens the user's login shell in a new pseudo-terminal and
// makes it the active session. The emulator draws it with Ghostty's
// libghostty-vt and sends it the keyboard.
func (a *app) newSession() {
	a.nextSession++
	term, err := terminal.New(terminal.Options{
		ShellIntegration: true,
		Font:             terminal.Font{Family: "Menlo", Size: 12.5, LineHeight: 1.2},
		Padding:          terminal.Padding{X: 20, Y: 16},
		Theme:            a.terminalTheme(),
		OnTitle:          func(string) { a.invalidate() },
		OnExit:           func(int) { a.invalidate() },
	})
	if err != nil {
		a.termErr = err
		a.invalidate()
		return
	}
	a.termErr = nil
	s := &session{
		id:   a.nextSession,
		name: fmt.Sprintf("终端 %d", a.nextSession),
		term: term,
	}
	a.sessions = append(a.sessions, s)
	a.activeID = s.id
	a.invalidate()
}

func (a *app) invalidate() {
	if a.win != nil {
		a.win.Invalidate()
	}
}

// syncTerminalTheme keeps every emulator's colors in step with the app theme.
func (a *app) syncTerminalTheme() {
	for _, s := range a.sessions {
		if s.term != nil {
			s.term.SetTheme(a.terminalTheme(), nil)
		}
	}
}

func (a *app) activeSession() *session {
	return a.sessionByID(a.activeID)
}

func (a *app) sessionByID(id int) *session {
	for _, s := range a.sessions {
		if s.id == id {
			return s
		}
	}
	return nil
}

func (a *app) sessionIndex(id int) int {
	for i, s := range a.sessions {
		if s.id == id {
			return i
		}
	}
	return -1
}

func (a *app) selectSession(id int) {
	if a.sessionByID(id) == nil {
		return
	}
	a.activeID = id
	a.invalidate()
}

func (a *app) closeSession(id int) {
	i := a.sessionIndex(id)
	if i < 0 {
		return
	}
	s := a.sessions[i]
	a.sessions = append(a.sessions[:i], a.sessions[i+1:]...)
	if s.term != nil {
		_ = s.term.Close()
	}
	if a.renameID == id {
		a.renaming = false
	}
	if a.activeID == id {
		switch {
		case len(a.sessions) == 0:
			a.activeID = 0
		case i < len(a.sessions):
			a.activeID = a.sessions[i].id
		default:
			a.activeID = a.sessions[len(a.sessions)-1].id
		}
	}
	a.invalidate()
}

func (a *app) terminalState() (string, ui.Color) {
	p := a.palette()
	s := a.activeSession()
	switch {
	case s == nil && a.termErr != nil:
		return "未连接", p.red
	case s == nil:
		return "就绪", p.accent
	case !s.running():
		return "已退出", p.ink3
	default:
		return "就绪", p.accent
	}
}

func (a *app) terminalPane(c *ui.Context, p palette) {
	if a.renaming {
		// Keep the shell running while the dialog owns the keyboard.
		ui.Box(c).Fill().Background(p.terminal)
		return
	}
	if s := a.activeSession(); s != nil && s.term != nil {
		terminal.View(c, s.term).Fill().AutoFocus()
		return
	}
	if a.termErr != nil {
		ui.Column(c).FillHeight().Padding(22, 18, 22, 18).Gap(6).Children(func() {
			ui.Text(c, "本机终端启动失败").FontSize(12.5).FontWeight(600).TextColor(p.red)
			ui.Text(c, a.termErr.Error()).Font("Menlo").FontSize(10.5).TextColor(p.ink3)
		})
		return
	}
	a.terminalPrompt(c, p)
}

// beginRename opens the rename dialog for s, selecting its current name.
func (a *app) beginRename(c *ui.Context, s *session) {
	a.renaming = true
	a.renameID = s.id
	a.renameText = s.name
	a.renameFocus.SelectAll(c)
}

func (a *app) commitRename() {
	name := strings.TrimSpace(a.renameText)
	if name == "" {
		return
	}
	if s := a.sessionByID(a.renameID); s != nil {
		s.name = name
	}
	a.renaming = false
}

func (a *app) renameDialog(c *ui.Context, p palette) {
	if !a.renaming {
		return
	}
	ui.DialogBase(c, &a.renaming, func(backdrop, panel ui.Element) {
		backdrop.Background(ui.RGBA(0, 0, 0, 0.5))
		panel.Width(360).Padding(18, 18, 18, 18).Radius(10).
			Background(p.elevated).Border(1, p.border2).Gap(14).Children(func() {
			ui.Text(c, "重命名会话").FontSize(13).FontWeight(600).TextColor(p.ink)
			field := ui.Row(c).Height(34).Padding(0, 10, 0, 10).Radius(6).
				Background(p.field)
			if a.renameFocus.Focused(c) {
				field.Border(1, p.accent)
			} else {
				field.Border(1, p.border)
			}
			field.Children(func() {
				ui.TextInputBase(c, &a.renameText).Bind(&a.renameFocus).Grow(1).
					Placeholder("会话名称").PlaceholderColor(p.ink3).TextColor(p.ink).
					OnSubmit(func() { a.commitRename() })
			})
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				cancel := a.dialogButton(c, p, "rename-cancel", "取消", false)
				cancel.OnClick(func() {
					a.renaming = false
				})
				save := a.dialogButton(c, p, "rename-save", "保存", true)
				save.Disabled(strings.TrimSpace(a.renameText) == "")
				save.OnClick(func() { a.commitRename() })
			})
		})
	})
}

func (a *app) dialogButton(c *ui.Context, p palette, key, label string, primary bool) ui.Element {
	btn := ui.ButtonBase(c.Key(key)).Height(30).Padding(0, 14, 0, 14).Radius(6).
		AlignItems(ui.Center).Cursor(ui.CursorPointer).Label(label)
	if primary {
		btn.Background(p.accent)
	} else {
		btn.Background(p.field).Border(1, p.border)
	}
	if btn.Hovered() {
		if primary {
			btn.Background(p.mint)
		} else {
			btn.Border(1, p.border2)
		}
	}
	col := p.ink2
	if primary {
		col = p.onAccent
	}
	btn.Children(func() {
		ui.Text(c, label).FontSize(11.5).FontWeight(500).TextColor(col).SingleLine()
	})
	return btn
}

// sessionPath is the working directory the shell reports, with the home
// directory shortened to ~.
func (a *app) sessionPath(s *session) string {
	if s != nil && s.term != nil {
		if dir := s.term.Dir(); dir != "" {
			return shortHome(dir)
		}
	}
	return "~"
}

func shortHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if path == home {
		return "~"
	}
	if strings.HasPrefix(path, home+string(os.PathSeparator)) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

func shellName() string {
	if sh := os.Getenv("SHELL"); sh != "" {
		return filepath.Base(sh)
	}
	return "zsh"
}

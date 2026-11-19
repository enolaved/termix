package main

import (
	"strconv"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// palette holds the theme tokens from the Termix design board.
type palette struct {
	canvas         ui.Color
	terminal       ui.Color
	surface        ui.Color
	elevated       ui.Color
	field          ui.Color
	border         ui.Color
	border2        ui.Color
	ink            ui.Color
	ink2           ui.Color
	ink3           ui.Color
	accent         ui.Color
	onAccent       ui.Color
	onAccentMuted  ui.Color
	mint           ui.Color
	green          ui.Color
	blue           ui.Color
	cyan           ui.Color
	purple         ui.Color
	amber          ui.Color
	red            ui.Color
	redMuted       ui.Color
	errorBG        ui.Color
	selected       ui.Color
	selectedBorder ui.Color
	tabBorder      ui.Color
	subtle         ui.Color
	footerBorder   ui.Color
	tabBar         ui.Color
	runBG          ui.Color
	runBorder      ui.Color
	paneHeader     ui.Color
	paneBorder     ui.Color
	statusBar      ui.Color
	statusBorder   ui.Color
}

var dark = palette{
	canvas:         ui.Hex("#0C0F14"),
	terminal:       ui.Hex("#090B0F"),
	surface:        ui.Hex("#12151B"),
	elevated:       ui.Hex("#151A21"),
	field:          ui.Hex("#171B22"),
	border:         ui.Hex("#1F252E"),
	border2:        ui.Hex("#252C36"),
	ink:            ui.Hex("#E8EBEF"),
	ink2:           ui.Hex("#A6AFBD"),
	ink3:           ui.Hex("#5C6572"),
	accent:         ui.Hex("#3DDC97"),
	onAccent:       ui.Hex("#07130D"),
	onAccentMuted:  ui.RGBA(7, 19, 13, 0.55),
	mint:           ui.Hex("#6FDCA4"),
	green:          ui.Hex("#9BCB79"),
	blue:           ui.Hex("#61AFEF"),
	cyan:           ui.Hex("#56B6C2"),
	purple:         ui.Hex("#C678DD"),
	amber:          ui.Hex("#E5C07B"),
	red:            ui.Hex("#E96C75"),
	redMuted:       ui.RGBA(233, 108, 117, 0.65),
	errorBG:        ui.Hex("#2A1417"),
	selected:       ui.Hex("#1A2029"),
	selectedBorder: ui.Hex("#222A34"),
	tabBorder:      ui.Hex("#2A323D"),
	subtle:         ui.Hex("#1A202B"),
	footerBorder:   ui.Hex("#191D25"),
	tabBar:         ui.Hex("#0F1218"),
	runBG:          ui.Hex("#12251D"),
	runBorder:      ui.Hex("#1E3B2D"),
	paneHeader:     ui.Hex("#0D1015"),
	paneBorder:     ui.Hex("#13171E"),
	statusBar:      ui.Hex("#0E1116"),
	statusBorder:   ui.Hex("#12161D"),
}

var light = palette{
	canvas:         ui.Hex("#EEF0F3"),
	terminal:       ui.Hex("#FFFFFF"),
	surface:        ui.Hex("#F7F8FA"),
	elevated:       ui.Hex("#FFFFFF"),
	field:          ui.Hex("#FFFFFF"),
	border:         ui.Hex("#E2E6EC"),
	border2:        ui.Hex("#D6DCE4"),
	ink:            ui.Hex("#16191E"),
	ink2:           ui.Hex("#4C5664"),
	ink3:           ui.Hex("#687382"),
	accent:         ui.Hex("#12B87A"),
	onAccent:       ui.Hex("#053021"),
	onAccentMuted:  ui.RGBA(5, 48, 33, 0.6),
	mint:           ui.Hex("#0A7A56"),
	green:          ui.Hex("#3E8E3A"),
	blue:           ui.Hex("#2C6BCB"),
	cyan:           ui.Hex("#0E7490"),
	purple:         ui.Hex("#8A4BC2"),
	amber:          ui.Hex("#A66A16"),
	red:            ui.Hex("#C93B52"),
	redMuted:       ui.RGBA(201, 59, 82, 0.72),
	errorBG:        ui.Hex("#FCEDEF"),
	selected:       ui.Hex("#E9EDF2"),
	selectedBorder: ui.Hex("#D6DEE8"),
	tabBorder:      ui.Hex("#CCD5E1"),
	subtle:         ui.Hex("#EDF0F4"),
	footerBorder:   ui.Hex("#E7EAEF"),
	tabBar:         ui.Hex("#F2F4F7"),
	runBG:          ui.Hex("#E2F6ED"),
	runBorder:      ui.Hex("#BFE9D7"),
	paneHeader:     ui.Hex("#F5F7F9"),
	paneBorder:     ui.Hex("#E7EAEF"),
	statusBar:      ui.Hex("#F2F4F7"),
	statusBorder:   ui.Hex("#E5E9EE"),
}

type app struct {
	light       bool
	win         *mygo.Window
	sessions    []*session
	activeID    int
	nextSession int
	termErr     error

	renaming    bool
	renameID    int
	renameText  string
	renameFocus ui.Handle
}

func (a *app) palette() palette {
	if a.light {
		return light
	}
	return dark
}

func (a *app) view(c *ui.Context) {
	c.OnShortcut(ui.Cmd, ui.KeyT, func() { a.newSession() })
	p := a.palette()
	ui.Row(c).Fill().Background(p.canvas).Children(func() {
		a.sidebar(c, p)
		ui.Column(c).Grow(1).FillHeight().Children(func() {
			a.tabBar(c, p)
			a.sessionToolbar(c, p)
			a.panes(c, p)
			a.workspaceStatus(c, p)
		})
	})
	a.renameDialog(c, p)
}

func (a *app) sidebar(c *ui.Context, p palette) {
	ui.Column(c).Width(300).FillHeight().Background(p.surface).
		BorderWidth(0, 1, 0, 0).BorderColor(p.border).Children(func() {
		a.brand(c, p)
		a.newTerminalButton(c, p)
		a.searchField(c, p)
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Column(c).Padding(18, 8, 8, 8).Children(func() {
				a.sectionLabel(c, p, "会话", strconv.Itoa(len(a.sessions)), false)
				if len(a.sessions) == 0 {
					a.emptySessions(c, p)
				} else {
					for _, s := range a.sessions {
						a.localSession(c, p, s)
					}
				}
				a.shortcuts(c, p)
			})
		})
		a.sidebarFooter(c, p)
	})
}

func (a *app) brand(c *ui.Context, p palette) {
	ui.Row(c).Height(68).Padding(0, 16, 0, 16).AlignItems(ui.Center).DragWindow().Children(func() {
		ui.Box(c).Size(36, 36).Radius(10).Background(p.accent).Center().Children(func() {
			ui.Text(c, ">_").Font("Menlo").FontSize(15).Bold().TextColor(p.onAccent)
		})
		ui.Column(c).Padding(0, 0, 0, 12).Gap(1).Children(func() {
			ui.Text(c, "termix").FontSize(17).FontWeight(600).TextColor(p.ink).SingleLine()
			ui.Text(c, "本地工作区").FontSize(9.5).TextColor(p.ink3).SingleLine()
		})
		ui.Spacer(c)
		icon(c, svgBell, 18, p.ink3)
	})
}

func (a *app) newTerminalButton(c *ui.Context, p palette) {
	btn := ui.ButtonBase(c.Key("new-terminal")).Height(40).Margin(0, 16, 0, 16).
		Padding(0, 12, 0, 12).Radius(8).Background(p.accent).Gap(8).
		AlignItems(ui.Center).Cursor(ui.CursorPointer).
		Label("新建终端").Tooltip("新建终端 ⌘T")
	btn.Children(func() {
		icon(c, svgPlus, 15, p.onAccent)
		ui.Text(c, "新建终端").FontSize(12.5).FontWeight(600).TextColor(p.onAccent)
		ui.Spacer(c)
		ui.Text(c, "⌘T").FontSize(10).FontWeight(500).TextColor(p.onAccentMuted)
	})
	btn.OnClick(func() { a.newSession() })
}

func (a *app) searchField(c *ui.Context, p palette) {
	ui.Row(c).Height(32).Margin(13, 16, 0, 16).Padding(0, 10, 0, 10).
		Radius(6).Background(p.field).Border(1, p.border).Gap(7).
		AlignItems(ui.Center).Children(func() {
		icon(c, svgSearch, 14, p.ink3)
		ui.Text(c, "搜索会话或路径").FontSize(11.5).TextColor(p.ink3).SingleLine()
		ui.Spacer(c)
		keycap(c, p, "⌘K")
	})
}

func (a *app) sectionLabel(c *ui.Context, p palette, name, count string, filter bool) {
	ui.Row(c).Height(24).Padding(0, 8, 0, 8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, name).FontSize(11.5).FontWeight(500).TextColor(p.ink2)
		ui.Text(c, " "+count).FontSize(11.5).TextColor(p.ink3)
		ui.Spacer(c)
		if filter {
			icon(c, svgSliders, 14, p.ink3)
		}
	})
}

func (a *app) emptySessions(c *ui.Context, p palette) {
	ui.Column(c).Margin(0, 8, 0, 0).Padding(16, 14, 16, 14).Radius(8).
		Background(p.elevated).Border(1, p.border).Gap(6).Children(func() {
		ui.Text(c, "还没有打开的会话").FontSize(11.5).FontWeight(500).TextColor(p.ink2)
		ui.Text(c, "点击“新建终端”开始一个新的 shell。").FontSize(10.5).TextColor(p.ink3)
	})
}

func (a *app) localSession(c *ui.Context, p palette, s *session) {
	active := s.id == a.activeID
	row := ui.ButtonBase(c.Key("session-"+strconv.Itoa(s.id))).Height(56).
		Margin(0, 0, 0, 8).Padding(0, 16, 0, 16).Radius(6).
		Cursor(ui.CursorPointer).Label(s.name)
	switch {
	case active:
		row.Background(p.selected).Border(1, p.selectedBorder)
	case row.Hovered():
		row.Background(p.elevated)
	}
	row.ContextMenu(func(m *ui.Menu) {
		if m.Item("重命名会话…").Chosen() {
			a.beginRename(c, s)
		}
		m.Separator()
		if m.Item("关闭会话").Chosen() {
			a.closeSession(s.id)
		}
	})
	row.Children(func() {
		if active {
			ui.Box(c).Absolute().Left(0).Top(12).Bottom(12).Width(2).
				Radius(1).Background(p.accent)
		}
		ui.Column(c).FillHeight().FillWidth().Justify(ui.Center).Gap(3).Children(func() {
			ui.Row(c).AlignItems(ui.Center).Children(func() {
				col := p.ink
				if !active {
					col = p.ink2
				}
				dot(c, 6, sessionDot(p, s))
				ui.Text(c, s.name).Padding(0, 0, 0, 8).FontSize(12).
					FontWeight(500).TextColor(col).SingleLine()
				ui.Spacer(c)
				ui.Box(c).Height(16).Padding(0, 5, 0, 5).Radius(4).
					Background(p.subtle).Center().Children(func() {
					ui.Text(c, "本地").FontSize(9).TextColor(p.ink2).SingleLine()
				})
			})
			ui.Text(c, a.sessionPath(s)).FontSize(10.5).TextColor(p.ink3).SingleLine()
		})
	})
	row.OnClick(func() { a.selectSession(s.id) })
}

func dot(c *ui.Context, size float32, col ui.Color) {
	ui.Box(c).Size(size, size).Radius(size / 2).Background(col)
}

func sessionDot(p palette, s *session) ui.Color {
	if s.running() {
		return p.accent
	}
	return p.ink3
}

func keycap(c *ui.Context, p palette, label string) {
	ui.Box(c).Height(16).MinWidth(30).Padding(0, 6, 0, 6).Radius(4).
		Background(p.subtle).Border(1, p.border2).Center().Children(func() {
		ui.Text(c, label).FontSize(9).TextColor(p.ink3).SingleLine()
	})
}

func (a *app) shortcuts(c *ui.Context, p palette) {
	ui.Column(c).Margin(38, 16, 0, 16).Padding(12, 12, 12, 12).Radius(8).
		Background(p.elevated).Border(1, p.border).Gap(4).Children(func() {
		ui.Text(c, "快捷键").FontSize(10).TextColor(p.ink3)
		shortcutRow(c, p, "⌘T", "新建终端")
		shortcutRow(c, p, "⌘K", "快速切换会话")
		shortcutRow(c, p, "⌘D", "纵向分屏")
	})
}

func shortcutRow(c *ui.Context, p palette, keys, label string) {
	ui.Row(c).Height(16).AlignItems(ui.Center).Gap(10).Children(func() {
		keycap(c, p, keys)
		ui.Text(c, label).FontSize(10.5).TextColor(p.ink2).SingleLine()
	})
}

func (a *app) sidebarFooter(c *ui.Context, p palette) {
	ui.Row(c).Height(54).Padding(12, 16, 0, 16).AlignItems(ui.Start).
		BorderWidth(1, 0, 0, 0).BorderColor(p.footerBorder).Children(func() {
		dot(c, 5, p.accent)
		ui.Text(c, " 本地").FontSize(10).TextColor(p.ink2)
		ui.Spacer(c)
		ui.Text(c, "v0.1.0").FontSize(9.5).TextColor(p.ink3)
	})
}

func (a *app) tabBar(c *ui.Context, p palette) {
	ui.Row(c).Height(44).Padding(0, 16, 0, 16).Background(p.tabBar).
		BorderWidth(0, 0, 1, 0).BorderColor(p.border).AlignItems(ui.Center).
		DragWindow().Children(func() {
		for _, s := range a.sessions {
			a.tab(c, p, s)
		}
		a.newTabButton(c, p)
		ui.Spacer(c)
		ui.Box(c).Width(10)
		icon(c, svgPanelRight, 17, p.ink2)
		ui.Box(c).Width(18)
		icon(c, svgSearch, 16, p.ink2)
		ui.Box(c).Width(18)
		icon(c, svgMore, 17, p.ink2)
		ui.Box(c).Width(14)
		ui.Box(c).Width(1).Height(16).Background(p.border)
		ui.Box(c).Width(14)
		a.themeToggle(c, p)
	})
}

// themeToggle switches the palette and shows the theme in use.
func (a *app) themeToggle(c *ui.Context, p palette) {
	label := "切换到浅色主题"
	if a.light {
		label = "切换到深色主题"
	}
	btn := ui.ButtonBase(c.Key("theme-toggle")).Height(26).Padding(0, 8, 0, 8).
		Radius(6).Gap(6).AlignItems(ui.Center).
		Background(p.subtle).Border(1, p.border).
		Cursor(ui.CursorPointer).Label(label).Tooltip(label)
	if btn.Hovered() {
		btn.Border(1, p.border2)
	}
	btn.Children(func() {
		if a.light {
			icon(c, svgSun, 13, p.ink2)
			ui.Text(c, "浅色").FontSize(10.5).FontWeight(500).TextColor(p.ink2).SingleLine()
			return
		}
		icon(c, svgMoon, 13, p.ink2)
		ui.Text(c, "深色").FontSize(10.5).FontWeight(500).TextColor(p.ink2).SingleLine()
	})
	btn.OnClick(func() {
		a.light = !a.light
		a.syncTerminalTheme()
	})
}

func (a *app) newTabButton(c *ui.Context, p palette) {
	btn := ui.ButtonBase(c.Key("tab-new")).Size(22, 22).Radius(5).Center().
		Cursor(ui.CursorPointer).Label("新建终端").Tooltip("新建终端 ⌘T")
	if btn.Hovered() {
		btn.Background(p.subtle)
	}
	btn.Children(func() {
		icon(c, svgPlus, 17, p.ink2)
	})
	btn.OnClick(func() { a.newSession() })
	ui.Box(c).Width(4)
}

func (a *app) tab(c *ui.Context, p palette, s *session) {
	active := s.id == a.activeID
	tab := ui.Row(c).Width(132).Height(32).Padding(0, 8, 0, 8).Radius(6).
		Gap(4).AlignItems(ui.Center).AlignSelf(ui.End)
	if active {
		tab.Background(p.selected).Border(1, p.tabBorder)
	}
	tab.Children(func() {
		textColor := p.ink2
		if active {
			textColor = p.ink
		}
		hit := ui.ButtonBase(c.Key("tab-" + strconv.Itoa(s.id))).Grow(1).Height(30).
			Gap(8).AlignItems(ui.Center).Cursor(ui.CursorPointer).Label(s.name)
		hit.Children(func() {
			dot(c, 6, sessionDot(p, s))
			ui.Text(c, s.name).FontSize(12).TextColor(textColor).SingleLine()
			ui.Spacer(c)
		})
		hit.OnClick(func() { a.selectSession(s.id) })
		close := ui.ButtonBase(c.Key("close-"+strconv.Itoa(s.id))).Size(16, 16).
			Radius(4).Center().Cursor(ui.CursorPointer).Label("关闭 " + s.name)
		if close.Hovered() {
			close.Background(p.border2)
		}
		close.Children(func() {
			icon(c, svgX, 11, p.ink3)
		})
		close.OnClick(func() { a.closeSession(s.id) })
	})
	ui.Box(c).Width(4)
}

func (a *app) sessionToolbar(c *ui.Context, p palette) {
	ui.Row(c).Height(60).Padding(0, 22, 0, 22).BorderWidth(0, 0, 1, 0).
		BorderColor(p.subtle).AlignItems(ui.Center).Children(func() {
		ui.Text(c, a.sessionPath(a.activeSession())).Font("Menlo").FontSize(12.5).
			TextColor(p.ink2).SingleLine()
		ui.Box(c).Width(22)
		pill(c, p, nil, shellName())
		ui.Spacer(c)
		icon(c, svgSettings, 15, p.ink3)
		ui.Box(c).Width(14)
		icon(c, svgMore, 16, p.ink3)
	})
}

func pill(c *ui.Context, p palette, ic *ui.SVG, label string) {
	ui.Row(c).Height(22).Padding(0, 8, 0, 8).Radius(4).Background(p.subtle).
		Gap(5).AlignItems(ui.Center).Children(func() {
		if ic != nil {
			icon(c, ic, 12, p.ink3)
		}
		ui.Text(c, label).FontSize(10.5).TextColor(p.ink2).SingleLine()
	})
}

func (a *app) panes(c *ui.Context, p palette) {
	ui.Column(c).Grow(1).Background(p.terminal).Clip().Children(func() {
		a.terminalPane(c, p)
	})
}

func (a *app) terminalPrompt(c *ui.Context, p palette) {
	ui.Column(c).Grow(1).Padding(18, 22, 0, 22).Children(func() {
		ui.Row(c).Height(20).AlignItems(ui.Center).Children(func() {
			mono(ui.RichText(c,
				ui.Span{Text: "~", Color: p.blue},
				ui.Span{Text: " ❯ ", Color: p.accent},
			)).NoWrap()
			ui.Box(c).Size(7, 14).Background(p.accent).Margin(0, 1)
		})
	})
}

func (a *app) workspaceStatus(c *ui.Context, p palette) {
	ui.Row(c).Height(30).Background(p.statusBar).Padding(0, 22, 0, 22).
		BorderWidth(1, 0, 0, 0).BorderColor(p.statusBorder).
		AlignItems(ui.Center).Children(func() {
		ui.Text(c, "UTF-8").FontSize(10).TextColor(p.ink3)
		ui.Box(c).Width(16)
		ui.Text(c, shellName()).FontSize(10).TextColor(p.ink3)
		ui.Spacer(c)
		state, color := a.terminalState()
		dot(c, 5, color)
		ui.Text(c, " "+state).FontSize(10).TextColor(p.ink2)
	})
}

func mono(e ui.Element) ui.Element {
	return e.Font("Menlo").FontSize(12).FixedLineHeight(20)
}

func icon(c *ui.Context, s *ui.SVG, size float32, col ui.Color) {
	ui.Icon(c, s).Size(size, size).TextColor(col)
}

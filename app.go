package main

import "github.com/egoist/mygo/ui"

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
	light bool
}

func (a *app) palette() palette {
	if a.light {
		return light
	}
	return dark
}

func (a *app) view(c *ui.Context) {
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
}

func (a *app) sidebar(c *ui.Context, p palette) {
	ui.Column(c).Width(300).FillHeight().Background(p.surface).
		BorderWidth(0, 1, 0, 0).BorderColor(p.border).Children(func() {
		a.brand(c, p)
		a.newTerminalButton(c, p)
		a.searchField(c, p)
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Column(c).Padding(18, 8, 8, 8).Children(func() {
				a.sectionLabel(c, p, "会话", "7", true)
				a.sectionSub(c, p, "运行中", "3")
				a.sessionRow(c, p, session{name: "atlas-web", path: "~/projects/atlas", time: "12:48", badge: "dev", running: true, selected: true})
				a.sessionRow(c, p, session{name: "api-server", path: "~/projects/atlas/api", time: "12:47", badge: "node", running: true})
				a.sessionRow(c, p, session{name: "deploy@prod", path: "ssh://atlas-prod", time: "11:02", badge: "ssh", running: true})

				a.sectionLabelSpaced(c, p, "已停止", "2")
				a.sessionRow(c, p, session{name: "postgres", path: "~/docker/atlas-db", time: "昨天", badge: "db"})
				a.sessionRow(c, p, session{name: "nginx-logs", path: "/var/log/nginx", time: "昨天", badge: "log"})

				a.sectionLabelSpaced(c, p, "最近关闭", "2")
				a.sessionRow(c, p, session{name: "redis-cli", path: "127.0.0.1:6379", time: "周一", badge: "redis"})
				a.sessionRow(c, p, session{name: "prisma-studio", path: "~/projects/atlas", time: "周一", badge: "db"})
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
	ui.Row(c).Height(40).Margin(0, 16, 0, 16).Padding(0, 12, 0, 12).
		Radius(8).Background(p.accent).Gap(8).AlignItems(ui.Center).Children(func() {
		icon(c, svgPlus, 15, p.onAccent)
		ui.Text(c, "新建终端").FontSize(12.5).FontWeight(600).TextColor(p.onAccent)
		ui.Spacer(c)
		ui.Text(c, "⌘T").FontSize(10).FontWeight(500).TextColor(p.onAccentMuted)
	})
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

func (a *app) sectionLabelSpaced(c *ui.Context, p palette, name, count string) {
	ui.Box(c).Height(18)
	a.sectionLabel(c, p, name, count, false)
}

func (a *app) sectionSub(c *ui.Context, p palette, name, count string) {
	ui.Row(c).Height(18).Padding(0, 8, 0, 8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, name+" "+count).FontSize(10).TextColor(p.ink3)
	})
}

type session struct {
	name, path, time, badge string
	running, selected       bool
}

func (a *app) sessionRow(c *ui.Context, p palette, s session) {
	row := ui.Box(c).Height(56).Margin(0, 0, 0, 8).Padding(0, 16, 0, 16).Radius(6)
	if s.selected {
		row.Background(p.selected).Border(1, p.selectedBorder)
	}
	row.Children(func() {
		if s.selected {
			ui.Box(c).Absolute().Left(0).Top(12).Bottom(12).Width(2).
				Radius(1).Background(p.accent)
		}
		ui.Column(c).FillHeight().Justify(ui.Center).Gap(3).Children(func() {
			ui.Row(c).AlignItems(ui.Center).Children(func() {
				dot(c, 6, dotColor(p, s))
				ui.Text(c, s.name).Padding(0, 0, 0, 8).FontSize(12).FontWeight(500).
					TextColor(nameColor(p, s)).SingleLine()
				ui.Spacer(c)
				ui.Text(c, s.time).FontSize(10).TextColor(p.ink3).SingleLine()
			})
			ui.Row(c).AlignItems(ui.Center).Children(func() {
				ui.Text(c, s.path).Padding(0, 0, 0, 14).FontSize(10.5).
					TextColor(pathColor(p, s)).SingleLine()
				ui.Spacer(c)
				badge(c, p, s.badge, s.running)
			})
		})
	})
}

func dotColor(p palette, s session) ui.Color {
	if s.running {
		return p.accent
	}
	return p.ink3
}

func nameColor(p palette, s session) ui.Color {
	if s.running {
		return p.ink
	}
	return p.ink2
}

func pathColor(p palette, s session) ui.Color {
	return p.ink3
}

func dot(c *ui.Context, size float32, col ui.Color) {
	ui.Box(c).Size(size, size).Radius(size / 2).Background(col)
}

func badge(c *ui.Context, p palette, label string, bright bool) {
	col := p.ink3
	if bright {
		col = p.ink2
	}
	ui.Box(c).Height(16).Padding(0, 5, 0, 5).Radius(4).
		Background(p.subtle).Center().Children(func() {
		ui.Text(c, label).FontSize(9).TextColor(col).SingleLine()
	})
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
		ui.Text(c, " · zsh 5.9").FontSize(10).TextColor(p.ink3)
		ui.Spacer(c)
		ui.Text(c, "v1.8.2").FontSize(9.5).TextColor(p.ink3)
	})
}

func (a *app) tabBar(c *ui.Context, p palette) {
	ui.Row(c).Height(44).Padding(0, 16, 0, 16).Background(p.tabBar).
		BorderWidth(0, 0, 1, 0).BorderColor(p.border).AlignItems(ui.Center).
		DragWindow().Children(func() {
		a.tab(c, p, "atlas-web", true, p.accent)
		a.tab(c, p, "api-server", false, p.accent)
		a.tab(c, p, "deploy", false, p.ink3)
		ui.Spacer(c)
		icon(c, svgPlus, 17, p.ink2)
		ui.Box(c).Width(18)
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
	})
}

func (a *app) tab(c *ui.Context, p palette, name string, active bool, dotCol ui.Color) {
	width := float32(120)
	switch name {
	case "atlas-web":
		width = 174
	case "api-server":
		width = 132
	}
	tab := ui.Row(c).Width(width).Height(32).Padding(0, 12, 0, 12).Radius(6).
		Gap(8).AlignItems(ui.Center).AlignSelf(ui.End)
	if active {
		tab.Background(p.selected).Border(1, p.tabBorder)
	}
	textColor := p.ink2
	if active {
		textColor = p.ink
	}
	tab.Children(func() {
		dot(c, 6, dotCol)
		ui.Text(c, name).FontSize(12).TextColor(textColor).SingleLine()
		ui.Spacer(c)
		icon(c, svgX, 12, p.ink3)
	})
	if name != "deploy" {
		ui.Box(c).Width(4)
	}
}

func (a *app) sessionToolbar(c *ui.Context, p palette) {
	ui.Row(c).Height(60).Padding(0, 22, 0, 22).BorderWidth(0, 0, 1, 0).
		BorderColor(p.subtle).AlignItems(ui.Center).Children(func() {
		ui.Text(c, "~/projects/atlas").Font("Menlo").FontSize(12.5).TextColor(p.ink2)
		ui.Box(c).Width(22)
		pill(c, p, svgBranch, "main")
		ui.Box(c).Width(8)
		pill(c, p, nil, "zsh")
		ui.Box(c).Width(8)
		pill(c, p, nil, "node 20")
		ui.Spacer(c)
		ui.Row(c).Height(24).Padding(0, 9, 0, 9).Radius(5).Gap(6).
			Background(p.runBG).Border(1, p.runBorder).
			AlignItems(ui.Center).Children(func() {
			dot(c, 6, p.accent)
			ui.Text(c, "运行中").FontSize(10.5).TextColor(p.mint)
		})
		ui.Box(c).Width(16)
		ui.Text(c, "12:48:24").FontSize(10.5).TextColor(p.ink3)
		ui.Box(c).Width(18)
		icon(c, svgTrash, 15, p.ink3)
		ui.Box(c).Width(14)
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
		a.terminalOne(c, p)
		a.terminalTwo(c, p)
	})
}

func (a *app) terminalOne(c *ui.Context, p palette) {
	ui.Box(c).Grow(61).Clip().Children(func() {
		ui.Column(c).Padding(18, 22, 0, 22).Gap(0).Children(func() {
			promptLine(c, p, "pnpm install")
			monoLine(c, p, ui.Span{Text: "Packages: +412", Color: p.ink3})
			monoLine(c, p, ui.Span{Text: "Progress: resolved 412, reused 398, downloaded 14, added 412", Color: p.ink3})
			monoLine(c, p,
				ui.Span{Text: "✓", Color: p.green},
				ui.Span{Text: " Done in 8.4s", Color: p.ink2})
			blankLine(c)
			promptLine(c, p, "pnpm dev")
			blankLine(c)
			monoLine(c, p,
				ui.Span{Text: "▲", Color: p.amber},
				ui.Span{Text: " Next.js 15.1.0", Color: p.ink, Weight: 600})
			monoLine(c, p, ui.Span{Text: "- Local: http://localhost:3000", Color: p.ink2})
			monoLine(c, p, ui.Span{Text: "- Environments: .env.local", Color: p.ink2})
			blankLine(c)
			monoLine(c, p,
				ui.Span{Text: "✓", Color: p.green},
				ui.Span{Text: " Ready in 1.2s", Color: p.ink2})
			monoLine(c, p, ui.Span{Text: "○ Compiling / ...", Color: p.ink2})
			monoLine(c, p,
				ui.Span{Text: "✓", Color: p.green},
				ui.Span{Text: " Compiled / in 842ms", Color: p.ink2})
			blankLine(c)
			monoLine(c, p,
				ui.Span{Text: "⚠", Color: p.amber},
				ui.Span{Text: " Fast Refresh had to perform a full reload.", Color: p.ink2})
			errorLine(c, p)
			ui.Box(c).Height(8)
			commandWithCursor(c, p)
			ui.Row(c).Height(20).Margin(56, 0, 0, 0).Justify(ui.End).
				AlignItems(ui.Center).Gap(14).Children(func() {
				ui.Text(c, "tab 补全 · ↑ 历史命令 · ⌘K 命令面板").FontSize(10).TextColor(p.ink3)
				ui.Text(c, "自动滚动").FontSize(10).TextColor(p.ink3)
			})
		})
	})
}

func (a *app) terminalTwo(c *ui.Context, p palette) {
	ui.Column(c).Grow(39).BorderWidth(1, 0, 0, 0).BorderColor(p.paneBorder).Children(func() {
		ui.Row(c).Height(35).Padding(0, 22, 0, 22).Background(p.paneHeader).
			AlignItems(ui.Center).Children(func() {
			dot(c, 6, p.accent)
			ui.Text(c, " api-server").FontSize(11.5).FontWeight(600).TextColor(p.ink)
			ui.Box(c).Width(8)
			badge(c, p, "node", false)
			ui.Text(c, " · ").FontSize(10).TextColor(p.ink3)
			ui.Text(c, ":4000").FontSize(10).TextColor(p.ink3)
			ui.Spacer(c)
			ui.Text(c, "自动滚动").FontSize(10).TextColor(p.ink3)
		})
		ui.Box(c).Grow(1).Clip().Children(func() {
			ui.Column(c).Padding(16, 22, 0, 22).Gap(0).Children(func() {
				logLine(c, p, "12:47:02", "ready", p.mint, "server listening on 4000")
				logLine(c, p, "12:47:03", "info", p.ink2, "prisma connected · pool 5")
				logLine(c, p, "12:48:11", "warn", p.amber, "slow query 1240ms · GET /api/projects")
				logLine(c, p, "12:48:19", "error", p.red, "ECONNREFUSED 127.0.0.1:5432")
				logLine(c, p, "12:48:19", "info", p.ink2, "retrying in 2s (1/5)")
				logLine(c, p, "12:48:21", "ready", p.mint, "reconnected · 6ms")
				logLine(c, p, "12:48:22", "info", p.ink2, "GET /api/health 200 4ms")
				logLine(c, p, "12:48:24", "info", p.ink2, "GET /api/projects 200 38ms")
				logLine(c, p, "12:48:26", "info", p.ink2, "HMR update · src/app/page.tsx")
				logLine(c, p, "12:48:27", "ready", p.mint, "compiled successfully in 124ms")
				logLine(c, p, "12:48:29", "info", p.ink2, "ws connected · 2 clients")
				ui.Row(c).Height(20).Margin(10, 0, 0, 0).Justify(ui.End).
					AlignItems(ui.Center).Children(func() {
					ui.Text(c, "tail -f  自动滚动").FontSize(10).TextColor(p.ink3)
				})
			})
		})
	})
}

func (a *app) workspaceStatus(c *ui.Context, p palette) {
	ui.Row(c).Height(30).Background(p.statusBar).Padding(0, 22, 0, 22).
		BorderWidth(1, 0, 0, 0).BorderColor(p.statusBorder).
		AlignItems(ui.Center).Children(func() {
		icon(c, svgBranch, 12, p.ink3)
		ui.Text(c, " main").FontSize(10).TextColor(p.ink2)
		ui.Box(c).Width(16)
		ui.Text(c, "UTF-8").FontSize(10).TextColor(p.ink3)
		ui.Box(c).Width(16)
		ui.Text(c, "zsh 5.9").FontSize(10).TextColor(p.ink3)
		ui.Box(c).Width(16)
		ui.Text(c, "node 20.11").FontSize(10).TextColor(p.ink3)
		ui.Spacer(c)
		dot(c, 5, p.accent)
		ui.Text(c, " 3 个进程组运行中").FontSize(10).TextColor(p.mint)
	})
}

func mono(e ui.Element) ui.Element {
	return e.Font("Menlo").FontSize(12).FixedLineHeight(20)
}

func promptLine(c *ui.Context, p palette, cmd string) {
	mono(ui.RichText(c,
		ui.Span{Text: "enola@studio", Color: p.mint},
		ui.Span{Text: ":", Color: p.ink3},
		ui.Span{Text: " ~/projects/atlas", Color: p.blue},
		ui.Span{Text: " main ", Color: p.purple},
		ui.Span{Text: "❯ ", Color: p.accent},
		ui.Span{Text: cmd, Color: p.ink},
	)).NoWrap()
}

func commandWithCursor(c *ui.Context, p palette) {
	ui.Row(c).Height(20).AlignItems(ui.Center).Children(func() {
		mono(ui.RichText(c,
			ui.Span{Text: "enola@studio", Color: p.mint},
			ui.Span{Text: ":", Color: p.ink3},
			ui.Span{Text: " ~/projects/atlas", Color: p.blue},
			ui.Span{Text: " main ", Color: p.purple},
			ui.Span{Text: "❯ ", Color: p.accent},
			ui.Span{Text: "pnpm run buil", Color: p.ink},
		)).NoWrap()
		ui.Box(c).Size(7, 14).Background(p.accent).Margin(0, 1)
	})
}

func monoLine(c *ui.Context, p palette, spans ...ui.Span) {
	mono(ui.RichText(c, spans...)).NoWrap()
}

func blankLine(c *ui.Context) {
	ui.Box(c).Height(20)
}

func errorLine(c *ui.Context, p palette) {
	ui.Row(c).Height(24).FillWidth().MaxWidth(793).Margin(0, 0, 0, -6).Padding(0, 6, 0, 6).
		Background(p.errorBG).AlignItems(ui.Center).Children(func() {
		mono(ui.RichText(c,
			ui.Span{Text: "error", Color: p.red, Weight: 600},
			ui.Span{Text: " connect ECONNREFUSED 127.0.0.1:5432", Color: p.red},
		)).NoWrap()
		ui.Spacer(c)
		ui.Text(c, "retry 2/5").Font("Menlo").FontSize(10).TextColor(p.redMuted)
	})
}

func logLine(c *ui.Context, p palette, at, level string, levelColor ui.Color, msg string) {
	ui.Row(c).Height(19).AlignItems(ui.Center).Children(func() {
		ui.Text(c, at).Font("Menlo").FontSize(11).TextColor(p.ink3).Width(62)
		ui.Text(c, level).Font("Menlo").FontSize(11).TextColor(levelColor).Width(46)
		ui.Text(c, msg).Font("Menlo").FontSize(11.5).TextColor(p.ink2).SingleLine()
	})
}

func icon(c *ui.Context, s *ui.SVG, size float32, col ui.Color) {
	ui.Icon(c, s).Size(size, size).TextColor(col)
}

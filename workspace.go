package main

import (
	"os"
	"strings"

	"github.com/egoist/mygo/ui"
)

// emptyWorkspace is the content module that replaces the terminal surface
// while no session is open. It owns the whole content area, so closing the
// last terminal leaves no stale terminal view behind to hold input or focus.
func (a *app) emptyWorkspace(c *ui.Context, p palette) {
	ui.Column(c).FillWidth().FillHeight().Center().
		Padding(28, 40, 28, 40).Children(func() {
		ui.Column(c).Width(540).MaxWidthPercent(100).
			AlignItems(ui.Center).Children(func() {
			a.workspaceEmblem(c, p)
			ui.Box(c).Height(22)
			ui.Text(c, "还没有打开的终端").FontSize(21).FontWeight(600).
				TextColor(p.ink).TextAlign(ui.Center).SingleLine()
			ui.Box(c).Height(8)
			ui.Text(c, "新建一个本地终端，每个会话都在独立的 shell 中运行。").
				Width(390).FontSize(12).LineHeight(1.5).
				TextColor(p.ink2).TextAlign(ui.Center)
			if a.termErr != nil {
				ui.Box(c).Height(20)
				a.workspaceError(c, p)
			}
			ui.Box(c).Height(26)
			a.workspaceNewTerminal(c, p)
			ui.Box(c).Height(18)
			a.workspaceMeta(c, p)
			if dir := workspaceDir(); dir != "" {
				ui.Box(c).Height(10)
				ui.Text(c, "起始目录  "+dir).Font("Menlo").FontSize(10).
					TextColor(p.ink3).SingleLine()
			}
		})
	})
}

// workspaceEmblem is the small terminal window that anchors the empty
// state, in the same visual language as the app's brand mark.
func (a *app) workspaceEmblem(c *ui.Context, p palette) {
	ui.Column(c).Width(96).Height(74).Radius(14).Background(p.elevated).
		Border(1, p.border2).Shadow(0, 16, 36, 0, ui.RGBA(0, 0, 0, 0.22)).
		Clip().Children(func() {
		ui.Row(c).Height(21).Padding(0, 9, 0, 9).Gap(4).
			AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).
			BorderColor(p.border).Children(func() {
			dot(c, 4, p.redMuted)
			dot(c, 4, p.amber)
			dot(c, 4, p.accent)
		})
		ui.Box(c).Grow(1).Center().Children(func() {
			ui.Text(c, ">_").Font("Menlo").FontSize(18).FontWeight(600).
				TextColor(p.accent)
		})
	})
}

// workspaceNewTerminal is the primary action of the empty workspace. Its
// label is distinct from the sidebar's button so accessibility tools and
// tests can target each one.
func (a *app) workspaceNewTerminal(c *ui.Context, p palette) {
	btn := ui.ButtonBase(c.Key("empty-new-terminal")).Height(42).
		Padding(0, 18, 0, 18).Radius(9).Background(p.accent).Gap(9).
		AlignItems(ui.Center).Cursor(ui.CursorPointer).
		Label("在空白工作区新建终端").Tooltip("新建终端 ⌘T")
	if btn.Hovered() {
		btn.Background(p.mint)
	}
	btn.Children(func() {
		icon(c, svgPlus, 15, p.onAccent)
		ui.Text(c, "新建终端").FontSize(12.5).FontWeight(600).
			TextColor(p.onAccent).SingleLine()
		ui.Box(c).Width(6)
		ui.Text(c, "⌘T").FontSize(10.5).FontWeight(500).
			TextColor(p.onAccentMuted).SingleLine()
	})
	btn.OnClick(func() {
		a.termErr = nil
		a.newSession()
	})
}

func (a *app) workspaceMeta(c *ui.Context, p palette) {
	ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
		a.workspaceTag(c, p, "本地", true)
		a.workspaceTag(c, p, shellName(), false)
		a.workspaceTag(c, p, "独立 PTY", false)
	})
}

func (a *app) workspaceTag(c *ui.Context, p palette, label string, live bool) {
	ui.Row(c).Height(22).Padding(0, 8, 0, 8).Radius(5).
		Background(p.subtle).Border(1, p.border).Gap(5).
		AlignItems(ui.Center).Children(func() {
		if live {
			dot(c, 5, p.accent)
		}
		ui.Text(c, label).FontSize(10).TextColor(p.ink3).SingleLine()
	})
}

func (a *app) workspaceError(c *ui.Context, p palette) {
	ui.Row(c).Width(440).Padding(11, 13, 11, 13).Radius(8).
		Background(p.errorBG).Border(1, p.redMuted).Gap(10).
		AlignItems(ui.Start).Children(func() {
		ui.Box(c).Size(6, 6).Radius(3).Background(p.red).Margin(3, 0)
		ui.Column(c).Grow(1).Gap(3).Children(func() {
			ui.Text(c, "终端启动失败").FontSize(11.5).FontWeight(600).
				TextColor(p.red).SingleLine()
			ui.Text(c, a.termErr.Error()).Font("Menlo").FontSize(10.5).
				TextColor(p.ink2)
		})
	})
}

// workspaceDir is the directory a new terminal starts in, shortened for
// display.
func workspaceDir() string {
	dir, err := os.UserHomeDir()
	if err != nil || dir == "" {
		return ""
	}
	if dir == "/" {
		return dir
	}
	return strings.TrimSuffix(shortHome(dir), string(os.PathSeparator))
}

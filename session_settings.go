package main

import (
	"github.com/egoist/mygo/ui"
)

func (a *app) sessionSettingsButton(c *ui.Context, p palette, s *session) ui.Element {
	btn := ui.ButtonBase(c.Key("session-settings")).Size(22, 22).Radius(5).
		Center().Cursor(ui.CursorPointer).Label("当前终端设置").
		Tooltip("当前终端设置")
	if btn.Hovered() {
		btn.Background(p.subtle)
	}
	btn.Children(func() {
		icon(c, svgSettings, 15, p.ink3)
	})
	btn.OnClick(func() { a.openSessionSettings(s, "gear") })
	return btn
}

func (a *app) sessionSettingsPanel(c *ui.Context, p palette, s *session, panel ui.Element) {
	info := a.sessionLang(s)
	panel.Width(352).Padding(16, 16, 16, 16).Radius(12).
		Background(p.elevated).Border(1, p.border2).
		Shadow(0, 18, 44, 0, ui.RGBA(0, 0, 0, 0.32)).Children(func() {
		ui.Column(c).Gap(13).Children(func() {
			ui.Row(c).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "终端设置").FontSize(13).FontWeight(600).
					TextColor(p.ink).SingleLine()
				ui.Spacer(c)
				a.workspaceTag(c, p, "仅此会话", false)
			})
			ui.Box(c).Height(1).Background(p.border)
			a.toolchainSettings(c, p, s, info)
			ui.Box(c).Height(1).Background(p.border)
			a.sessionDetails(c, p, s)
		})
	})
}

func (a *app) toolchainSettings(c *ui.Context, p palette, s *session, info langInfo) {
	tool := a.sessionSettings.tool
	ui.Row(c).AlignItems(ui.Center).Children(func() {
		ui.Text(c, "工具链版本").FontSize(10.5).TextColor(p.ink3).
			SingleLine()
		ui.Spacer(c)
		if manager := a.ensureVersionManager(); manager != nil {
			a.workspaceTag(c, p, manager.Name(), false)
		} else {
			a.workspaceTag(c, p, "未检测到管理器", false)
		}
	})

	if tool == "" {
		ui.Text(c, "当前目录没有可识别的语言工具链。").FontSize(11).
			LineHeight(1.4).TextColor(p.ink2)
	} else {
		project := toolchainLabel(tool)
		detail := "跟随项目"
		if info.Version != "" {
			detail += " " + info.Version
		}
		selected := a.sessionSettings.draft == ""
		row := a.toolchainOption(c, p, "toolchain-follow", project, detail,
			selected, "工具链跟随项目配置")
		row.OnClick(func() {
			a.sessionSettings.draft = ""
			a.sessionSettings.err = ""
			a.invalidate()
		})

		switch {
		case a.sessionSettings.loading:
			ui.Row(c).Height(30).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Spinner(c).Size(14, 14).Label("正在读取已安装版本")
				ui.Text(c, "正在读取已安装版本…").FontSize(10.5).
					TextColor(p.ink3)
			})
		case len(a.sessionSettings.versions) > 0:
			ui.Scroll(c).MaxHeight(186).Children(func() {
				ui.Column(c).Gap(3).Children(func() {
					for _, version := range a.sessionSettings.versions {
						row := a.toolchainVersion(c, p, version,
							a.sessionSettings.draft == version)
						row.OnClick(func() {
							a.sessionSettings.draft = version
							a.sessionSettings.err = ""
							a.invalidate()
						})
					}
				})
			})
		}
		a.toolchainCustomVersion(c, p)
	}

	if a.sessionSettings.err != "" {
		ui.Text(c, a.sessionSettings.err).FontSize(10.5).
			LineHeight(1.35).TextColor(p.red)
	}

	changed := a.toolchainChanged(s)
	ui.Row(c).AlignItems(ui.Center).Children(func() {
		note := "重启终端后生效"
		if !changed {
			note = "当前设置"
		}
		ui.Text(c, note).FontSize(10).TextColor(p.ink3)
		ui.Spacer(c)
		apply := ui.ButtonBase(c.Key("toolchain-apply")).Height(30).
			Padding(0, 13, 0, 13).Radius(6).Background(p.accent).
			AlignItems(ui.Center).Cursor(ui.CursorPointer).
			Label("应用并重启终端").Disabled(!changed)
		if apply.Hovered() && changed {
			apply.Background(p.mint)
		}
		apply.Children(func() {
			ui.Text(c, "应用并重启").FontSize(11).FontWeight(600).
				TextColor(p.onAccent).SingleLine()
		})
		apply.OnClick(func() { a.applySessionToolchain(s) })
	})
}

func (a *app) toolchainOption(c *ui.Context, p palette, key, name, detail string, selected bool, label string) ui.Element {
	row := ui.ButtonBase(c.Key(key)).Height(32).Padding(0, 9, 0, 9).
		Radius(6).Gap(8).AlignItems(ui.Center).Cursor(ui.CursorPointer).
		Label(label)
	switch {
	case selected:
		row.Background(p.selected).Border(1, p.selectedBorder)
	case row.Hovered():
		row.Background(p.subtle)
	}
	row.Children(func() {
		a.selectionDot(c, p, selected)
		ui.Text(c, name).FontSize(11).FontWeight(500).TextColor(p.ink).
			SingleLine()
		ui.Spacer(c)
		ui.Text(c, detail).FontSize(9.5).TextColor(p.ink3).SingleLine()
	})
	return row
}

func (a *app) toolchainVersion(c *ui.Context, p palette, version string, selected bool) ui.Element {
	row := ui.ButtonBase(c.Key("toolchain-version-"+version)).Height(30).
		Padding(0, 9, 0, 9).Radius(6).Gap(8).
		AlignItems(ui.Center).Cursor(ui.CursorPointer).
		Label("工具链版本 " + version)
	switch {
	case selected:
		row.Background(p.selected).Border(1, p.selectedBorder)
	case row.Hovered():
		row.Background(p.subtle)
	}
	row.Children(func() {
		a.selectionDot(c, p, selected)
		ui.Text(c, version).Font("Menlo").FontSize(10.5).
			TextColor(p.ink).SingleLine()
		ui.Spacer(c)
		if selected {
			ui.Text(c, "已选择").FontSize(9.5).TextColor(p.accent).SingleLine()
		} else {
			ui.Text(c, "已安装").FontSize(9.5).TextColor(p.ink3).SingleLine()
		}
	})
	return row
}

func (a *app) selectionDot(c *ui.Context, p palette, selected bool) {
	ui.Box(c).Size(12, 12).Radius(6).
		Border(1, func() ui.Color {
			if selected {
				return p.accent
			}
			return p.border2
		}()).Center().Children(func() {
		if selected {
			ui.Box(c).Size(6, 6).Radius(3).Background(p.accent)
		}
	})
}

func (a *app) toolchainCustomVersion(c *ui.Context, p palette) {
	ui.Row(c).Height(32).Padding(0, 8, 0, 8).Radius(6).Gap(6).
		Background(p.field).Border(1, p.border).AlignItems(ui.Center).
		Children(func() {
			ui.TextInputBase(c, &a.sessionSettings.customText).Grow(1).
				Label("自定义工具链版本").
				Placeholder("输入已安装版本，例如 16").
				PlaceholderColor(p.ink3).TextColor(p.ink).
				OnSubmit(func() { a.useCustomToolchainVersion() })
			use := ui.ButtonBase(c.Key("toolchain-custom-use")).Height(24).
				Padding(0, 9, 0, 9).Radius(4).Background(p.subtle).
				Border(1, p.border).Center().Cursor(ui.CursorPointer).
				Label("使用自定义工具链版本")
			if use.Hovered() {
				use.Border(1, p.border2)
			}
			use.Children(func() {
				ui.Text(c, "使用").FontSize(10).TextColor(p.ink2).SingleLine()
			})
			use.OnClick(func() { a.useCustomToolchainVersion() })
		})
}

func (a *app) sessionDetails(c *ui.Context, p palette, s *session) {
	detail := func(name, value string) {
		ui.Row(c).Height(20).AlignItems(ui.Center).Children(func() {
			ui.Text(c, name).FontSize(10.5).TextColor(p.ink3).SingleLine()
			ui.Spacer(c)
			ui.Text(c, value).Font("Menlo").FontSize(10).
				TextColor(p.ink2).SingleLine()
		})
	}
	detail("Shell", shellName())
	detail("工作目录", a.sessionPath(s))
}

func (a *app) toolchainChanged(s *session) bool {
	draft := a.sessionSettings.draft
	if s.toolchain == nil {
		return draft != ""
	}
	return s.toolchain.Tool != a.sessionSettings.tool ||
		s.toolchain.Version != draft
}

package main

import (
	"regexp"
	"strings"

	"github.com/egoist/mygo/ui"
)

type closeAction int

const (
	closeTerminal closeAction = iota
	restartTerminal
)

// requestCloseSession asks before a running terminal is closed. The
// terminal package does not expose its child process group, so a running
// session always needs confirmation. Exited sessions and sessions without
// an emulator close at once.
func (a *app) requestCloseSession(id int) {
	a.requestSessionClose(id, closeTerminal, nil)
}

func (a *app) requestRestartSession(id int, pin *sessionToolchain) {
	a.requestSessionClose(id, restartTerminal, pin)
}

func (a *app) requestSessionClose(id int, action closeAction, pin *sessionToolchain) {
	s := a.sessionByID(id)
	if s == nil {
		return
	}
	a.sessionSettings.open = false
	if !s.running() {
		if action == restartTerminal {
			a.restartSession(id, pin)
		} else {
			a.closeSession(id)
		}
		return
	}
	a.renaming = false
	a.confirmClose = true
	a.confirmCloseID = id
	a.confirmAction = action
	a.restartPin = pin
	a.invalidate()
}

func (a *app) confirmCloseSession() {
	id := a.confirmCloseID
	action := a.confirmAction
	pin := a.restartPin
	a.confirmClose = false
	a.confirmCloseID = 0
	a.confirmAction = closeTerminal
	a.restartPin = nil
	if action == restartTerminal {
		a.restartSession(id, pin)
		return
	}
	a.closeSession(id)
}

func (a *app) cancelCloseSession() {
	a.confirmClose = false
	a.confirmCloseID = 0
	a.confirmAction = closeTerminal
	a.restartPin = nil
	a.invalidate()
}

// closeConfirmDialog warns that closing a terminal ends its tasks. The
// service hint only changes the wording: without a process-level view of
// the pseudo-terminal, every running session is confirmed.
func (a *app) closeConfirmDialog(c *ui.Context, p palette) {
	if !a.confirmClose {
		return
	}
	s := a.sessionByID(a.confirmCloseID)
	if s == nil {
		return
	}
	service := false
	if s.term != nil {
		service = serviceHintFromText(s.term.Text())
	}
	restart := a.confirmAction == restartTerminal
	ui.DialogBase(c, &a.confirmClose, func(backdrop, panel ui.Element) {
		backdrop.Background(ui.RGBA(0, 0, 0, 0.55))
		panel.Width(436).Padding(20, 20, 20, 20).Radius(12).
			Background(p.elevated).Border(1, p.border2).Gap(16).Children(func() {
			ui.Row(c).AlignItems(ui.Start).Gap(12).Children(func() {
				ui.Box(c).Size(36, 36).Radius(10).Background(p.errorBG).
					Border(1, p.redMuted).Center().Children(func() {
					icon(c, svgAlert, 18, p.red)
				})
				ui.Column(c).Grow(1).Gap(5).Children(func() {
					title := "关闭这个终端？"
					if restart {
						title = "应用新版本并重启？"
					}
					ui.Text(c, title).FontSize(14.5).FontWeight(600).
						TextColor(p.ink).SingleLine()
					message := "关闭终端会结束其中正在进行的所有任务。如果还有后台服务在运行，也会一并终止。"
					if restart {
						message = "重启终端会结束其中正在进行的所有任务。如果还有后台服务在运行，也会一并终止。"
					}
					if service {
						if restart {
							message = "终端输出中出现了服务启动信息。重启后，这些服务及进行中的任务都会被立即终止。"
						} else {
							message = "终端输出中出现了服务启动信息。关闭后，这些服务及进行中的任务都会被立即终止。"
						}
					}
					ui.Text(c, message).FontSize(11.5).LineHeight(1.45).
						TextColor(p.ink2)
				})
			})
			a.closeTarget(c, p, s, service)
			if restart {
				next := "重启后将跟随项目配置"
				if a.restartPin != nil {
					next = "重启后使用 " + toolchainLabel(a.restartPin.Tool) + " " + a.restartPin.Version
				}
				ui.Text(c, next).FontSize(10.5).TextColor(p.ink2).SingleLine()
			}
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				cancel := a.dialogButton(c, p, "close-cancel", "取消", false)
				cancel.OnClick(func() { a.cancelCloseSession() })
				label, access := "关闭终端", "确认关闭终端"
				if restart {
					label, access = "重启终端", "确认重启终端"
				}
				confirm := a.destructiveButton(c, p, "close-confirm", label, access)
				confirm.OnClick(func() { a.confirmCloseSession() })
			})
		})
	})
}

func (a *app) closeTarget(c *ui.Context, p palette, s *session, service bool) {
	ui.Row(c).Height(48).Padding(0, 12, 0, 12).Radius(8).
		Background(p.field).Border(1, p.border).Gap(10).
		AlignItems(ui.Center).Children(func() {
		dot(c, 7, sessionDot(p, s))
		ui.Column(c).Grow(1).Gap(2).Children(func() {
			ui.Text(c, s.name).FontSize(11.5).FontWeight(500).
				TextColor(p.ink).SingleLine()
			ui.Text(c, a.sessionPath(s)).Font("Menlo").FontSize(10).
				TextColor(p.ink3).SingleLine()
		})
		if service {
			ui.Row(c).Height(20).Padding(0, 7, 0, 7).Radius(4).
				Background(ui.RGBA(p.amber.R, p.amber.G, p.amber.B, 0.14)).
				Center().Children(func() {
				ui.Text(c, "可能仍有服务运行").FontSize(9.5).
					TextColor(p.amber).SingleLine()
			})
			return
		}
		ui.Text(c, "运行中").FontSize(10).TextColor(p.accent).SingleLine()
	})
}

func (a *app) destructiveButton(c *ui.Context, p palette, key, label, accessLabel string) ui.Element {
	btn := ui.ButtonBase(c.Key(key)).Height(30).Padding(0, 14, 0, 14).
		Radius(6).Background(p.red).AlignItems(ui.Center).
		Cursor(ui.CursorPointer).Label(accessLabel)
	if btn.Hovered() {
		btn.Background(ui.RGBA(p.red.R, p.red.G, p.red.B, 0.86))
	}
	btn.Children(func() {
		ui.Text(c, label).FontSize(11.5).FontWeight(500).
			TextColor(ui.RGB(255, 255, 255)).SingleLine()
	})
	return btn
}

// serviceOutput finds common signs that a long-running server or watcher
// is up in the terminal's recent output. It is a heuristic for the dialog
// copy only, never a reason to skip confirmation.
var serviceOutput = regexp.MustCompile(`(?i)(` +
	`listening on|ready on|server (?:is )?(?:running|started)|running at|` +
	`dev server|watching for (?:file )?changes|compiled successfully|` +
	`https?://(?:localhost|127\.0\.0\.1|0\.0\.0\.0|\[::1\]):\d+|` +
	`(?:localhost|127\.0\.0\.1|0\.0\.0\.0|\[::1\]):\d+` +
	`)`)

func serviceHintFromText(text string) bool {
	if len(text) > 8192 {
		text = text[len(text)-8192:]
	}
	lines := strings.Split(strings.TrimSpace(text), "\n")
	for i := len(lines) - 1; i >= 0 && i >= len(lines)-24; i-- {
		if serviceOutput.MatchString(lines[i]) {
			return true
		}
	}
	return false
}

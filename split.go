package main

import (
	"slices"
	"strconv"
	"strings"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// sessionDrag is dragged from the session list into the content area.
type sessionDrag struct {
	ID int
}

type dropEdge uint8

const (
	dropLeft dropEdge = iota
	dropRight
	dropTop
	dropBottom
)

type paneDropPreview struct {
	edge     dropEdge
	anchorID int
	column   int
	columns  int
	width    float32
	height   float32
}

// visiblePaneIDs returns the sessions shown in the content area, in row
// order. A one-pane layout is represented by no rows, so the normal
// full-width terminal remains the default.
func (a *app) visiblePaneIDs() []int {
	var ids []int
	seen := make(map[int]bool)
	for _, row := range a.paneRows {
		for _, id := range row {
			s := a.sessionByID(id)
			if s == nil || s.term == nil || seen[id] {
				continue
			}
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids
}

func (a *app) splitting() bool {
	return len(a.visiblePaneIDs()) > 1 && a.inSplit(a.activeID)
}

func (a *app) inSplit(id int) bool {
	for _, row := range a.paneRows {
		if slices.Contains(row, id) {
			return true
		}
	}
	return false
}

func (a *app) paneLocation(id int) (row, column int) {
	for r, ids := range a.paneRows {
		for c, paneID := range ids {
			if paneID == id {
				return r, c
			}
		}
	}
	return -1, -1
}

func (a *app) paneLayoutKey() string {
	var b strings.Builder
	for _, row := range a.paneRows {
		b.WriteByte('/')
		for i, id := range row {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(strconv.Itoa(id))
		}
	}
	return b.String()
}

func (a *app) setPaneRows(rows [][]int) {
	rows = compactPaneRows(rows)
	if len(flattenPaneRows(rows)) <= 1 {
		a.paneRows = nil
		return
	}
	a.paneRows = rows
}

// showSession makes id active. Sessions inside the saved split layout show
// that layout; any other session opens on its own full-screen surface.
func (a *app) showSession(id int) {
	if a.sessionByID(id) == nil {
		return
	}
	a.activeID = id
	a.focusPaneID = id
	a.sessionSettings.open = false
	a.saveState()
	a.invalidate()
}

func (a *app) focusSplitPane(id int) {
	if !a.inSplit(id) {
		return
	}
	a.activeID = id
	a.focusPaneID = id
	a.sessionSettings.open = false
	a.saveState()
	a.invalidate()
}

func (a *app) startSplit(otherID int) {
	current := a.activeSession()
	other := a.sessionByID(otherID)
	if current == nil || other == nil || current.id == other.id ||
		current.term == nil || other.term == nil {
		return
	}
	a.paneRows = [][]int{{current.id, other.id}}
	a.focusPaneID = current.id
	a.sessionSettings.open = false
	a.invalidate()
}

func (a *app) splitWithNewSession() {
	current := a.activeSession()
	next := a.addSession()
	if next == nil {
		a.invalidate()
		return
	}
	if current == nil {
		a.showSession(next.id)
		return
	}
	a.paneRows = [][]int{{current.id, next.id}}
	a.activeID = next.id
	a.focusPaneID = next.id
	a.sessionSettings.open = false
	a.saveState()
	a.invalidate()
}

func (a *app) stopSplit() {
	if len(a.paneRows) == 0 {
		return
	}
	a.paneRows = nil
	if a.activeSession() == nil && len(a.sessions) > 0 {
		a.activeID = a.sessions[len(a.sessions)-1].id
	}
	a.focusPaneID = a.activeID
	a.saveState()
	a.invalidate()
}

// removePaneSession removes id from the layout but leaves the shell open.
// If the focused pane goes, focus moves to its neighbor.
func (a *app) removePaneSession(id int) []int {
	before := a.visiblePaneIDs()
	index := slices.Index(before, id)
	rows := removeSessionFromRows(a.paneRows, id)
	after := flattenPaneRows(compactPaneRows(rows))
	a.setPaneRows(rows)
	if a.activeID == id {
		a.activeID = 0
		if len(after) > 0 {
			if index < 0 {
				index = 0
			}
			if index >= len(after) {
				index = len(after) - 1
			}
			a.activeID = after[index]
		}
		a.focusPaneID = a.activeID
	}
	a.saveState()
	return after
}

func (a *app) removeSplitPane(id int) {
	if !a.inSplit(id) {
		return
	}
	a.removePaneSession(id)
	a.invalidate()
}

func (a *app) toggleSplit() {
	if a.splitting() {
		a.stopSplit()
		return
	}
	for i := len(a.sessions) - 1; i >= 0; i-- {
		if a.sessions[i].id != a.activeID {
			a.startSplit(a.sessions[i].id)
			return
		}
	}
	a.splitWithNewSession()
}

// placeDraggedSession inserts id beside anchor, or in a full-width row
// above or below it. Dragging a pane that is already visible moves it.
func (a *app) placeDraggedSession(id, anchorID int, edge dropEdge) {
	s := a.sessionByID(id)
	if s == nil || s.term == nil {
		return
	}
	if id == anchorID {
		a.focusSplitPane(id)
		return
	}

	current := a.activeSession()
	if current == nil {
		return
	}
	rows := [][]int{{current.id}}
	if a.splitting() {
		rows = clonePaneRows(a.paneRows)
	}
	rows = removeSessionFromRows(rows, id)
	row, column := locatePane(rows, anchorID)
	if row < 0 {
		rows = append(rows, []int{id})
	} else {
		switch edge {
		case dropLeft:
			rows[row] = insertPane(rows[row], column, id)
		case dropRight:
			rows[row] = insertPane(rows[row], column+1, id)
		case dropTop:
			rows = insertPaneRow(rows, row, []int{id})
		case dropBottom:
			rows = insertPaneRow(rows, row+1, []int{id})
		}
	}
	a.setPaneRows(rows)
	a.activeID = id
	a.focusPaneID = id
	a.sessionSettings.open = false
	a.saveState()
	a.invalidate()
}

func (a *app) splitButton(c *ui.Context, p palette, s *session) ui.Element {
	btn := ui.ButtonBase(c.Key("session-split")).Size(22, 22).Radius(5).
		Center().Cursor(ui.CursorPointer).Label("分屏显示").
		Tooltip("分屏显示 ⌘D")
	if a.splitting() {
		btn.Background(p.subtle)
	}
	if btn.Hovered() {
		btn.Background(p.subtle)
	}
	btn.Children(func() {
		icon(c, svgPanelRight, 15, p.ink2)
	})
	btn.PopoverMenuButton(func(m *ui.Menu) {
		a.splitMenu(c, s, m)
	}, ui.PopoverMenuButtonOptions{
		At: ui.AnchorBottomEnd, Self: ui.AnchorTopEnd,
	})
	return btn
}

func (a *app) splitToggleButton(c *ui.Context, p palette) ui.Element {
	btn := ui.ButtonBase(c.Key("split-toggle")).Size(22, 22).Radius(5).
		Center().Cursor(ui.CursorPointer).Label("切换分屏布局").
		Tooltip("切换分屏布局 ⌘D")
	if a.splitting() {
		btn.Background(p.subtle)
	}
	if btn.Hovered() {
		btn.Background(p.subtle)
	}
	btn.Children(func() {
		icon(c, svgPanelRight, 17, p.ink2)
	})
	btn.OnClick(func() { a.toggleSplit() })
	return btn
}

func (a *app) splitMenu(c *ui.Context, s *session, m *ui.Menu) {
	if a.splitting() {
		if m.Item("退出分屏").Chosen() {
			a.stopSplit()
		}
		m.Separator()
	}
	if m.Item("新建终端并分屏").Chosen() {
		a.splitWithNewSession()
	}
	others := 0
	for _, other := range a.sessions {
		if other.id == s.id || other.term == nil {
			continue
		}
		item := m.Item("与 " + other.name + " 分屏")
		if a.splitting() && a.inSplit(other.id) {
			item.Checked(true)
		}
		if item.Chosen() {
			a.startSplit(other.id)
		}
		others++
	}
	if others > 0 {
		m.Separator()
	}
	if others == 0 {
		m.Item("暂无可分屏的终端").Disabled(true)
	}
}

func (a *app) splitPanes(c *ui.Context, p palette) {
	rows := clonePaneRows(a.paneRows)
	ui.Column(c).Fill().Background(p.paneBorder).Children(func() {
		for i, ids := range rows {
			if i > 0 {
				ui.Box(c).FillWidth().Height(1).Background(p.paneBorder)
			}
			a.splitRow(c, p, i, ids)
		}
	})
}

func (a *app) splitRow(c *ui.Context, p palette, index int, ids []int) {
	row := ui.Row(c).Key("split-row-" + strconv.Itoa(index)).
		Grow(1).FillWidth().AlignItems(ui.Stretch).Clip()
	preview, dropped := a.paneDropFeedback(row, ids)
	row.Children(func() {
		for i, id := range ids {
			if i > 0 {
				ui.Box(c).Width(1).FillHeight().Background(p.paneBorder)
			}
			s := a.sessionByID(id)
			if s != nil && s.term != nil {
				a.splitPane(c, p, s)
			}
		}
		if preview != nil {
			a.dropPreview(c, p, preview)
		}
	})
	if dropped != nil && preview != nil {
		a.placeDraggedSession(dropped.ID, preview.anchorID, preview.edge)
	}
}

func (a *app) splitPane(c *ui.Context, p palette, s *session) {
	active := s.id == a.activeID
	pane := ui.Column(c).Key("split-pane-" + strconv.Itoa(s.id)).
		Grow(1).FillHeight().MinWidth(0).Background(p.terminal)
	if active {
		pane.Border(1, p.accent)
	} else {
		pane.Border(1, p.paneBorder)
	}
	pane.Children(func() {
		a.splitPaneHeader(c, p, s, active)
		view := terminal.View(c, s.term).Grow(1).Label(s.name + " 终端面板")
		if a.focusPaneID == s.id {
			view = view.Focus()
			a.focusPaneID = 0
		}
		view.OnClick(func() { a.focusSplitPane(s.id) })
	})
}

func (a *app) singlePane(c *ui.Context, p palette, s *session) {
	pane := ui.Column(c).Key("single-pane-" + strconv.Itoa(s.id)).
		Fill().Background(p.terminal).Clip()
	preview, dropped := a.paneDropFeedback(pane, []int{s.id})
	pane.Children(func() {
		terminal.View(c, s.term).Fill().Label(s.name + " 终端面板").AutoFocus()
		if preview != nil {
			a.dropPreview(c, p, preview)
		}
	})
	if dropped != nil && preview != nil {
		a.placeDraggedSession(dropped.ID, preview.anchorID, preview.edge)
	}
}

func (a *app) paneDropFeedback(target ui.Element, ids []int) (*paneDropPreview, *sessionDrag) {
	value, over := ui.DragOver[sessionDrag](target)
	dropped, wasDropped := ui.Drop[sessionDrag](target)
	if !over && !wasDropped {
		return nil, nil
	}
	x, y, _ := target.PointerPosition()
	bounds := target.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 || len(ids) == 0 {
		return nil, nil
	}
	preview := &paneDropPreview{
		edge:     nearestDropEdge(x, y, bounds.W, bounds.H),
		column:   paneColumnAt(x, bounds.W, len(ids)),
		columns:  len(ids),
		width:    bounds.W,
		height:   bounds.H,
		anchorID: ids[paneColumnAt(x, bounds.W, len(ids))],
	}
	if wasDropped {
		return preview, &dropped
	}
	if value.ID == 0 {
		return nil, nil
	}
	return preview, nil
}

func (a *app) dropPreview(c *ui.Context, p palette, preview *paneDropPreview) {
	const gap = float32(1)
	paneWidth := (preview.width - gap*float32(preview.columns-1)) / float32(preview.columns)
	if paneWidth < 0 {
		paneWidth = preview.width
	}
	x, y := float32(0), float32(0)
	width, height := preview.width, preview.height
	switch preview.edge {
	case dropLeft:
		x = float32(preview.column) * (paneWidth + gap)
		width = paneWidth / 2
	case dropRight:
		x = float32(preview.column)*(paneWidth+gap) + paneWidth/2
		width = paneWidth / 2
	case dropTop:
		height = preview.height / 2
	case dropBottom:
		y = preview.height / 2
		height = preview.height / 2
	}
	const inset = float32(3)
	width = max(width-inset*2, 0)
	height = max(height-inset*2, 0)
	if width == 0 || height == 0 {
		return
	}
	ui.Box(c).Absolute().Left(x+inset).Top(y+inset).
		Width(width).Height(height).Radius(7).Role(ui.RoleNone).
		Background(ui.RGBA(61, 220, 151, 0.16)).
		Border(1, p.accent).Center().Label(a.dropPreviewLabel(preview)).
		Children(func() {
			ui.Text(c, "+").Font("Menlo").FontSize(18).FontWeight(600).
				TextColor(p.mint).SingleLine()
		})
}

func (a *app) dropPreviewLabel(preview *paneDropPreview) string {
	return "放置终端到" + dropEdgeLabel(preview.edge)
}

func dropEdgeLabel(edge dropEdge) string {
	switch edge {
	case dropLeft:
		return "左侧"
	case dropRight:
		return "右侧"
	case dropTop:
		return "上方"
	default:
		return "下方"
	}
}

func (a *app) splitPaneHeader(c *ui.Context, p palette, s *session, active bool) {
	row := ui.Row(c).Height(30).Padding(0, 6, 0, 6).Gap(4).
		AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).
		BorderColor(p.border)
	if active {
		row.Background(p.selected)
	} else {
		row.Background(p.paneHeader)
	}
	row.Children(func() {
		focus := ui.ButtonBase(c.Key("split-focus-" + strconv.Itoa(s.id))).
			Grow(1).Height(29).Gap(7).AlignItems(ui.Center).
			Cursor(ui.CursorPointer).Label("切换到 " + s.name + " 面板").
			Tooltip("切换到 " + s.name + " 面板")
		focus.Children(func() {
			dot(c, 6, sessionDot(p, s))
			col := p.ink2
			if active {
				col = p.ink
			}
			ui.Text(c, s.name).FontSize(11).FontWeight(500).
				TextColor(col).SingleLine()
			ui.Spacer(c)
			state := "已分屏"
			if active {
				state = "当前"
			}
			ui.Text(c, state).FontSize(9).TextColor(p.ink3).SingleLine()
		})
		focus.OnClick(func() { a.focusSplitPane(s.id) })

		remove := ui.ButtonBase(c.Key("split-remove-"+strconv.Itoa(s.id))).
			Size(20, 20).Radius(4).Center().Cursor(ui.CursorPointer).
			Label("从分屏中移除 " + s.name).
			Tooltip("从分屏中移除 " + s.name)
		if remove.Hovered() {
			remove.Background(p.border2)
		}
		remove.Children(func() {
			icon(c, svgX, 11, p.ink3)
		})
		remove.OnClick(func() { a.removeSplitPane(s.id) })
	})
}

func nearestDropEdge(x, y, width, height float32) dropEdge {
	distance := [...]float32{x, width - x, y, height - y}
	edge := dropLeft
	for i := 1; i < len(distance); i++ {
		if distance[i] < distance[edge] {
			edge = dropEdge(i)
		}
	}
	return edge
}

func paneColumnAt(x, width float32, count int) int {
	if count <= 1 || width <= 0 {
		return 0
	}
	column := int(x / (width / float32(count)))
	if column < 0 {
		return 0
	}
	if column >= count {
		return count - 1
	}
	return column
}

func clonePaneRows(rows [][]int) [][]int {
	out := make([][]int, len(rows))
	for i, row := range rows {
		out[i] = slices.Clone(row)
	}
	return out
}

func compactPaneRows(rows [][]int) [][]int {
	out := make([][]int, 0, len(rows))
	seen := make(map[int]bool)
	for _, row := range rows {
		next := make([]int, 0, len(row))
		for _, id := range row {
			if id == 0 || seen[id] {
				continue
			}
			seen[id] = true
			next = append(next, id)
		}
		if len(next) > 0 {
			out = append(out, next)
		}
	}
	return out
}

func flattenPaneRows(rows [][]int) []int {
	var ids []int
	for _, row := range rows {
		ids = append(ids, row...)
	}
	return ids
}

func locatePane(rows [][]int, id int) (row, column int) {
	for r, ids := range rows {
		for c, paneID := range ids {
			if paneID == id {
				return r, c
			}
		}
	}
	return -1, -1
}

func removeSessionFromRows(rows [][]int, id int) [][]int {
	out := make([][]int, 0, len(rows))
	seen := make(map[int]bool)
	for _, row := range rows {
		next := make([]int, 0, len(row))
		for _, paneID := range row {
			if paneID == id || paneID == 0 || seen[paneID] {
				continue
			}
			seen[paneID] = true
			next = append(next, paneID)
		}
		if len(next) > 0 {
			out = append(out, next)
		}
	}
	return out
}

func insertPane(row []int, at int, id int) []int {
	at = min(max(at, 0), len(row))
	row = append(row, 0)
	copy(row[at+1:], row[at:])
	row[at] = id
	return row
}

func insertPaneRow(rows [][]int, at int, row []int) [][]int {
	at = min(max(at, 0), len(rows))
	rows = append(rows, nil)
	copy(rows[at+1:], rows[at:])
	rows[at] = row
	return rows
}

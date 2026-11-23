package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/egoist/mygo/ui"
)

// searchState is the Spotlight-style session switcher. List keeps the
// scroll position and keyboard choice across builds of the overlay.
type searchState struct {
	open      bool
	query     string
	selected  int
	scroll    bool
	lastQuery string
	list      ui.ListState
}

func (s *searchState) reset() {
	s.open = false
	s.query = ""
	s.selected = 0
	s.scroll = false
	s.lastQuery = ""
	s.list = ui.ListState{}
}

func (a *app) openSearch() {
	a.search.reset()
	a.search.open = true
	a.search.scroll = true
	a.invalidate()
}

func (a *app) closeSearch() {
	a.search.reset()
	a.invalidate()
}

// searchMatches ranks sessions whose name or working directory contains
// every word in the query. Name matches rank above path-only matches, and
// an exact name ranks first.
func (a *app) searchMatches(query string) []int {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	words := strings.Fields(q)
	type hit struct {
		id    int
		score int
	}
	var hits []hit
	for _, s := range a.sessions {
		name := strings.ToLower(s.name)
		path := strings.ToLower(a.sessionPath(s))
		dir := strings.ToLower(a.sessionDir(s))
		score := 0
		matched := true
		for _, word := range words {
			inName := strings.Contains(name, word)
			inPath := strings.Contains(path, word) || dir != "" && strings.Contains(dir, word)
			if !inName && !inPath {
				matched = false
				break
			}
			switch {
			case inName && strings.HasPrefix(name, word):
				score += 20
			case inName:
				score += 10
			default:
				score += 5
			}
		}
		if !matched {
			continue
		}
		switch {
		case name == q:
			score += 100
		case strings.HasPrefix(name, q):
			score += 50
		case strings.Contains(name, q):
			score += 25
		case strings.Contains(path, q) || dir != "" && strings.Contains(dir, q):
			score += 10
		}
		hits = append(hits, hit{id: s.id, score: score})
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	ids := make([]int, len(hits))
	for i, h := range hits {
		ids[i] = h.id
	}
	return ids
}

func (a *app) moveSearchSelection(delta int) {
	ids := a.searchMatches(a.search.query)
	if len(ids) == 0 {
		return
	}
	if a.search.selected < 0 || a.search.selected >= len(ids) {
		a.search.selected = 0
	} else {
		a.search.selected = (a.search.selected + delta + len(ids)) % len(ids)
	}
	a.search.scroll = true
	a.invalidate()
}

func (a *app) activateSearchResult(index int) {
	ids := a.searchMatches(a.search.query)
	if index < 0 || index >= len(ids) {
		return
	}
	id := ids[index]
	a.closeSearch()
	a.selectSession(id)
}

func (a *app) searchDialog(c *ui.Context, p palette) {
	if !a.search.open {
		return
	}
	query := strings.TrimSpace(a.search.query)
	if a.search.lastQuery != a.search.query {
		a.search.lastQuery = a.search.query
		a.search.selected = 0
		a.search.scroll = true
		a.search.list = ui.ListState{}
	}
	ids := a.searchMatches(a.search.query)
	if len(ids) == 0 {
		a.search.selected = 0
	} else {
		a.search.selected = max(0, min(a.search.selected, len(ids)-1))
	}

	veil := ui.RGBA(0, 0, 0, 0.46)
	shadow := ui.RGBA(0, 0, 0, 0.5)
	if a.light {
		veil = ui.RGBA(18, 23, 31, 0.24)
		shadow = ui.RGBA(15, 23, 42, 0.2)
	}

	panel := ui.DialogBase(c, &a.search.open, func(backdrop, panel ui.Element) {
		backdrop.Background(veil).Justify(ui.Start).AlignItems(ui.Center)
		panel.Width(560).Margin(88, 0, 0, 0).Radius(14).
			Background(p.elevated).Border(1, p.border2).
			Shadow(0, 20, 56, 0, shadow).Clip().Children(func() {
			ui.Row(c).Height(54).Padding(0, 16, 0, 16).Gap(12).
				AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).
				BorderColor(p.border).Children(func() {
				icon(c, svgSearch, 18, p.ink2)
				input := ui.TextInputBase(c.Key("session-search-input"), &a.search.query).
					Grow(1).FontSize(14.5).TextColor(p.ink).
					Placeholder("搜索会话或路径").PlaceholderColor(p.ink3).
					Label("搜索会话或路径").Search().AutoFocus()
				input.OnTextKey(func(k ui.TextKey) bool {
					switch {
					case k.Key == ui.KeyDown && k.Mods == 0:
						a.moveSearchSelection(1)
						return true
					case k.Key == ui.KeyUp && k.Mods == 0:
						a.moveSearchSelection(-1)
						return true
					case k.Key == ui.KeyEnter && k.Mods == 0:
						a.activateSearchResult(a.search.selected)
						return true
					case k.Key == ui.KeyEscape && k.Mods == 0:
						a.closeSearch()
						return true
					}
					return false
				})
				keycap(c, p, "esc")
			})
			switch {
			case query == "":
				ui.Column(c).Padding(22, 20, 22, 22).Gap(6).Children(func() {
					ui.Text(c, "输入会话名称或路径开始搜索").FontSize(12.5).
						FontWeight(500).TextColor(p.ink2)
					ui.Text(c, "支持按会话名称和当前工作目录匹配").FontSize(10.5).
						TextColor(p.ink3)
				})
			case len(ids) == 0:
				a.searchNoResults(c, p)
			default:
				a.searchResults(c, p, ids)
			}
			if query != "" {
				a.searchFooter(c, p, len(ids))
			}
		})
	})
	if panel.Dismissed() {
		a.search.reset()
	}
}

func (a *app) searchResults(c *ui.Context, p palette, ids []int) {
	a.search.list.Key = func(row int) any { return ids[row] }
	a.search.list.Label = func(row int) string {
		s := a.sessionByID(ids[row])
		if s == nil {
			return ""
		}
		return s.name + " " + a.sessionPath(s)
	}
	height := min(340, len(ids)*54+10)
	list := ui.List(c, &a.search.list, len(ids)).Height(float32(height)).
		Padding(6, 6).Gap(2)
	list.Rows(func(row ui.ListRow) {
		s := a.sessionByID(ids[row.Index])
		if s == nil {
			return
		}
		selected := row.Index == a.search.selected
		rc := row.Context
		item := ui.Row(rc).Height(52).FillWidth().Padding(0, 10, 0, 10).
			Radius(8).Gap(10).AlignItems(ui.Center)
		switch {
		case selected:
			item.Background(p.selected).Border(1, p.selectedBorder)
		case row.Hovered():
			item.Background(p.field)
		}
		item.Children(func() {
			dot(rc, 7, sessionDot(p, s))
			ui.Column(rc).Grow(1).Gap(2).Children(func() {
				ui.Row(rc).AlignItems(ui.Center).Gap(8).Children(func() {
					ui.Text(rc, s.name).FontSize(12.5).FontWeight(500).
						TextColor(p.ink).SingleLine()
					if s.id == a.activeID {
						a.searchBadge(rc, p, "当前", p.mint, p.runBG)
					} else if badge := sessionToolchainBadge(s); badge != "" {
						a.searchBadge(rc, p, badge, p.mint, p.runBG)
					}
				})
				ui.Text(rc, a.sessionPath(s)).Font("Menlo").FontSize(10).
					TextColor(p.ink3).SingleLine()
			})
			if selected {
				ui.Text(rc, "↵").FontSize(10).TextColor(p.ink3).SingleLine()
			}
		})
		if item.RowClicked() {
			a.activateSearchResult(row.Index)
		}
	})
	if a.search.scroll {
		a.search.list.ScrollIntoView(a.search.selected)
		a.search.scroll = false
	}
}

func (a *app) searchBadge(c *ui.Context, p palette, label string, fg, bg ui.Color) {
	ui.Box(c).Height(16).Padding(0, 5, 0, 5).Radius(4).
		Background(bg).Center().Children(func() {
		ui.Text(c, label).FontSize(9).TextColor(fg).SingleLine()
	})
}

func (a *app) searchNoResults(c *ui.Context, p palette) {
	ui.Column(c).Padding(22, 20, 22, 22).Gap(6).Children(func() {
		ui.Text(c, "没有匹配的会话").FontSize(12.5).FontWeight(500).
			TextColor(p.ink2)
		ui.Text(c, "试试会话名称或路径中的其他关键词").FontSize(10.5).
			TextColor(p.ink3)
	})
}

func (a *app) searchFooter(c *ui.Context, p palette, count int) {
	ui.Row(c).Height(34).Padding(0, 14, 0, 14).
		BorderWidth(1, 0, 0, 0).BorderColor(p.border).
		AlignItems(ui.Center).Children(func() {
		ui.Text(c, fmt.Sprintf("%d 个会话", count)).FontSize(9.5).
			TextColor(p.ink3).SingleLine()
		ui.Spacer(c)
		ui.Text(c, "↑↓ 选择").FontSize(9.5).TextColor(p.ink3).SingleLine()
		ui.Box(c).Width(12)
		ui.Text(c, "↵ 打开").FontSize(9.5).TextColor(p.ink3).SingleLine()
		ui.Box(c).Width(12)
		ui.Text(c, "esc 关闭").FontSize(9.5).TextColor(p.ink3).SingleLine()
	})
}

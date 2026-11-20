package main

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

func TestSplitShowsTwoTerminalsAndSwitchesFocus(t *testing.T) {
	first, second := testTerminal(t), testTerminal(t)
	a := &app{
		sessions: []*session{
			{id: 1, name: "终端 1", term: first},
			{id: 2, name: "终端 2", term: second},
		},
		activeID: 1,
	}
	a.startSplit(2)
	if !a.splitting() {
		t.Fatal("split did not start")
	}

	tt := ui.NewTester(a.view, 1380, 780)
	if got := len(a.visiblePaneIDs()); got != 2 {
		t.Fatalf("split pane count = %d, want 2", got)
	}
	for _, want := range []string{"终端 1 终端面板", "终端 2 终端面板"} {
		if _, ok := tt.Access(want); !ok {
			t.Fatalf("missing pane %q in accessibility tree", want)
		}
	}

	if err := tt.Click("切换到 终端 2 面板"); err != nil {
		t.Fatal(err)
	}
	if a.activeID != 2 {
		t.Fatalf("active session = %d, want 2", a.activeID)
	}
	if !tt.Focused("终端 2 终端面板") {
		t.Fatal("clicking a pane header did not focus its terminal")
	}

	a.activeID = 1
	tt.Frame()
	if err := tt.Click("终端 2 终端面板"); err != nil {
		t.Fatal(err)
	}
	if a.activeID != 2 {
		t.Fatalf("active session after clicking terminal = %d, want 2", a.activeID)
	}
}

func TestSplitFromToolbarMenu(t *testing.T) {
	first, second := testTerminal(t), testTerminal(t)
	a := &app{
		sessions: []*session{
			{id: 1, name: "终端 1", term: first},
			{id: 2, name: "终端 2", term: second},
		},
		activeID: 1,
	}
	tt := ui.NewTester(a.view, 1380, 780)

	if err := tt.Click("分屏显示"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"新建终端并分屏", "与 终端 2 分屏"} {
		if !tt.HasText(want) {
			t.Fatalf("missing %q in split menu: %q", want, tt.Texts())
		}
	}
	if err := tt.Click("与 终端 2 分屏"); err != nil {
		t.Fatal(err)
	}
	if !a.splitting() {
		t.Fatal("choosing a session did not start the split")
	}
	if !tt.HasText("已分屏") {
		t.Fatalf("split pane headers are missing: %q", tt.Texts())
	}
}

func TestSplitShortcutTogglesLayout(t *testing.T) {
	first, second := testTerminal(t), testTerminal(t)
	a := &app{
		sessions: []*session{
			{id: 1, name: "终端 1", term: first},
			{id: 2, name: "终端 2", term: second},
		},
		activeID: 1,
	}
	tt := ui.NewTester(a.view, 1380, 780)
	tt.Key(ui.Cmd, ui.KeyD)
	if !a.splitting() {
		t.Fatal("⌘D did not open the split")
	}
	tt.Key(ui.Cmd, ui.KeyD)
	if a.splitting() {
		t.Fatal("⌘D did not close the split")
	}
}

func TestSplitWithNewTerminal(t *testing.T) {
	first, second := testTerminal(t), testTerminal(t)
	a := &app{
		sessions:    []*session{{id: 1, name: "终端 1", term: first}},
		activeID:    1,
		nextSession: 1,
	}
	a.newTerminal = func(terminal.Options) (*terminal.Terminal, error) {
		return second, nil
	}

	a.splitWithNewSession()
	if !a.splitting() {
		t.Fatal("new terminal did not start a split")
	}
	if len(a.sessions) != 2 || a.sessions[1].term != second {
		t.Fatalf("sessions after split = %+v", a.sessions)
	}
	if a.activeID != 2 {
		t.Fatalf("active session = %d, want the new terminal", a.activeID)
	}
	if !slices.Equal(a.paneRows[0], []int{1, 2}) {
		t.Fatalf("split panes = %v, want [[1 2]]", a.paneRows)
	}
}

func TestClosingSplitPaneKeepsOtherSession(t *testing.T) {
	first, second := testTerminal(t), testTerminal(t)
	a := &app{
		sessions: []*session{
			{id: 1, name: "终端 1", term: first},
			{id: 2, name: "终端 2", term: second},
		},
		activeID: 1,
	}
	a.startSplit(2)
	a.closeSession(1)

	if a.splitting() {
		t.Fatal("split stayed active after one pane's session closed")
	}
	if len(a.sessions) != 1 || a.sessions[0].id != 2 {
		t.Fatalf("sessions after close = %+v", a.sessions)
	}
	if a.activeID != 2 {
		t.Fatalf("active session = %d, want 2", a.activeID)
	}
}

func TestSelectingHiddenSessionOpensFullScreenAndKeepsSplit(t *testing.T) {
	first, second, third := testTerminal(t), testTerminal(t), testTerminal(t)
	a := &app{
		sessions: []*session{
			{id: 1, name: "终端 1", term: first},
			{id: 2, name: "终端 2", term: second},
			{id: 3, name: "终端 3", term: third},
		},
		activeID: 1,
	}
	a.startSplit(2)
	a.selectSession(3)

	if a.splitting() {
		t.Fatal("selecting a session outside the split kept the split visible")
	}
	if !slices.Equal(a.paneRows[0], []int{1, 2}) {
		t.Fatalf("saved split panes = %v, want [[1 2]]", a.paneRows)
	}
	if a.activeID != 3 {
		t.Fatalf("active session = %d, want 3", a.activeID)
	}

	tt := ui.NewTester(a.view, 1380, 780)
	if _, ok := tt.Access("终端 1 终端面板"); ok {
		t.Fatal("saved split is still drawn over the full-screen terminal")
	}
	if _, ok := tt.Access("终端 3 终端面板"); !ok {
		t.Fatal("the selected session is not shown full screen")
	}

	a.selectSession(1)
	if !a.splitting() {
		t.Fatal("returning to a split member did not restore the layout")
	}
	if !slices.Equal(a.paneRows[0], []int{1, 2}) {
		t.Fatalf("restored split panes = %v, want [[1 2]]", a.paneRows)
	}
}

func TestNewTerminalFromSplitOpensNewFullScreenSurface(t *testing.T) {
	first, second, third := testTerminal(t), testTerminal(t), testTerminal(t)
	a := &app{
		sessions: []*session{
			{id: 1, name: "终端 1", term: first},
			{id: 2, name: "终端 2", term: second},
		},
		activeID:    1,
		nextSession: 2,
		paneRows:    [][]int{{1, 2}},
	}
	a.newTerminal = func(terminal.Options) (*terminal.Terminal, error) {
		return third, nil
	}

	a.newSession()
	if a.splitting() {
		t.Fatal("a new terminal stayed inside the existing split")
	}
	if !slices.Equal(a.paneRows[0], []int{1, 2}) {
		t.Fatalf("saved split panes = %v, want [[1 2]]", a.paneRows)
	}
	if a.activeID != 3 {
		t.Fatalf("active session = %d, want 3", a.activeID)
	}

	tt := ui.NewTester(a.view, 1380, 780)
	if _, ok := tt.Access("终端 1 终端面板"); ok {
		t.Fatal("the previous split is still drawn behind the new terminal")
	}
	if _, ok := tt.Access("终端 3 终端面板"); !ok {
		t.Fatal("the new terminal is not shown on its own full-screen surface")
	}

	a.selectSession(2)
	if !a.splitting() {
		t.Fatal("returning to a split member did not restore the previous split")
	}
}

func TestClosingSavedSplitMemberUpdatesLayout(t *testing.T) {
	first, second, third := testTerminal(t), testTerminal(t), testTerminal(t)
	a := &app{
		sessions: []*session{
			{id: 1, name: "终端 1", term: first},
			{id: 2, name: "终端 2", term: second},
			{id: 3, name: "终端 3", term: third},
		},
		activeID: 3,
		paneRows: [][]int{{1, 2}},
	}
	a.closeSession(1)

	if len(a.paneRows) != 0 {
		t.Fatalf("saved split = %v, want cleared", a.paneRows)
	}
	if a.activeID != 3 {
		t.Fatalf("active session = %d, want 3", a.activeID)
	}
}

func TestDroppingSavedSplitMemberIntoFullScreenUsesVisiblePane(t *testing.T) {
	first, second, third := testTerminal(t), testTerminal(t), testTerminal(t)
	a := &app{
		sessions: []*session{
			{id: 1, name: "终端 1", term: first},
			{id: 2, name: "终端 2", term: second},
			{id: 3, name: "终端 3", term: third},
		},
		activeID: 3,
		paneRows: [][]int{{1, 2}},
	}

	a.placeDraggedSession(1, 3, dropRight)
	if len(a.paneRows) != 1 || !slices.Equal(a.paneRows[0], []int{3, 1}) {
		t.Fatalf("pane rows = %v, want [[3 1]]", a.paneRows)
	}
}

func TestDraggingSessionBelowRowCreatesFullWidthPane(t *testing.T) {
	first, second, third := testTerminal(t), testTerminal(t), testTerminal(t)
	a := &app{
		sessions: []*session{
			{id: 1, name: "终端 1", term: first},
			{id: 2, name: "终端 2", term: second},
			{id: 3, name: "终端 3", term: third},
		},
		activeID: 1,
		paneRows: [][]int{{1, 2}},
	}
	tt := ui.NewTester(a.view, 1380, 780)

	source, ok := tt.Find("打开会话 终端 3")
	if !ok {
		t.Fatal("session drag source is missing")
	}
	target, ok := tt.Find("终端 1 终端面板")
	if !ok {
		t.Fatal("terminal drop target is missing")
	}
	x, y := target.X+target.W/2, target.Y+target.H-8
	tt.Press(source.X+20, source.Y+source.H/2)
	tt.Move(source.X+36, source.Y+source.H/2+8)
	tt.Move(target.X+30, target.Y+30)
	tt.Move(x, y)
	tt.Frame()
	if !tt.HasText("放置终端到下方") {
		t.Fatalf("missing bottom drop preview: %q", tt.Texts())
	}
	tt.Release(x, y)

	if len(a.paneRows) != 2 ||
		!slices.Equal(a.paneRows[0], []int{1, 2}) ||
		!slices.Equal(a.paneRows[1], []int{3}) {
		t.Fatalf("pane rows = %v, want [[1 2] [3]]", a.paneRows)
	}
	if a.activeID != 3 {
		t.Fatalf("active session = %d, want 3", a.activeID)
	}

	tt.Frame()
	top, _ := tt.Find("终端 1 终端面板")
	bottom, ok := tt.Find("终端 3 终端面板")
	if !ok {
		t.Fatal("dropped terminal is missing")
	}
	if bottom.W < top.W*1.8 {
		t.Fatalf("bottom pane width = %.1f, top pane width = %.1f", bottom.W, top.W)
	}
}

func TestDraggingSessionIntoUnsplitContentCreatesSplit(t *testing.T) {
	first, second := testTerminal(t), testTerminal(t)
	a := &app{
		sessions: []*session{
			{id: 1, name: "终端 1", term: first},
			{id: 2, name: "终端 2", term: second},
		},
		activeID: 1,
	}
	tt := ui.NewTester(a.view, 1380, 780)

	source, ok := tt.Find("打开会话 终端 2")
	if !ok {
		t.Fatal("session drag source is missing")
	}
	target, ok := tt.Find("终端 1 终端面板")
	if !ok {
		t.Fatal("terminal drop target is missing")
	}
	x, y := target.X+target.W/2, target.Y+target.H-8
	tt.Press(source.X+20, source.Y+source.H/2)
	tt.Move(source.X+36, source.Y+source.H/2+8)
	tt.Move(target.X+30, target.Y+30)
	tt.Move(x, y)
	if !tt.HasText("放置终端到下方") {
		t.Fatalf("missing drop preview: %q", tt.Texts())
	}
	tt.Release(x, y)

	if len(a.paneRows) != 2 ||
		!slices.Equal(a.paneRows[0], []int{1}) ||
		!slices.Equal(a.paneRows[1], []int{2}) {
		t.Fatalf("pane rows = %v, want [[1] [2]]", a.paneRows)
	}
}

func TestPlaceDraggedSessionAroundPanes(t *testing.T) {
	first, second, third := testTerminal(t), testTerminal(t), testTerminal(t)
	a := &app{
		sessions: []*session{
			{id: 1, name: "终端 1", term: first},
			{id: 2, name: "终端 2", term: second},
			{id: 3, name: "终端 3", term: third},
		},
		activeID: 1,
		paneRows: [][]int{{1, 2}},
	}

	a.placeDraggedSession(3, 2, dropRight)
	if !slices.Equal(a.paneRows[0], []int{1, 2, 3}) {
		t.Fatalf("right drop panes = %v", a.paneRows)
	}

	a.placeDraggedSession(3, 1, dropLeft)
	if !slices.Equal(a.paneRows[0], []int{3, 1, 2}) {
		t.Fatalf("left drop panes = %v", a.paneRows)
	}

	a.placeDraggedSession(3, 1, dropBottom)
	if len(a.paneRows) != 2 ||
		!slices.Equal(a.paneRows[0], []int{1, 2}) ||
		!slices.Equal(a.paneRows[1], []int{3}) {
		t.Fatalf("bottom drop panes = %v, want [[1 2] [3]]", a.paneRows)
	}
}

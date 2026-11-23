package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSearchMatchesSessionsByNameAndPath(t *testing.T) {
	plain := t.TempDir()
	web := filepath.Join(t.TempDir(), "web-app")
	if err := os.MkdirAll(web, 0o755); err != nil {
		t.Fatal(err)
	}
	a := &app{
		sessions: []*session{
			{id: 1, name: "api server", dir: plain},
			{id: 2, name: "frontend", dir: web},
			{id: 3, name: "api", dir: plain},
		},
		activeID: 1,
	}

	if got, want := a.searchMatches("api"), []int{3, 1}; !slices.Equal(got, want) {
		t.Fatalf("searchMatches(api) = %v, want %v", got, want)
	}
	if got, want := a.searchMatches("WEB-APP"), []int{2}; !slices.Equal(got, want) {
		t.Fatalf("searchMatches(WEB-APP) = %v, want %v", got, want)
	}
	if got, want := a.searchMatches("front web"), []int{2}; !slices.Equal(got, want) {
		t.Fatalf("searchMatches(front web) = %v, want %v", got, want)
	}
	if got := a.searchMatches("   "); len(got) != 0 {
		t.Fatalf("searchMatches(blank) = %v, want no results", got)
	}
}

func TestSearchOverlayFiltersAndOpensSession(t *testing.T) {
	first, second := testTerminal(t), testTerminal(t)
	a := &app{
		sessions: []*session{
			{id: 1, name: "终端 1", term: first, dir: t.TempDir()},
			{id: 2, name: "build server", term: second, dir: t.TempDir()},
		},
		activeID: 1,
	}
	tt := ui.NewTester(a.view, 1380, 780)

	if err := tt.Click("搜索会话或路径"); err != nil {
		t.Fatal(err)
	}
	if !a.search.open {
		t.Fatal("clicking the search field did not open the overlay")
	}
	if !tt.HasText("输入会话名称或路径开始搜索") {
		t.Fatalf("missing search prompt: %q", tt.Texts())
	}

	tt.Type("build")
	if !tt.HasText("1 个会话") {
		t.Fatalf("search did not filter to one session: %q", tt.Texts())
	}
	tt.Key(0, ui.KeyEnter)
	if a.search.open {
		t.Fatal("search overlay stayed open after opening a result")
	}
	if a.activeID != 2 {
		t.Fatalf("active session = %d, want 2", a.activeID)
	}
}

func TestSearchOverlayKeyboardSelectionAndNoResults(t *testing.T) {
	first, second := testTerminal(t), testTerminal(t)
	a := &app{
		sessions: []*session{
			{id: 1, name: "api one", term: first, dir: t.TempDir()},
			{id: 2, name: "api two", term: second, dir: t.TempDir()},
		},
		activeID: 1,
	}
	tt := ui.NewTester(a.view, 1380, 780)

	tt.Key(ui.Cmd, ui.KeyP)
	if !a.search.open {
		t.Fatal("⌘P did not open the search overlay")
	}
	tt.Type("api")
	if !tt.HasText("2 个会话") {
		t.Fatalf("missing both search results: %q", tt.Texts())
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if a.activeID != 2 {
		t.Fatalf("active session = %d, want the second result", a.activeID)
	}

	tt.Key(ui.Cmd, ui.KeyP)
	tt.Type("no-such-session")
	if !tt.HasText("没有匹配的会话") {
		t.Fatalf("missing empty search state: %q", tt.Texts())
	}
	tt.Key(0, ui.KeyEscape)
	if a.search.open {
		t.Fatal("Escape did not close the search overlay")
	}
}

package main

import (
	"errors"
	"image/color"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// The view runs without a window, so the layout can be checked in tests.
func TestView(t *testing.T) {
	a := &app{}
	tt := ui.NewTester(a.view, 1380, 780)
	for _, want := range []string{
		"termix", "新建终端", "还没有打开的会话",
		"还没有打开的终端", "在空白工作区新建终端", "zsh", "等待终端",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if _, ok := tt.Access("在空白工作区新建终端"); !ok {
		t.Errorf("empty workspace action is missing from accessibility tree")
	}
	for _, unwanted := range []string{"atlas-web", "api-server", "运行中", "~/projects/atlas"} {
		if tt.HasText(unwanted) {
			t.Errorf("unexpected demo data %q in %q", unwanted, tt.Texts())
		}
	}
}

func TestThemeToggle(t *testing.T) {
	a := &app{}
	tt := ui.NewTester(a.view, 1380, 780)

	if got, want := tt.Image().RGBAAt(1200, 300), (color.RGBA{0x09, 0x0B, 0x0F, 0xFF}); got != want {
		t.Fatalf("dark terminal background = %v, want %v", got, want)
	}
	if !tt.HasText("深色") {
		t.Fatalf("missing dark theme label in %q", tt.Texts())
	}

	if err := tt.Click("切换到浅色主题"); err != nil {
		t.Fatal(err)
	}
	if !a.light {
		t.Fatal("theme did not switch to light")
	}
	if !tt.HasText("浅色") {
		t.Fatalf("missing light theme label in %q", tt.Texts())
	}
	if got, want := tt.Image().RGBAAt(1200, 300), (color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}); got != want {
		t.Fatalf("light terminal background = %v, want %v", got, want)
	}

	if err := tt.Click("切换到深色主题"); err != nil {
		t.Fatal(err)
	}
	if a.light {
		t.Fatal("theme did not switch back to dark")
	}
}

func TestRenameSession(t *testing.T) {
	a := &app{
		sessions:   []*session{{id: 1, name: "终端 1"}},
		activeID:   1,
		renaming:   true,
		renameID:   1,
		renameText: "  构建  ",
	}
	a.commitRename()
	if a.renaming {
		t.Fatal("dialog stayed open after a rename")
	}
	if got := a.sessions[0].name; got != "构建" {
		t.Fatalf("session name = %q, want %q", got, "构建")
	}
}

func TestRenameRejectsEmptyName(t *testing.T) {
	a := &app{
		sessions:   []*session{{id: 1, name: "终端 1"}},
		activeID:   1,
		renaming:   true,
		renameID:   1,
		renameText: "   ",
	}
	a.commitRename()
	if !a.renaming {
		t.Fatal("dialog closed without a name")
	}
	if got := a.sessions[0].name; got != "终端 1" {
		t.Fatalf("session name = %q, want %q", got, "终端 1")
	}
}

func TestCloseSessionSelectsNeighbor(t *testing.T) {
	a := &app{
		sessions: []*session{
			{id: 1, name: "终端 1"},
			{id: 2, name: "终端 2"},
			{id: 3, name: "终端 3"},
		},
		activeID: 2,
	}
	a.closeSession(2)
	if got := len(a.sessions); got != 2 {
		t.Fatalf("session count = %d, want 2", got)
	}
	if a.activeID != 3 {
		t.Fatalf("active session = %d, want 3", a.activeID)
	}
	a.closeSession(3)
	if a.activeID != 1 {
		t.Fatalf("active session = %d, want 1", a.activeID)
	}
}

func TestClosingLastSessionShowsEmptyWorkspace(t *testing.T) {
	a := &app{
		sessions: []*session{{id: 1, name: "终端 1"}},
		activeID: 1,
	}
	tt := ui.NewTester(a.view, 1380, 780)
	if !tt.HasText("终端 1") {
		t.Fatalf("missing session tab in %q", tt.Texts())
	}

	a.closeSession(1)
	if got := a.terminalSession(); got != nil {
		t.Fatalf("terminal session after closing the last one = %+v", got)
	}
	for range 3 {
		tt.Frame()
		if !tt.HasText("还没有打开的终端") {
			t.Fatalf("empty workspace missing after closing the last session: %q", tt.Texts())
		}
		if tt.HasText("终端 1") {
			t.Fatalf("closed session still shown in %q", tt.Texts())
		}
	}
}

func TestEmptyWorkspaceShowsTerminalError(t *testing.T) {
	a := &app{termErr: errors.New("shell could not start")}
	tt := ui.NewTester(a.view, 1380, 780)
	for _, want := range []string{"终端启动失败", "shell could not start", "在空白工作区新建终端"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestLastTerminalSwapsToEmptyWorkspace(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	term, err := terminal.New(terminal.Options{Conn: server})
	if err != nil {
		t.Skipf("terminal is unavailable: %v", err)
	}
	a := &app{
		sessions: []*session{{id: 1, name: "终端 1", term: term}},
		activeID: 1,
	}
	tt := ui.NewTester(a.view, 1380, 780)
	if tt.HasText("还没有打开的终端") {
		t.Fatalf("empty workspace shown with an open terminal: %q", tt.Texts())
	}

	a.closeSession(1)
	for range 2 {
		tt.Frame()
		if !tt.HasText("还没有打开的终端") {
			t.Fatalf("terminal content stayed after the last close: %q", tt.Texts())
		}
	}
	if a.terminalSession() != nil {
		t.Fatal("closed terminal is still the active content")
	}
}

func TestDetectLangGo(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example\n\ngo 1.27.4\n")
	info, ok := detectLangAt(dir)
	if !ok {
		t.Fatal("go.mod was not detected")
	}
	if info.Name != "Go" || info.Version != "1.27.4" {
		t.Fatalf("detected %+v, want Go 1.27.4", info)
	}
}

func TestDetectLangNodeVersionFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{"name":"demo"}`)
	writeFile(t, dir, ".nvmrc", "v20.11.0\n")
	info, ok := detectLangAt(dir)
	if !ok {
		t.Fatal("package.json was not detected")
	}
	if info.Name != "Node" || info.Version != "20.11.0" {
		t.Fatalf("detected %+v, want Node 20.11.0", info)
	}
}

func TestDetectLangTypeScript(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{"name":"demo"}`)
	writeFile(t, dir, "tsconfig.json", `{}`)
	info, ok := detectLangAt(dir)
	if !ok {
		t.Fatal("package.json was not detected")
	}
	if info.Name != "TypeScript" {
		t.Fatalf("detected %q, want TypeScript", info.Name)
	}
}

func TestDetectLangEmptyDir(t *testing.T) {
	if info, ok := detectLangAt(t.TempDir()); ok {
		t.Fatalf("detected %+v in an empty directory", info)
	}
}

func TestSessionLangWithoutSession(t *testing.T) {
	a := &app{}
	if got, want := a.sessionLang(nil).Name, shellName(); got != want {
		t.Fatalf("language = %q, want %q", got, want)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

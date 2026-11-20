package main

import (
	"testing"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

type fakeVersionManager struct {
	name     string
	versions []string
	err      error
}

func (f fakeVersionManager) Name() string { return f.name }

func (f fakeVersionManager) Installed(string) ([]string, error) {
	return f.versions, f.err
}

func TestResolveInstalledVersion(t *testing.T) {
	versions := []string{"24.18.0", "22.23.4", "18.20.8", "16.20.2"}
	for _, tc := range []struct {
		request string
		want    string
		ok      bool
	}{
		{"16", "16.20.2", true},
		{"v18.20.8", "18.20.8", true},
		{"node@22", "22.23.4", true},
		{"18.20", "18.20.8", true},
		{"20", "", false},
		{"", "", false},
	} {
		got, ok := resolveInstalledVersion(versions, tc.request)
		if got != tc.want || ok != tc.ok {
			t.Errorf("resolveInstalledVersion(%q) = %q, %v; want %q, %v",
				tc.request, got, ok, tc.want, tc.ok)
		}
	}
}

func TestSessionToolchainEnv(t *testing.T) {
	pin := &sessionToolchain{Manager: "mise", Tool: "node", Version: "16.20.2"}
	got := pin.Env()
	if len(got) != 1 || got[0] != "MISE_NODE_VERSION=16.20.2" {
		t.Fatalf("pin environment = %q", got)
	}
}

func TestTerminalOptionsCarryToolchainPin(t *testing.T) {
	a := &app{}
	pin := &sessionToolchain{Manager: "mise", Tool: "node", Version: "16.20.2"}
	opts := a.terminalOptions(pin, "/tmp/project")
	if opts.Dir != "/tmp/project" {
		t.Fatalf("terminal dir = %q", opts.Dir)
	}
	found := false
	for _, env := range opts.Env {
		if env == "MISE_NODE_VERSION=16.20.2" {
			found = true
		}
	}
	if !found {
		t.Fatalf("terminal environment lacks the pinned version: %q", opts.Env)
	}
}

func TestToolForLanguage(t *testing.T) {
	for language, want := range map[string]string{
		"Node":       "node",
		"TypeScript": "node",
		"Python":     "python",
		"C#":         "dotnet",
		"Shell":      "",
	} {
		if got := toolForLanguage(language); got != want {
			t.Errorf("toolForLanguage(%q) = %q, want %q", language, got, want)
		}
	}
}

func TestToolchainPanelAppliesByRestart(t *testing.T) {
	term := testTerminal(t)
	a := &app{
		sessions: []*session{{
			id: 1, name: "终端 1", term: term,
			toolchain: &sessionToolchain{Manager: "mise", Tool: "node", Version: "16.20.2"},
		}},
		activeID:       1,
		versionManager: fakeVersionManager{name: "mise", versions: []string{"24.18.0", "18.20.8", "16.20.2"}},
		sessionSettings: toolchainPanelState{
			open: true, anchor: "gear", tool: "node", draft: "16.20.2",
			versions: []string{"24.18.0", "18.20.8", "16.20.2"},
		},
	}
	tt := ui.NewTester(a.view, 1380, 780)
	for _, want := range []string{"终端设置", "仅此会话", "工具链版本", "16.20.2", "应用并重启"} {
		if !tt.HasText(want) {
			t.Fatalf("missing %q in settings panel: %q", want, tt.Texts())
		}
	}

	if err := tt.Click("工具链版本 18.20.8"); err != nil {
		t.Fatal(err)
	}
	if a.sessionSettings.draft != "18.20.8" {
		t.Fatalf("draft version = %q, want 18.20.8", a.sessionSettings.draft)
	}
	if err := tt.Click("应用并重启终端"); err != nil {
		t.Fatal(err)
	}
	if !a.confirmClose || a.confirmAction != restartTerminal {
		t.Fatal("applying a toolchain did not ask to restart the running terminal")
	}
	if a.restartPin == nil || a.restartPin.Version != "18.20.8" {
		t.Fatalf("restart pin = %+v, want Node 18.20.8", a.restartPin)
	}
	for _, want := range []string{"应用新版本并重启？", "重启后使用 Node 18.20.8"} {
		if !tt.HasText(want) {
			t.Fatalf("missing %q in restart dialog: %q", want, tt.Texts())
		}
	}

	if err := tt.Click("取消"); err != nil {
		t.Fatal(err)
	}
	if a.confirmClose || len(a.sessions) != 1 {
		t.Fatal("cancel changed the running session")
	}
	if a.sessions[0].toolchain == nil || a.sessions[0].toolchain.Version != "16.20.2" {
		t.Fatal("cancel changed the pinned toolchain")
	}
}

func TestSessionSettingsOpensFromGear(t *testing.T) {
	term := testTerminal(t)
	a := &app{
		sessions: []*session{{
			id: 1, name: "终端 1", term: term,
			toolchain: &sessionToolchain{Manager: "mise", Tool: "node", Version: "16.20.2"},
		}},
		activeID:       1,
		versionManager: fakeVersionManager{name: "mise", versions: []string{"16.20.2"}},
		toolchainCache: map[string][]string{"node": {"16.20.2"}},
	}
	tt := ui.NewTester(a.view, 1380, 780)
	if err := tt.Click("当前终端设置"); err != nil {
		t.Fatal(err)
	}
	if !a.sessionSettings.open {
		t.Fatal("settings did not open from the toolbar gear")
	}
	for _, want := range []string{"终端设置", "Node", "16.20.2"} {
		if !tt.HasText(want) {
			t.Fatalf("missing %q in opened settings: %q", want, tt.Texts())
		}
	}
}

func TestToolchainCanReturnToProjectVersion(t *testing.T) {
	term := testTerminal(t)
	a := &app{
		sessions: []*session{{
			id: 1, name: "终端 1", term: term,
			toolchain: &sessionToolchain{Manager: "mise", Tool: "node", Version: "16.20.2"},
		}},
		activeID:       1,
		versionManager: fakeVersionManager{name: "mise", versions: []string{"16.20.2"}},
		sessionSettings: toolchainPanelState{
			open: true, anchor: "gear", tool: "node", draft: "16.20.2",
			versions: []string{"16.20.2"},
		},
	}
	tt := ui.NewTester(a.view, 1380, 780)
	if err := tt.Click("工具链跟随项目配置"); err != nil {
		t.Fatal(err)
	}
	if a.sessionSettings.draft != "" {
		t.Fatalf("follow-project draft = %q", a.sessionSettings.draft)
	}
	if err := tt.Click("应用并重启终端"); err != nil {
		t.Fatal(err)
	}
	if !a.confirmClose || a.confirmAction != restartTerminal || a.restartPin != nil {
		t.Fatal("returning to the project version did not request a restart without a pin")
	}
	if !tt.HasText("重启后将跟随项目配置") {
		t.Fatalf("missing follow-project restart copy: %q", tt.Texts())
	}
}

func TestRestartSessionPreservesIdentityAndPin(t *testing.T) {
	oldTerm := testTerminal(t)
	nextTerm := testTerminal(t)
	a := &app{
		sessions: []*session{{id: 7, name: "构建", term: oldTerm}},
		activeID: 7,
	}
	var got terminal.Options
	a.newTerminal = func(opts terminal.Options) (*terminal.Terminal, error) {
		got = opts
		return nextTerm, nil
	}
	pin := &sessionToolchain{Manager: "mise", Tool: "node", Version: "16.20.2"}
	a.restartSession(7, pin)

	if len(a.sessions) != 1 {
		t.Fatalf("session count = %d", len(a.sessions))
	}
	gotSession := a.sessions[0]
	if gotSession.id != 7 || gotSession.name != "构建" {
		t.Fatalf("restarted session = %+v", gotSession)
	}
	if gotSession.term != nextTerm || gotSession.toolchain != pin {
		t.Fatal("restart did not install the new terminal and pin")
	}
	if a.activeID != 7 {
		t.Fatalf("active session = %d, want 7", a.activeID)
	}
	found := false
	for _, env := range got.Env {
		if env == "MISE_NODE_VERSION=16.20.2" {
			found = true
		}
	}
	if !found {
		t.Fatalf("restarted terminal lacks the pin: %q", got.Env)
	}
}

func TestCustomToolchainVersion(t *testing.T) {
	a := &app{sessionSettings: toolchainPanelState{
		versions:   []string{"24.18.0", "18.20.8", "16.20.2"},
		customText: "18",
	}}
	a.useCustomToolchainVersion()
	if got, want := a.sessionSettings.draft, "18.20.8"; got != want {
		t.Fatalf("custom draft = %q, want %q", got, want)
	}
}

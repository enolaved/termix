package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

func TestStateRoundTripRestoresThemeSessionsAndToolchains(t *testing.T) {
	project := t.TempDir()
	path := filepath.Join(t.TempDir(), "settings.json")
	a := &app{
		light:     true,
		statePath: path,
		sessions: []*session{{
			id:        1,
			name:      "构建",
			dir:       project,
			toolchain: &sessionToolchain{Manager: "mise", Tool: "node", Version: "20.11.0"},
		}},
		activeID: 1,
	}
	a.saveState()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved persistedState
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Version != stateVersion {
		t.Fatalf("state version = %d, want %d", saved.Version, stateVersion)
	}
	if saved.Settings.Theme != themeLight {
		t.Fatalf("saved theme = %q, want %q", saved.Settings.Theme, themeLight)
	}
	if len(saved.Sessions) != 1 {
		t.Fatalf("saved sessions = %+v", saved.Sessions)
	}
	if got := saved.Sessions[0]; got.Name != "构建" || got.PWD != project ||
		got.Toolchain == nil || got.Toolchain.Version != "20.11.0" {
		t.Fatalf("saved session = %+v", got)
	}

	b := &app{statePath: path}
	if !b.readState() {
		t.Fatal("saved state was not loaded")
	}
	if !b.light {
		t.Fatal("saved light theme was not restored")
	}
	var options []terminal.Options
	b.newTerminal = func(opts terminal.Options) (*terminal.Terminal, error) {
		options = append(options, opts)
		return testTerminal(t), nil
	}
	if !b.restoreSessions() {
		t.Fatal("saved sessions were not restored")
	}
	if len(b.sessions) != 1 {
		t.Fatalf("restored sessions = %+v", b.sessions)
	}
	s := b.sessions[0]
	if s.name != "构建" || s.dir != project || b.activeID != s.id {
		t.Fatalf("restored session = %+v, active = %d", s, b.activeID)
	}
	if !sameToolchain(s.toolchain, &sessionToolchain{
		Manager: "mise", Tool: "node", Version: "20.11.0",
	}) {
		t.Fatalf("restored toolchain = %+v", s.toolchain)
	}
	if len(options) != 1 || options[0].Dir != project {
		t.Fatalf("terminal options = %+v", options)
	}
	if !slices.Contains(options[0].Env, "MISE_NODE_VERSION=20.11.0") {
		t.Fatalf("terminal env lacks restored toolchain: %v", options[0].Env)
	}
}

func TestRestoreSkipsMissingDirectoryAndUnknownManager(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	state := persistedState{
		Version:  stateVersion,
		Settings: persistedSettings{Theme: themeDark},
		Sessions: []persistedSession{{
			Name:      "旧项目",
			PWD:       filepath.Join(t.TempDir(), "gone"),
			Toolchain: &sessionToolchain{Manager: "asdf", Tool: "node", Version: "18.0.0"},
		}},
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	var options []terminal.Options
	a := &app{statePath: path}
	a.newTerminal = func(opts terminal.Options) (*terminal.Terminal, error) {
		options = append(options, opts)
		return testTerminal(t), nil
	}
	if !a.readState() || !a.restoreSessions() {
		t.Fatal("saved state was not loaded")
	}
	if len(a.sessions) != 1 || a.sessions[0].dir != "" || a.sessions[0].toolchain != nil {
		t.Fatalf("restored session = %+v", a.sessions)
	}
	if len(options) != 1 || options[0].Dir != "" {
		t.Fatalf("terminal options = %+v", options)
	}
	for _, env := range options[0].Env {
		if env == "MISE_NODE_VERSION=18.0.0" {
			t.Fatalf("unsupported manager was applied: %v", options[0].Env)
		}
	}
}

func TestCorruptStateStartsWithoutRestore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	a := &app{statePath: path, light: true}
	if a.readState() {
		t.Fatal("corrupt state was accepted")
	}
	if !a.light {
		t.Fatal("corrupt state changed the current theme")
	}
	if a.restoreSessions() {
		t.Fatal("corrupt state offered sessions to restore")
	}
}

func TestRestoreKeepsSavedActiveSessionWhenAnotherFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	state := persistedState{
		Version:       stateVersion,
		Settings:      persistedSettings{Theme: themeDark},
		ActiveSession: 1,
		Sessions: []persistedSession{
			{Name: "失败"},
			{Name: "可用"},
		},
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	attempts := 0
	a := &app{statePath: path}
	a.newTerminal = func(terminal.Options) (*terminal.Terminal, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("shell failed")
		}
		return testTerminal(t), nil
	}
	if !a.readState() || !a.restoreSessions() {
		t.Fatal("saved state was not loaded")
	}
	if len(a.sessions) != 1 || a.sessions[0].name != "可用" ||
		a.activeID != a.sessions[0].id {
		t.Fatalf("restored sessions = %+v, active = %d", a.sessions, a.activeID)
	}
}

func TestThemeTogglePersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	a := &app{statePath: path}
	tt := ui.NewTester(a.view, 1380, 780)
	if err := tt.Click("切换到浅色主题"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved persistedState
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Settings.Theme != themeLight {
		t.Fatalf("theme in saved state = %q, want %q", saved.Settings.Theme, themeLight)
	}
}

func TestClosingSessionUpdatesSavedState(t *testing.T) {
	project := t.TempDir()
	path := filepath.Join(t.TempDir(), "settings.json")
	a := &app{
		statePath: path,
		sessions: []*session{
			{id: 1, name: "保留", dir: project},
			{id: 2, name: "关闭", dir: project},
		},
		activeID: 1,
	}
	a.closeSession(2)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved persistedState
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if len(saved.Sessions) != 1 || saved.Sessions[0].Name != "保留" {
		t.Fatalf("saved sessions = %+v", saved.Sessions)
	}
}

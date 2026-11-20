package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo"
)

// sessionToolchain pins one tool version to one terminal session. The
// requested version is resolved to an installed version before the pin is
// stored, so the environment uses an exact version.
type sessionToolchain struct {
	Manager string
	Tool    string
	Version string
}

// Env returns the variables that make a version manager select this tool
// version for the shell and every process it starts.
func (t *sessionToolchain) Env() []string {
	if t == nil {
		return nil
	}
	switch t.Manager {
	case "mise":
		return []string{"MISE_" + toolchainEnvKey(t.Tool) + "_VERSION=" + t.Version}
	}
	return nil
}

// versionManager lists installed versions. The interface leaves room for
// asdf, pyenv, rbenv and shell-based managers later.
type versionManager interface {
	Name() string
	Installed(tool string) ([]string, error)
}

type miseManager struct {
	path string
}

func (m *miseManager) Name() string { return "mise" }

func (m *miseManager) Installed(tool string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, m.path, "ls", tool, "--installed", "--json").Output()
	if err != nil {
		return nil, fmt.Errorf("读取 mise 版本失败: %w", err)
	}
	if strings.TrimSpace(string(out)) == "" {
		return nil, nil
	}
	var rows []struct {
		Version   string `json:"version"`
		Installed bool   `json:"installed"`
	}
	if err := json.Unmarshal(out, &rows); err != nil {
		return nil, fmt.Errorf("解析 mise 版本失败: %w", err)
	}
	versions := make([]string, 0, len(rows))
	seen := make(map[string]bool)
	for _, row := range rows {
		if row.Installed && row.Version != "" && !seen[row.Version] {
			seen[row.Version] = true
			versions = append(versions, row.Version)
		}
	}
	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i], versions[j]) > 0
	})
	return versions, nil
}

// toolchainPanelState is the per-window state of the session settings
// popover. It is main-thread state like the rest of the view model.
type toolchainPanelState struct {
	open       bool
	anchor     string
	tool       string
	versions   []string
	loading    bool
	err        string
	draft      string
	customText string
}

// ensureVersionManager finds the manager available on this machine. Only
// mise is supported in the first version; the interface keeps the UI and
// session model independent of it.
func (a *app) ensureVersionManager() versionManager {
	if a.versionManager != nil {
		return a.versionManager
	}
	if path := lookTool("mise"); path != "" {
		a.versionManager = &miseManager{path: path}
	}
	return a.versionManager
}

func (a *app) openSessionSettings(s *session, anchor string) {
	if s == nil {
		return
	}
	if a.sessionSettings.open && a.sessionSettings.anchor == anchor {
		a.sessionSettings.open = false
		return
	}
	panel := toolchainPanelState{open: true, anchor: anchor}
	if s.toolchain != nil {
		panel.tool = s.toolchain.Tool
		panel.draft = s.toolchain.Version
	} else {
		panel.tool = toolForLanguage(a.sessionLang(s).Name)
	}
	a.sessionSettings = panel
	a.loadToolchainVersions(panel.tool)
	a.invalidate()
}

func (a *app) loadToolchainVersions(tool string) {
	if tool == "" {
		a.sessionSettings.err = "当前目录未识别到可管理的工具链"
		return
	}
	if versions, ok := a.toolchainCache[tool]; ok {
		a.sessionSettings.versions = versions
	}
	manager := a.ensureVersionManager()
	if manager == nil {
		a.sessionSettings.err = "未检测到支持的版本管理器（mise）"
		return
	}
	a.sessionSettings.loading = len(a.sessionSettings.versions) == 0
	a.sessionSettings.err = ""
	go func() {
		versions, err := manager.Installed(tool)
		mygo.RunOnMain(func() {
			if err == nil {
				if a.toolchainCache == nil {
					a.toolchainCache = make(map[string][]string)
				}
				a.toolchainCache[tool] = versions
			}
			if a.sessionSettings.open && a.sessionSettings.tool == tool {
				a.sessionSettings.loading = false
				if err != nil {
					if len(a.sessionSettings.versions) == 0 {
						a.sessionSettings.err = err.Error()
					}
				} else {
					a.sessionSettings.versions = versions
					if len(versions) == 0 {
						a.sessionSettings.err = "没有找到已安装版本"
					}
				}
			}
			a.invalidate()
		})
	}()
}

func (a *app) useCustomToolchainVersion() {
	if strings.TrimSpace(a.sessionSettings.customText) == "" {
		a.sessionSettings.err = "请输入版本号"
		a.invalidate()
		return
	}
	version, ok := resolveInstalledVersion(a.sessionSettings.versions, a.sessionSettings.customText)
	if !ok {
		a.sessionSettings.err = fmt.Sprintf("本机未安装版本 %q", strings.TrimSpace(a.sessionSettings.customText))
		a.invalidate()
		return
	}
	a.sessionSettings.draft = version
	a.sessionSettings.customText = ""
	a.sessionSettings.err = ""
	a.invalidate()
}

func (a *app) applySessionToolchain(s *session) {
	if s == nil {
		return
	}
	draft := a.sessionSettings.draft
	if s.toolchain != nil && draft != "" &&
		s.toolchain.Tool == a.sessionSettings.tool && s.toolchain.Version == draft {
		a.sessionSettings.open = false
		a.invalidate()
		return
	}
	var pin *sessionToolchain
	if draft != "" {
		manager := a.ensureVersionManager()
		if manager == nil {
			a.sessionSettings.err = "未检测到支持的版本管理器（mise）"
			a.invalidate()
			return
		}
		pin = &sessionToolchain{
			Manager: manager.Name(),
			Tool:    a.sessionSettings.tool,
			Version: draft,
		}
	}
	if sameToolchain(s.toolchain, pin) {
		a.sessionSettings.open = false
		a.invalidate()
		return
	}
	a.sessionSettings.open = false
	a.requestRestartSession(s.id, pin)
}

func sameToolchain(a, b *sessionToolchain) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Manager == b.Manager && a.Tool == b.Tool && a.Version == b.Version
}

func (a *app) restartSession(id int, pin *sessionToolchain) {
	old := a.sessionByID(id)
	if old == nil {
		return
	}
	index := a.sessionIndex(id)
	dir := ""
	if old.term != nil {
		dir = old.term.Dir()
	}
	term, err := a.startTerminal(pin, dir)
	if err != nil {
		a.sessionSettings.open = true
		a.sessionSettings.anchor = "gear"
		a.sessionSettings.err = "重启失败: " + err.Error()
		a.invalidate()
		return
	}
	next := &session{id: old.id, name: old.name, term: term, toolchain: pin}
	a.sessions[index] = next
	a.activeID = next.id
	a.focusPaneID = next.id
	if old.term != nil {
		_ = old.term.Close()
	}
	a.watchDir = ""
	a.invalidate()
}

// toolForLanguage maps the language shown for a directory to the tool a
// version manager installs.
func toolForLanguage(name string) string {
	switch name {
	case "Node", "TypeScript":
		return "node"
	case "Deno":
		return "deno"
	case "Bun":
		return "bun"
	case "Go":
		return "go"
	case "Rust":
		return "rust"
	case "Python":
		return "python"
	case "Ruby":
		return "ruby"
	case "PHP":
		return "php"
	case "Java":
		return "java"
	case "Swift":
		return "swift"
	case "Zig":
		return "zig"
	case "C#":
		return "dotnet"
	case "Dart":
		return "dart"
	case "Elixir":
		return "elixir"
	case "Scala":
		return "scala"
	}
	return ""
}

var toolchainLabels = map[string]string{
	"node": "Node", "deno": "Deno", "bun": "Bun", "go": "Go",
	"rust": "Rust", "python": "Python", "ruby": "Ruby", "php": "PHP",
	"java": "Java", "swift": "Swift", "zig": "Zig", "dotnet": ".NET",
	"dart": "Dart", "elixir": "Elixir", "scala": "Scala",
}

func toolchainLabel(tool string) string {
	if label, ok := toolchainLabels[tool]; ok {
		return label
	}
	return tool
}

func sessionToolchainBadge(s *session) string {
	if s == nil || s.toolchain == nil {
		return ""
	}
	return toolchainLabel(s.toolchain.Tool) + " " + s.toolchain.Version
}

func toolchainEnvKey(tool string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(tool) {
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}

func resolveInstalledVersion(versions []string, request string) (string, bool) {
	request = strings.TrimSpace(request)
	if at := strings.LastIndex(request, "@"); at >= 0 {
		request = request[at+1:]
	}
	request = strings.TrimPrefix(request, "v")
	if request == "" {
		return "", false
	}
	for _, version := range versions {
		if version == request {
			return version, true
		}
	}
	best := ""
	for _, version := range versions {
		if strings.HasPrefix(version, request+".") &&
			(best == "" || compareVersions(version, best) > 0) {
			best = version
		}
	}
	return best, best != ""
}

func compareVersions(a, b string) int {
	pa := strings.FieldsFunc(a, func(r rune) bool { return r == '.' || r == '-' || r == '+' })
	pb := strings.FieldsFunc(b, func(r rune) bool { return r == '.' || r == '-' || r == '+' })
	for i := 0; i < len(pa) && i < len(pb); i++ {
		na, ea := strconv.Atoi(pa[i])
		nb, eb := strconv.Atoi(pb[i])
		switch {
		case ea == nil && eb == nil && na != nb:
			if na < nb {
				return -1
			}
			return 1
		case ea == nil && eb != nil:
			return 1
		case ea != nil && eb == nil:
			return -1
		case pa[i] != pb[i]:
			if pa[i] < pb[i] {
				return -1
			}
			return 1
		}
	}
	switch {
	case len(pa) < len(pb):
		return -1
	case len(pa) > len(pb):
		return 1
	}
	return 0
}

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// langInfo is the language a session's working directory belongs to, with
// the tool version that runs it when known.
type langInfo struct {
	Name    string
	Version string
	Tool    string
}

// sessionLang detects the language of the directory a session is in. Results
// are cached per directory and versions per tool: detection reads project
// files only, while a missing version is resolved in the background by asking
// the tool itself.
func (a *app) sessionLang(s *session) langInfo {
	if s == nil || s.term == nil {
		return langInfo{Name: shellName()}
	}
	dir := s.term.Dir()
	if dir == "" {
		dir, _ = os.UserHomeDir()
	}
	if dir == "" {
		return langInfo{Name: shellName()}
	}
	if a.langs == nil {
		a.langs = make(map[string]langInfo)
	}
	if a.toolVersions == nil {
		a.toolVersions = make(map[string]string)
	}
	if a.toolPending == nil {
		a.toolPending = make(map[string]bool)
	}

	if info, ok := a.langs[dir]; ok {
		return info
	}
	info := detectLang(dir)
	if info.Version == "" && info.Tool != "" {
		if v, ok := a.toolVersions[info.Tool]; ok {
			info.Version = v
		}
	}
	a.langs[dir] = info
	if info.Version == "" && info.Tool != "" {
		a.probeTool(info.Tool)
	}
	return info
}

// probeTool resolves a tool's version off the UI thread and fills in every
// directory that was waiting for it.
func (a *app) probeTool(tool string) {
	if a.toolPending[tool] {
		return
	}
	a.toolPending[tool] = true
	go func() {
		version := probeVersion(tool)
		mygo.RunOnMain(func() {
			a.toolPending[tool] = false
			a.toolVersions[tool] = version
			for dir, info := range a.langs {
				if info.Tool == tool && info.Version == "" {
					info.Version = version
					a.langs[dir] = info
				}
			}
			a.invalidate()
		})
	}()
}

// detectLang walks up from dir until a project marker is found, so a shell
// inside a package still shows the language of its project.
func detectLang(dir string) langInfo {
	home, _ := os.UserHomeDir()
	for {
		if info, ok := detectLangAt(dir); ok {
			return info
		}
		if dir == home {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return langInfo{Name: shellName(), Tool: shellTool()}
}

func detectLangAt(dir string) (langInfo, bool) {
	has := func(names ...string) bool {
		for _, name := range names {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				return true
			}
		}
		return false
	}
	hasGlob := func(pattern string) bool {
		matches, _ := filepath.Glob(filepath.Join(dir, pattern))
		return len(matches) > 0
	}
	versionIn := func(names ...string) string {
		for _, name := range names {
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				continue
			}
			line := strings.TrimSpace(strings.SplitN(string(data), "\n", 2)[0])
			line = strings.TrimPrefix(strings.TrimPrefix(line, "v"), "python-")
			if versionNumber.MatchString(line) {
				return line
			}
		}
		return ""
	}

	switch {
	case has("go.mod"):
		return langInfo{Name: "Go", Version: goModVersion(dir), Tool: "go"}, true
	case has("deno.json", "deno.jsonc"):
		return langInfo{Name: "Deno", Tool: "deno"}, true
	case has("bun.lockb", "bun.lock"):
		return langInfo{Name: "Bun", Tool: "bun"}, true
	case has("package.json"):
		if has("tsconfig.json") {
			return langInfo{Name: "TypeScript", Tool: "tsc"}, true
		}
		version := versionIn(".nvmrc", ".node-version")
		if version == "" {
			version = packageNodeEngine(dir)
		}
		return langInfo{Name: "Node", Version: version, Tool: "node"}, true
	case has("Cargo.toml"):
		return langInfo{Name: "Rust", Tool: "rustc"}, true
	case has("pyproject.toml", "requirements.txt", "setup.py", "Pipfile", ".python-version"):
		return langInfo{Name: "Python", Version: versionIn(".python-version"), Tool: "python3"}, true
	case has("Gemfile", ".ruby-version"):
		return langInfo{Name: "Ruby", Version: versionIn(".ruby-version"), Tool: "ruby"}, true
	case has("composer.json"):
		return langInfo{Name: "PHP", Tool: "php"}, true
	case has("pom.xml", "build.gradle", "build.gradle.kts"):
		return langInfo{Name: "Java", Tool: "java"}, true
	case has("Package.swift"):
		return langInfo{Name: "Swift", Tool: "swift"}, true
	case has("build.zig"):
		return langInfo{Name: "Zig", Tool: "zig"}, true
	case has("pubspec.yaml"):
		return langInfo{Name: "Dart", Tool: "dart"}, true
	case has("mix.exs"):
		return langInfo{Name: "Elixir", Tool: "elixir"}, true
	case has("build.sbt"):
		return langInfo{Name: "Scala", Tool: "scala"}, true
	case has("global.json") || hasGlob("*.csproj") || hasGlob("*.fsproj") || hasGlob("*.sln"):
		return langInfo{Name: "C#", Tool: "dotnet"}, true
	}
	return langInfo{}, false
}

var goVersionLine = regexp.MustCompile(`(?m)^go\s+(\S+)`)

func goModVersion(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return ""
	}
	if m := goVersionLine.FindSubmatch(data); m != nil {
		return string(m[1])
	}
	return ""
}

func packageNodeEngine(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return ""
	}
	var pkg struct {
		Engines struct {
			Node string `json:"node"`
		} `json:"engines"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return ""
	}
	return strings.TrimSpace(pkg.Engines.Node)
}

func shellTool() string {
	switch sh := shellName(); sh {
	case "zsh", "bash", "fish":
		return sh
	}
	return ""
}

var versionNumber = regexp.MustCompile(`^\d+(\.\d+)*$`)
var firstVersion = regexp.MustCompile(`\d+\.\d+(\.\d+)?`)
var elixirVersion = regexp.MustCompile(`Elixir\s+(\d+\.\d+(?:\.\d+)?)`)

var versionArgs = map[string][]string{
	"go":      {"version"},
	"node":    {"--version"},
	"tsc":     {"--version"},
	"python3": {"--version"},
	"rustc":   {"--version"},
	"ruby":    {"--version"},
	"php":     {"--version"},
	"java":    {"-version"},
	"swift":   {"--version"},
	"deno":    {"--version"},
	"bun":     {"--version"},
	"zig":     {"version"},
	"dotnet":  {"--version"},
	"dart":    {"--version"},
	"elixir":  {"--version"},
	"scala":   {"-version"},
	"zsh":     {"--version"},
	"bash":    {"--version"},
	"fish":    {"--version"},
}

// probeVersion asks a tool for its version, with a short timeout so a broken
// toolchain cannot stall the background probe.
func probeVersion(tool string) string {
	args, ok := versionArgs[tool]
	if !ok {
		return ""
	}
	path := lookTool(tool)
	if path == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer cancel()
	out, _ := exec.CommandContext(ctx, path, args...).CombinedOutput()
	text := string(out)
	if tool == "elixir" {
		if m := elixirVersion.FindStringSubmatch(text); m != nil {
			return m[1]
		}
		return ""
	}
	return firstVersion.FindString(text)
}

// lookTool finds a tool even when the app was started outside a login shell
// and PATH lacks Homebrew or per-user installs.
func lookTool(name string) string {
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	dirs := []string{"/opt/homebrew/bin", "/usr/local/bin", "/usr/bin", "/bin"}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		dirs = append(dirs,
			filepath.Join(home, ".local/bin"),
			filepath.Join(home, ".cargo/bin"),
			filepath.Join(home, ".bun/bin"),
			filepath.Join(home, ".deno/bin"),
			filepath.Join(home, "go/bin"),
			filepath.Join(home, ".pyenv/shims"),
			filepath.Join(home, ".local/share/mise/shims"),
			filepath.Join(home, ".asdf/shims"),
		)
	}
	for _, dir := range dirs {
		path := filepath.Join(dir, name)
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			return path
		}
	}
	return ""
}

// startDirWatch notices when a shell changes directory, which is when its
// language chip and saved working directory may need to change. It wakes on
// a real change only.
func (a *app) startDirWatch() {
	if a.watchStop != nil {
		return
	}
	stop := make(chan struct{})
	a.watchStop = stop
	go func() {
		ticker := time.NewTicker(1500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				mygo.RunOnMain(a.checkSessionDirs)
			}
		}
	}()
}

func (a *app) stopDirWatch() {
	if a.watchStop == nil {
		return
	}
	close(a.watchStop)
	a.watchStop = nil
}

func (a *app) checkSessionDirs() {
	changed := false
	for _, s := range a.sessions {
		if s.term == nil {
			continue
		}
		dir := s.term.Dir()
		if dir == "" || dir == s.dir {
			continue
		}
		s.dir = dir
		changed = true
	}
	if !changed {
		return
	}
	a.saveState()
	a.invalidate()
}

// langTone returns the text and chip background colors for a language, with
// its own hue per theme so chips stay readable on dark and light surfaces.
func langTone(p palette, light bool, name string) (ui.Color, ui.Color) {
	type tone struct{ dark, light ui.Color }
	tones := map[string]tone{
		"Go":         {ui.Hex("#67D4F0"), ui.Hex("#0B6E86")},
		"Node":       {ui.Hex("#9BD97A"), ui.Hex("#3C7A2C")},
		"TypeScript": {ui.Hex("#7FB3F7"), ui.Hex("#235EA8")},
		"Deno":       {ui.Hex("#7DE3C3"), ui.Hex("#0A7A66")},
		"Bun":        {ui.Hex("#F2C48D"), ui.Hex("#9A5E14")},
		"Python":     {ui.Hex("#7FB3F0"), ui.Hex("#2B5FA8")},
		"Rust":       {ui.Hex("#E9A277"), ui.Hex("#A4521F")},
		"Ruby":       {ui.Hex("#EE8A90"), ui.Hex("#A62A38")},
		"PHP":        {ui.Hex("#B3A5F2"), ui.Hex("#5A45B5")},
		"Java":       {ui.Hex("#E7B06A"), ui.Hex("#8F5A12")},
		"Swift":      {ui.Hex("#F49B72"), ui.Hex("#B44B1F")},
		"Zig":        {ui.Hex("#EACB72"), ui.Hex("#7E6410")},
		"C#":         {ui.Hex("#B99CF2"), ui.Hex("#6042A8")},
		"Dart":       {ui.Hex("#6FCBE0"), ui.Hex("#0D6D80")},
		"Elixir":     {ui.Hex("#C39BF0"), ui.Hex("#6A3FA8")},
		"Scala":      {ui.Hex("#EC8080"), ui.Hex("#A82F2F")},
	}
	t, ok := tones[name]
	if !ok {
		return p.ink2, p.subtle
	}
	fg := t.dark
	if light {
		fg = t.light
	}
	alpha := float32(0.16)
	if light {
		alpha = 0.12
	}
	return fg, ui.RGBA(fg.R, fg.G, fg.B, alpha)
}

// langPill is the language chip under the tab bar: the language in its own
// hue and the version beside it. It opens the session settings for quick
// version changes.
func (a *app) langPill(c *ui.Context, p palette, s *session, info langInfo) ui.Element {
	name, version := info.Name, info.Version
	pinned := s != nil && s.toolchain != nil
	if pinned {
		name = toolchainLabel(s.toolchain.Tool)
		version = s.toolchain.Version
	}
	toneName := name
	if pinned && s.toolchain.Tool == "dotnet" {
		toneName = "C#"
	}
	fg, bg := langTone(p, a.light, toneName)
	label := "切换终端工具链版本"
	if pinned {
		label = "切换终端工具链版本，当前固定 " + name + " " + version
	}
	btn := ui.ButtonBase(c.Key("toolchain-pill")).Height(22).Padding(0, 8, 0, 8).
		Radius(4).Background(bg).Gap(5).AlignItems(ui.Center).
		Cursor(ui.CursorPointer).Label(label).
		Tooltip("工具链版本 · 当前会话")
	if btn.Hovered() {
		btn.Border(1, ui.RGBA(fg.R, fg.G, fg.B, 0.35))
	}
	btn.Children(func() {
		if pinned {
			icon(c, svgPin, 10, fg)
		}
		ui.Text(c, name).FontSize(10.5).FontWeight(500).
			TextColor(fg).SingleLine()
		if version != "" {
			ui.Text(c, version).FontSize(10).
				TextColor(ui.RGBA(fg.R, fg.G, fg.B, 0.7)).SingleLine()
		}
	})
	btn.OnClick(func() { a.openSessionSettings(s, "pill") })
	return btn
}

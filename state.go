package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
)

const (
	stateVersion  = 1
	stateFileName = "settings.json"

	themeDark  = "dark"
	themeLight = "light"
)

// persistedState is the app state kept between launches. Version lets a
// later release add settings without guessing about this layout.
type persistedState struct {
	Version       int                `json:"version"`
	Settings      persistedSettings  `json:"settings"`
	ActiveSession int                `json:"activeSession,omitempty"`
	Sessions      []persistedSession `json:"sessions"`
}

// persistedSettings holds app-wide settings. Only the current theme is
// stored for now; new settings belong here.
type persistedSettings struct {
	Theme string `json:"theme"`
}

// persistedSession keeps what is needed to reopen a terminal: its name,
// working directory and optional toolchain pin. Terminal output and other
// runtime-only state are deliberately not stored.
type persistedSession struct {
	Name      string            `json:"name,omitempty"`
	PWD       string            `json:"pwd,omitempty"`
	Toolchain *sessionToolchain `json:"toolchain,omitempty"`
}

// readState loads app state before the first window is built, so the saved
// theme applies without a flash of the default theme. It returns false when
// there is no usable state and the caller should open a fresh session.
func (a *app) readState() bool {
	path, err := a.resolveStatePath()
	if err != nil {
		return false
	}
	a.statePath = path
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	if err != nil {
		log.Printf("termix: read settings: %v", err)
		return false
	}
	var state persistedState
	if err := json.Unmarshal(data, &state); err != nil {
		log.Printf("termix: parse settings: %v", err)
		return false
	}
	if state.Version != stateVersion {
		return false
	}
	a.light = state.Settings.Theme == themeLight
	a.restored = &state
	return true
}

// restoreSessions reopens the terminals from the saved state. It reports
// false when the app should create a default session instead.
func (a *app) restoreSessions() bool {
	if a.restored == nil {
		return false
	}
	state := *a.restored
	a.restored = nil
	a.restoringState = true
	a.nextSession = 0
	ids := make([]int, len(state.Sessions))
	for i, saved := range state.Sessions {
		if s := a.addSessionWith(saved.Name, restoreToolchain(saved.Toolchain), restoreWorkingDir(saved.PWD)); s != nil {
			ids[i] = s.id
		}
	}
	if len(a.sessions) > 0 {
		index := state.ActiveSession
		switch {
		case index >= 0 && index < len(ids) && ids[index] != 0:
			a.activeID = ids[index]
		default:
			a.activeID = a.sessions[0].id
		}
		a.focusPaneID = a.activeID
	}
	a.restoringState = false
	return true
}

// saveState writes the current settings and session summary. It is called
// on the main thread whenever a value that belongs in the state changes.
func (a *app) saveState() {
	if a.restoringState || a.statePath == "" {
		return
	}
	data, err := json.MarshalIndent(a.snapshotState(), "", "  ")
	if err != nil {
		log.Printf("termix: encode settings: %v", err)
		return
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(a.statePath), 0o755); err != nil {
		log.Printf("termix: create settings directory: %v", err)
		return
	}
	tmp := a.statePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		log.Printf("termix: write settings: %v", err)
		return
	}
	if err := os.Rename(tmp, a.statePath); err != nil {
		_ = os.Remove(tmp)
		log.Printf("termix: replace settings: %v", err)
	}
}

func (a *app) snapshotState() persistedState {
	settings := persistedSettings{Theme: themeDark}
	if a.light {
		settings.Theme = themeLight
	}
	state := persistedState{
		Version:  stateVersion,
		Settings: settings,
		Sessions: make([]persistedSession, 0, len(a.sessions)),
	}
	for i, s := range a.sessions {
		if s.id == a.activeID {
			state.ActiveSession = i
		}
		state.Sessions = append(state.Sessions, persistedSession{
			Name:      s.name,
			PWD:       a.sessionDir(s),
			Toolchain: cloneToolchain(s.toolchain),
		})
	}
	return state
}

// sessionDir returns the directory a session should reopen in, preferring
// the shell's latest report over the directory it started in.
func (a *app) sessionDir(s *session) string {
	if s == nil {
		return ""
	}
	if s.term != nil {
		if dir := s.term.Dir(); dir != "" {
			s.dir = dir
		}
	}
	return s.dir
}

func (a *app) resolveStatePath() (string, error) {
	if a.statePath != "" {
		return a.statePath, nil
	}
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, stateFileName), nil
}

// restoreWorkingDir accepts an absolute directory that still exists. A
// missing directory starts the shell at home rather than failing the restore.
func restoreWorkingDir(pwd string) string {
	pwd = strings.TrimSpace(pwd)
	if pwd == "" {
		return ""
	}
	if pwd == "~" || strings.HasPrefix(pwd, "~/") {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			return ""
		}
		pwd = filepath.Join(home, strings.TrimPrefix(pwd, "~/"))
	}
	if !filepath.IsAbs(pwd) {
		return ""
	}
	info, err := os.Stat(pwd)
	if err != nil || !info.IsDir() {
		return ""
	}
	return pwd
}

func restoreToolchain(pin *sessionToolchain) *sessionToolchain {
	if pin == nil || pin.Tool == "" || pin.Version == "" {
		return nil
	}
	if pin.Manager != "mise" {
		return nil
	}
	return cloneToolchain(pin)
}

func cloneToolchain(pin *sessionToolchain) *sessionToolchain {
	if pin == nil {
		return nil
	}
	next := *pin
	return &next
}

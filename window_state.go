package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
)

const (
	defaultWindowWidth  = 1440
	defaultWindowHeight = 900
	minWindowWidth      = 1100
	minWindowHeight     = 680
	windowScreenMargin  = 80
)

// initialWindowSize picks a roomy default that still fits the primary
// display's work area, leaving space around the window instead of filling
// the screen.
func initialWindowSize() (int, int) {
	area := mygo.Screen.PrimaryDisplay().WorkArea
	return windowSizeFor(area.Width, area.Height)
}

func windowSizeFor(workWidth, workHeight int) (int, int) {
	width, height := defaultWindowWidth, defaultWindowHeight
	if workWidth > 0 {
		w := workWidth - windowScreenMargin
		if w < 900 {
			w = workWidth - 24
		}
		width = min(width, w)
	}
	if workHeight > 0 {
		h := workHeight - windowScreenMargin
		if h < 600 {
			h = workHeight - 24
		}
		height = min(height, h)
	}
	return max(width, 320), max(height, 240)
}

// windowStateStore keeps MyGo's window bounds in the app's data directory,
// but forgets a saved size that is too small to be useful. Without this,
// a window resized to the old minimum would keep reopening tiny.
type windowStateStore struct {
	path          string
	defaultWidth  int
	defaultHeight int
}

func (s *windowStateStore) LoadWindowStates() (map[string]mygo.WindowState, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var states map[string]mygo.WindowState
	if err := json.Unmarshal(data, &states); err != nil {
		return nil, err
	}
	minWidth := min(1160, s.defaultWidth)
	minHeight := min(700, s.defaultHeight)
	for key, state := range states {
		if state.Maximized || state.FullScreen {
			continue
		}
		if state.Width < minWidth || state.Height < minHeight {
			delete(states, key)
		}
	}
	return states, nil
}

func (s *windowStateStore) SaveWindowStates(states map[string]mygo.WindowState) error {
	data, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (a *app) useWindowStateStore(width, height int) {
	if a.statePath == "" {
		return
	}
	mygo.App.SetWindowStateStore(&windowStateStore{
		path:          filepath.Join(filepath.Dir(a.statePath), "window-state.json"),
		defaultWidth:  width,
		defaultHeight: height,
	})
}

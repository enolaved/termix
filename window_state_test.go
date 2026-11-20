package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo"
)

func TestWindowSizeFor(t *testing.T) {
	for _, tc := range []struct {
		workWidth, workHeight int
		width, height         int
	}{
		{0, 0, 1440, 900},
		{1440, 875, 1360, 795},
		{1920, 1055, 1440, 900},
		{1280, 775, 1200, 695},
		{800, 600, 776, 576},
	} {
		width, height := windowSizeFor(tc.workWidth, tc.workHeight)
		if width != tc.width || height != tc.height {
			t.Errorf("windowSizeFor(%d, %d) = %dx%d, want %dx%d",
				tc.workWidth, tc.workHeight, width, height, tc.width, tc.height)
		}
	}
}

func TestWindowStateStoreDropsTinyWindowSize(t *testing.T) {
	path := filepath.Join(t.TempDir(), "window-state.json")
	states := map[string]mygo.WindowState{
		"main":  {X: 269, Y: 134, Width: 1000, Height: 620},
		"other": {X: 40, Y: 50, Width: 1300, Height: 800},
		"max":   {Width: 1000, Height: 620, Maximized: true},
	}
	data, err := json.Marshal(states)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	store := &windowStateStore{path: path, defaultWidth: 1360, defaultHeight: 795}
	loaded, err := store.LoadWindowStates()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := loaded["main"]; ok {
		t.Fatalf("tiny window state was restored: %+v", loaded["main"])
	}
	if _, ok := loaded["other"]; !ok {
		t.Fatalf("normal window state was dropped: %+v", loaded)
	}
	if _, ok := loaded["max"]; !ok {
		t.Fatalf("maximized window state was dropped: %+v", loaded)
	}
}

func TestWindowStateStoreSavesBounds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "window-state.json")
	store := &windowStateStore{path: path, defaultWidth: 1440, defaultHeight: 900}
	want := map[string]mygo.WindowState{
		"main": {X: 20, Y: 30, Width: 1340, Height: 820},
	}
	if err := store.SaveWindowStates(want); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.LoadWindowStates()
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded["main"]; got != want["main"] {
		t.Fatalf("saved window state = %+v, want %+v", got, want["main"])
	}
}

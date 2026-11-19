package main

import (
	"image/color"
	"testing"

	"github.com/egoist/mygo/ui"
)

// The view runs without a window, so the layout can be checked in tests.
func TestView(t *testing.T) {
	a := &app{}
	tt := ui.NewTester(a.view, 1380, 780)
	for _, want := range []string{"termix", "新建终端", "atlas-web", "api-server", "运行中"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
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

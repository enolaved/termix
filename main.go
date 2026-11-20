package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func main() {
	mygo.App.SetName("termix")
	a := &app{}
	saved := a.readState()
	mygo.App.WhenReady(func() {
		width, height := initialWindowSize()
		a.useWindowStateStore(width, height)
		a.win = mygo.NewWindow(mygo.WindowOptions{
			Title:     "termix",
			Width:     width,
			Height:    height,
			MinWidth:  min(minWindowWidth, width),
			MinHeight: min(minWindowHeight, height),
			// The app draws its own chrome: no native title bar or buttons.
			TitleBarStyle:     mygo.TitleBarHidden,
			HideWindowButtons: true,
			StateKey:          "main",
			Content:           ui.View(a.view),
		})
		a.win.OnClosed(func() {
			a.stopDirWatch()
			a.saveState()
			for _, s := range a.sessions {
				if s.term != nil {
					_ = s.term.Close()
				}
			}
			cleanupShellShim()
		})
		if !saved || !a.restoreSessions() {
			a.newSession()
		}
		a.startDirWatch()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

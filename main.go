package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func main() {
	a := &app{}
	mygo.App.WhenReady(func() {
		a.win = mygo.NewWindow(mygo.WindowOptions{
			Title:     "termix",
			Width:     1380,
			Height:    780,
			MinWidth:  1000,
			MinHeight: 620,
			// The app draws its own chrome: no native title bar or buttons.
			TitleBarStyle:     mygo.TitleBarHidden,
			HideWindowButtons: true,
			Content:           ui.View(a.view),
		})
		a.win.OnClosed(func() {
			for _, s := range a.sessions {
				if s.term != nil {
					_ = s.term.Close()
				}
			}
		})
		a.newSession()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

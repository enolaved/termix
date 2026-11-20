package main

import "github.com/egoist/mygo/ui"

const iconAttrs = `xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"`

var (
	svgBell = ui.MustParseSVG([]byte(`<svg ` + iconAttrs + `>
		<path d="M10.268 21a2 2 0 0 0 3.464 0"/>
		<path d="M3.262 15.326A1 1 0 0 0 4 17h16a1 1 0 0 0 .74-1.673C19.41 13.956 18 12.499 18 8A6 6 0 0 0 6 8c0 4.499-1.411 5.956-2.738 7.326"/>
	</svg>`))
	svgPlus = ui.MustParseSVG([]byte(`<svg ` + iconAttrs + `>
		<path d="M5 12h14"/><path d="M12 5v14"/>
	</svg>`))
	svgSearch = ui.MustParseSVG([]byte(`<svg ` + iconAttrs + `>
		<circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/>
	</svg>`))
	svgSun = ui.MustParseSVG([]byte(`<svg ` + iconAttrs + `>
		<circle cx="12" cy="12" r="4"/>
		<path d="M12 2v2"/><path d="M12 20v2"/>
		<path d="m4.93 4.93 1.41 1.41"/><path d="m17.66 17.66 1.41 1.41"/>
		<path d="M2 12h2"/><path d="M20 12h2"/>
		<path d="m6.34 17.66-1.41 1.41"/><path d="m19.07 4.93-1.41 1.41"/>
	</svg>`))
	svgMoon = ui.MustParseSVG([]byte(`<svg ` + iconAttrs + `>
		<path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/>
	</svg>`))
	svgSliders = ui.MustParseSVG([]byte(`<svg ` + iconAttrs + `>
		<line x1="21" x2="14" y1="4" y2="4"/><line x1="10" x2="3" y1="4" y2="4"/>
		<line x1="21" x2="12" y1="12" y2="12"/><line x1="8" x2="3" y1="12" y2="12"/>
		<line x1="21" x2="16" y1="20" y2="20"/><line x1="12" x2="3" y1="20" y2="20"/>
		<line x1="14" x2="14" y1="2" y2="6"/><line x1="8" x2="8" y1="10" y2="14"/>
		<line x1="16" x2="16" y1="18" y2="22"/>
	</svg>`))
	svgX = ui.MustParseSVG([]byte(`<svg ` + iconAttrs + `>
		<path d="M18 6 6 18"/><path d="m6 6 12 12"/>
	</svg>`))
	svgPanelRight = ui.MustParseSVG([]byte(`<svg ` + iconAttrs + `>
		<rect width="18" height="18" x="3" y="3" rx="2"/><path d="M15 3v18"/>
	</svg>`))
	svgMore = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="currentColor" stroke="none">
		<circle cx="5" cy="12" r="1.6"/><circle cx="12" cy="12" r="1.6"/><circle cx="19" cy="12" r="1.6"/>
	</svg>`))
	svgBranch = ui.MustParseSVG([]byte(`<svg ` + iconAttrs + `>
		<line x1="6" x2="6" y1="3" y2="15"/>
		<circle cx="18" cy="6" r="3"/><circle cx="6" cy="18" r="3"/>
		<path d="M18 9a9 9 0 0 1-9 9"/>
	</svg>`))
	svgTrash = ui.MustParseSVG([]byte(`<svg ` + iconAttrs + `>
		<path d="M3 6h18"/>
		<path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"/>
		<path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/>
		<line x1="10" x2="10" y1="11" y2="17"/><line x1="14" x2="14" y1="11" y2="17"/>
	</svg>`))
	svgSettings = ui.MustParseSVG([]byte(`<svg ` + iconAttrs + `>
		<path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/>
		<circle cx="12" cy="12" r="3"/>
	</svg>`))
	svgAlert = ui.MustParseSVG([]byte(`<svg ` + iconAttrs + `>
		<path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/>
		<path d="M12 9v4"/><path d="M12 17h.01"/>
	</svg>`))
	svgPin = ui.MustParseSVG([]byte(`<svg ` + iconAttrs + `>
		<path d="M12 17v5"/>
		<path d="M9 10.76a2 2 0 0 1-1.11 1.79l-1.78.9A2 2 0 0 0 5 15.24V16a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1v-.76a2 2 0 0 0-1.11-1.79l-1.78-.9A2 2 0 0 1 15 10.76V7a1 1 0 0 1 1-1 2 2 0 0 0 0-4H8a2 2 0 0 0 0 4 1 1 0 0 1 1 1Z"/>
	</svg>`))
)

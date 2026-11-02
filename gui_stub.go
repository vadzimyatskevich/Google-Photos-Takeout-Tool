//go:build !gui

package main

const guiEnabled = false

func launchGUI(lang Lang, langExplicit bool) {
	// Portable builds without the gui tag fall back to the localhost web UI.
	serveUI(0, true, lang, langExplicit)
}

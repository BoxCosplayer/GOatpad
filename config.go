package main

import "github.com/nsf/termbox-go"

// View mode binds
// Case sensitive
const (
	CURSOR_LEFT  rune = 'h'
	CURSOR_DOWN  rune = 'j'
	CURSOR_UP    rune = 'k'
	CURSOR_RIGHT rune = 'l'
	// alternatively use arrow keys

	QUIT_SAVE   rune        = 'w'
	QUIT_NOSAVE rune        = 'q'
	SAVE_NOQUIT termbox.Key = termbox.KeyCtrlS

	TAB_WIDTH int = 4

	TOGGLE_MODE_KEY termbox.Key = termbox.KeyEsc

	COPY_KEY     rune = 'c'
	CUT_KEY      rune = 'x'
	PASTE_KEY    rune = 'v'
	DEL_LINE_KEY rune = 'd'

	SAVE_STATE     rune = 'z'
	ROLLBACK_STATE rune = 'y'
)

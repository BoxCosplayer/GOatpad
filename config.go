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

	QUIT_SAVE   rune        = 'q'
	QUIT_NOSAVE rune        = 'w'
	SAVE_NOQUIT termbox.Key = termbox.KeyCtrlS

	TAB_WIDTH int = 4

	TOGGLE_MODE_KEY termbox.Key = termbox.KeyEsc
)

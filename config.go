package main

import "github.com/nsf/termbox-go"

// Case sensitive

// Preferences
const (
	TAB_WIDTH    int               = 4
	SCROLLMARGIN int               = 5
	RULER_COL    int               = 80
	RULER_BG     termbox.Attribute = termbox.ColorGreen
)

// Controls
const (
	TOGGLE_MODE_KEY termbox.Key = termbox.KeyEsc
	QUIT_NOSAVE     rune        = 'z'
	QUIT_SAVE       rune        = 'x'
	SAVE_NOQUIT     termbox.Key = termbox.KeyCtrlS
)

// Navigation
const (
	CURSOR_LEFT  rune = 'j'
	CURSOR_DOWN  rune = 'k'
	CURSOR_UP    rune = 'l'
	CURSOR_RIGHT rune = ';'
	JUMP_UP      rune = 'g'
	JUMP_DOWN    rune = 'h'
)

// Copy-Paste
const (
	COPY_SYMBOL_KEY  rune = '1'
	CUT_SYMBOL_KEY   rune = '2'
	PASTE_SYMBOL_KEY rune = '3'
	DEL_SYMBOL_KEY   rune = '4'

	COPY_LINE_KEY  rune = 'q'
	CUT_LINE_KEY   rune = 'w'
	PASTE_LINE_KEY rune = 'e'
	DEL_LINE_KEY   rune = 'r'

	COPY_BLOCK_KEY  rune = 'a'
	CUT_BLOCK_KEY   rune = 's'
	PASTE_BLOCK_KEY rune = 'd'
	DEL_BLOCK_KEY   rune = 'f'

	MANUAL_SAVE_STATE rune = 'c'
	ROLLBACK_STATE    rune = 'v'
)

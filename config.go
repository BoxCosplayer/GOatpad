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

	QUIT_NOSAVE rune        = 'z'
	QUIT_SAVE   rune        = 'x'
	SAVE_NOQUIT termbox.Key = termbox.KeyCtrlS

	TAB_WIDTH int = 4

	TOGGLE_MODE_KEY termbox.Key = termbox.KeyEsc

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

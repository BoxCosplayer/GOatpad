package main

import (
	"os"

	termbox "github.com/nsf/termbox-go"
)

// globals:
// currentCol, currentRow  = cursor position
// textBuffer = text buffer

func process_key() {
	key_event := get_key()
	if key_event.Key == termbox.KeyEsc {
		termbox.Close()
		os.Exit(0)
	}

	// First, check the mode.
	switch mode {
	// case [VIEW] mode
	case 0:
		if key_event.Ch != 0 {
			// Printable Character Pressed
			switch key_event.Ch {

			// ---------- Cursor movement ----------

			// Cursor Left
			case CURSOR_LEFT:
				if currentCol != 0 {
					currentCol--
				} else if currentRow > 0 {
					currentRow--
					currentCol = len(textBuffer[currentRow])
				}

			// Cursor down
			case CURSOR_DOWN:
				if currentRow < len(textBuffer)-1 {
					currentRow++
				}

			// Cursor Up
			case CURSOR_UP:
				if currentRow != 0 {
					currentRow--
				}

			// Cursor Right
			case CURSOR_RIGHT:
				if currentCol < len(textBuffer[currentRow]) {
					currentCol++
				} else if currentRow < len(textBuffer)-1 {
					currentRow++
					currentCol = 0
				}
			}

			// Bound Cursor within buffer
			if currentCol > len(textBuffer[currentRow]) {
				currentCol = len(textBuffer[currentRow])
			}

		} else {

			switch key_event.Key {
			case termbox.KeyArrowUp:
				if currentRow != 0 {
					currentRow--
				}

			case termbox.KeyArrowDown:
				if currentRow < len(textBuffer)-1 {
					currentRow++
				}

			case termbox.KeyArrowLeft:
				if currentCol != 0 {
					currentCol--
				} else if currentRow > 0 {
					currentRow--
					currentCol = len(textBuffer[currentRow])
				}

			case termbox.KeyArrowRight:
				if currentCol < len(textBuffer[currentRow]) {
					currentCol++
				} else if currentRow < len(textBuffer)-1 {
					currentRow++
					currentCol = 0
				}
			}
		}

	// case [EDIT] mode
	case 1:

	}
}

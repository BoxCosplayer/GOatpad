package main

import (
	"os"

	termbox "github.com/nsf/termbox-go"
)

// globals:
// currentCol, currentRow  = cursor position
// textBuffer = text buffer

func process_key() {
	keyEvent := get_key()
	if keyEvent.Key == termbox.KeyEsc {
		toggle_mode()
	}

	// First, check the mode.
	switch mode {
	// case [VIEW] mode
	case 0:
		if keyEvent.Ch != 0 {
			// Printable Character Pressed
			switch keyEvent.Ch {

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

			case QUIT:
				termbox.Close()
				os.Exit(0)

			}

			// Bound Cursor within buffer
			if currentCol > len(textBuffer[currentRow]) {
				currentCol = len(textBuffer[currentRow])
			}

		} else {
			// Special Character Pressed
			switch keyEvent.Key {
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

			// Bound Cursor within buffer
			if currentCol > len(textBuffer[currentRow]) {
				currentCol = len(textBuffer[currentRow])
			}
		}

	// case [EDIT] mode
	case 1:

		if keyEvent.Ch != 0 {
			insert_rune(keyEvent)
		} else {
			switch keyEvent.Key {

			case termbox.KeySpace:
				insert_rune(keyEvent)
			case termbox.KeyTab:
				for i := 0; i <= 4; i++ {
					insert_rune(keyEvent)
				}

			}
		}
	}
}

func insert_rune(event termbox.Event) {
	rowBuffer := make([]rune, len(textBuffer[currentRow])+1)

	// Populate line buffer with currentRow contents
	// from the start of row, to the cursor
	copy(rowBuffer[:currentCol], textBuffer[currentRow][:currentCol])

	currentRune := &rowBuffer[currentCol]

	switch event.Key {
	case termbox.KeySpace:
		*currentRune = rune(' ')
	case termbox.KeyTab:
		*currentRune = rune('\t')
	default:
		*currentRune = rune(event.Ch)
	}

	// Finish populating line buffer with currentRow contents
	// from the cursor to the end of the line
	copy(rowBuffer[currentCol+1:], textBuffer[currentRow][currentCol:])

	textBuffer[currentRow] = rowBuffer
	currentCol++
}

func toggle_mode() {
	if mode != 1 {
		mode++
	} else {
		mode--
	}
}

package main

import (
	"os"

	termbox "github.com/nsf/termbox-go"
)

func get_key() termbox.Event {
	// Function to detect and grab keypresses,
	// handled by process_key in keybinds.go
	var keyEvent termbox.Event

	switch event := termbox.PollEvent(); event.Type {
	case termbox.EventKey:
		keyEvent = event
	case termbox.EventError:
		panic(event.Err)
	}
	return keyEvent
}

func process_key() {
	keyEvent := get_key()

	// Binds that happen regardless of mode
	switch keyEvent.Key {

	case TOGGLE_MODE_KEY:
		switch_mode("Toggle")

	case termbox.KeyCtrlS:
		write_file(filename, fileExtension)

	// cursor up
	case termbox.KeyArrowUp:
		if currentRow != 0 {
			currentRow--
		}

	// cursor down
	case termbox.KeyArrowDown:
		if currentRow < len(textBuffer)-1 {
			currentRow++
		}

	// cursor left
	case termbox.KeyArrowLeft:
		if currentCol != 0 {
			currentCol--
		} else if currentRow > 0 {
			currentRow--
			currentCol = len(textBuffer[currentRow])
		}

	// cursor right
	case termbox.KeyArrowRight:
		if currentCol < len(textBuffer[currentRow]) {
			currentCol++
		} else if currentRow < len(textBuffer)-1 {
			currentRow++
			currentCol = 0
		}
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

			// ---------- Program Saving ----------

			// Exit program with saving
			case QUIT_SAVE:
				write_file(filename, fileExtension)
				termbox.Close()
				os.Exit(0)

			// Exit program without saving
			case QUIT_NOSAVE:
				termbox.Close()
				os.Exit(0)

			// ---------- Copy/Paste & Undo/Redo ----------

			// copy line
			case COPY_LINE_KEY:
				copy_line()

			// cut line
			case CUT_LINE_KEY:
				cut_line()

			// paste line
			case PASTE_LINE_KEY:
				paste_line()

			// delete line
			case DEL_LINE_KEY:
				delete_line()

			// save state
			case MANUAL_SAVE_STATE:
				push_state()

			// rollback state
			case ROLLBACK_STATE:
				pull_state()
			}

			// Bound Cursor within buffer
			if currentCol > len(textBuffer[currentRow]) {
				currentCol = len(textBuffer[currentRow])
			}
		}
		// } else {
		// 	// Special Character Pressed
		// 	switch keyEvent.Key {
		// 	}
		// }

	// case [EDIT] mode
	case 1:

		// If character is printable, insert it
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

			case termbox.KeyBackspace:
				delete_rune(keyEvent)
			case termbox.KeyBackspace2:
				delete_rune(keyEvent)
			case termbox.KeyDelete:
				delete_rune(keyEvent)

			case termbox.KeyEnter:
				insert_line()

			}
		}
	}
}

func switch_mode(modeInp string) {
	switch modeInp {
	case "View":
		mode = 0
	case "Insert":
		mode = 1

	// toggle cycles every mode
	case "Toggle":
		mode = (mode + 1) % MAX_MODES
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

	modified = true
}

func delete_rune(event termbox.Event) {
	switch event.Key {

	// delete the character to the left
	case termbox.KeyBackspace, termbox.KeyBackspace2:

		// If not deleting a newline character
		if currentCol > 0 {
			currentCol--

			deleteLine := make([]rune, len(textBuffer[currentRow])-1)
			copy(deleteLine[:currentCol], textBuffer[currentRow][:currentCol])
			copy(deleteLine[currentCol:], textBuffer[currentRow][currentCol+1:])
			textBuffer[currentRow] = deleteLine

		} else if currentRow > 0 {

			// If deleting a newline character, wrap the text onto the previous line
			wrapText := make([]rune, len(textBuffer[currentRow]))
			copy(wrapText, textBuffer[currentRow][currentCol:])

			prevLineLen := len(textBuffer[currentRow-1])
			appendLine := make([]rune, len(textBuffer[currentRow-1])+len(wrapText))
			copy(appendLine[:len(textBuffer[currentRow-1])], textBuffer[currentRow-1])
			copy(appendLine[len(textBuffer[currentRow-1]):], wrapText)

			textBuffer[currentRow-1] = appendLine
			textBuffer = append(textBuffer[:currentRow], textBuffer[currentRow+1:]...)

			currentRow--
			currentCol = prevLineLen

		}

	// Delete the character to the right
	case termbox.KeyDelete:

		// If not deleting a newline character
		if currentCol < len(textBuffer[currentRow]) {

			deleteLine := make([]rune, len(textBuffer[currentRow])-1)
			copy(deleteLine[:currentCol], textBuffer[currentRow][:currentCol+1])
			copy(deleteLine[currentCol:], textBuffer[currentRow][currentCol+1:])
			textBuffer[currentRow] = deleteLine

		} else if currentRow < len(textBuffer)-1 {

			// If deleting a newline character, wrap the next line on top of the currentRow
			wrapText := textBuffer[currentRow+1]

			currentLineLen := len(textBuffer[currentRow])
			appendLine := make([]rune, len(textBuffer[currentRow])+len(wrapText))
			copy(appendLine[:len(textBuffer[currentRow])], textBuffer[currentRow])
			copy(appendLine[len(textBuffer[currentRow]):], wrapText)

			textBuffer[currentRow] = appendLine
			textBuffer = append(textBuffer[:currentRow+1], textBuffer[currentRow+2:]...)

			currentCol = currentLineLen

		}
	}

	modified = true
}

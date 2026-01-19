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
	prevRow := currentRow

	if mode == 0 && jumpPending {
		if handle_jump_digit(keyEvent.Ch) {
			if currentRow != prevRow {
				mark_viewport_dirty()
			}
			return
		}
	}

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

			// Controls
			// Exit program with saving
			case QUIT_SAVE:
				write_file(filename, fileExtension)
				termbox.Close()
				os.Exit(0)

			// Exit program without saving
			case QUIT_NOSAVE:
				termbox.Close()
				os.Exit(0)

			// Navigation
			case CURSOR_LEFT:
				if currentCol != 0 {
					currentCol--
				} else if currentRow > 0 {
					currentRow--
					currentCol = len(textBuffer[currentRow])
				}

			case CURSOR_DOWN:
				if currentRow < len(textBuffer)-1 {
					currentRow++
				}

			case CURSOR_UP:
				if currentRow != 0 {
					currentRow--
				}

			case CURSOR_RIGHT:
				if currentCol < len(textBuffer[currentRow]) {
					currentCol++
				} else if currentRow < len(textBuffer)-1 {
					currentRow++
					currentCol = 0
				}

			case JUMP_UP:
				jump_up()

			case JUMP_DOWN:
				jump_down()

			// ---------- Copy/Paste ----------

			// Symbol Controls
			case COPY_SYMBOL_KEY:
				copy_symbol()
			case CUT_SYMBOL_KEY:
				cut_symbol()
			case PASTE_SYMBOL_KEY:
				paste_symbol()
			case DEL_SYMBOL_KEY:
				delete_symbol()

			// Line Controls
			case COPY_LINE_KEY:
				copy_line()
			case CUT_LINE_KEY:
				cut_line()
			case PASTE_LINE_KEY:
				paste_line()
			case DEL_LINE_KEY:
				delete_line()

			// Block Controls
			case COPY_BLOCK_KEY:
				copy_block()
			case CUT_BLOCK_KEY:
				cut_block()
			case PASTE_BLOCK_KEY:
				paste_block()
			case DEL_BLOCK_KEY:
				delete_block()

			// Save state (push state onto stack)
			case MANUAL_SAVE_STATE:
				push_state()

			// Rollback state (pop state from stack)
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
				for i := 0; i < TAB_WIDTH; i++ {
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

	if currentRow != prevRow {
		mark_viewport_dirty()
	}
}

func switch_mode(modeInp string) {
	switch modeInp {
	case "View":
		mode = 0
	case "Insert":
		mode = 1
		reset_jump_state()

	// toggle cycles every mode
	case "Toggle":
		mode = (mode + 1) % MAX_MODES
		if mode != 0 {
			reset_jump_state()
		}
	}
}

func insert_line() {

	// Create a new line, and wrap the remainder of the current line onto the next one
	currentLine := textBuffer[currentRow]
	if currentCol > len(currentLine) {
		currentCol = len(currentLine)
	}

	indentLen := 0
	for indentLen < currentCol {
		ch := currentLine[indentLen]
		if ch != ' ' && ch != '\t' {
			break
		}
		indentLen++
	}

	extraIndent := 0
	if currentCol == len(currentLine) && len(currentLine) > 0 {
		switch currentLine[len(currentLine)-1] {
		case '(', '{', '[', ':':
			extraIndent = TAB_WIDTH
		}
	}

	newIndentLen := indentLen + extraIndent
	newLine := make([]rune, newIndentLen+len(currentLine)-currentCol)
	copy(newLine[:indentLen], currentLine[:indentLen])
	if extraIndent > 0 {
		copy(newLine[indentLen:newIndentLen], tabExpansion)
	}
	copy(newLine[newIndentLen:], currentLine[currentCol:])

	textBuffer[currentRow] = currentLine[:currentCol]
	textBuffer = append(textBuffer[:currentRow+1], append([][]rune{newLine}, textBuffer[currentRow+1:]...)...)

	currentRow++
	currentCol = newIndentLen

	mark_viewport_dirty()
	mark_line_dirty(currentRow)
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

	mark_line_dirty(currentRow)
}

func delete_rune(event termbox.Event) {
	lineShifted := false

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
			lineShifted = true

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
			lineShifted = true

		}
	}

	if lineShifted {
		mark_viewport_dirty()
	}
	mark_line_dirty(currentRow)
}

func jump_up() {
	start_jump(-1)
}

func jump_down() {
	start_jump(1)
}

func start_jump(direction int) {
	jumpPending = true
	jumpDirection = direction
	jumpDigitsCount = 0
	jumpValue = 0
}

func reset_jump_state() {
	jumpPending = false
	jumpDirection = 0
	jumpDigitsCount = 0
	jumpValue = 0
}

func handle_jump_digit(ch rune) bool {
	if !jumpPending {
		return false
	}
	if ch < '0' || ch > '9' {
		if ch == JUMP_UP && jumpDirection == -1 {
			currentCol = 0
			currentRow = 0
			reset_jump_state()
		} else if ch == JUMP_DOWN && jumpDirection == 1 {
			currentCol = 0
			currentRow = len(textBuffer) - 1
			reset_jump_state()
		} else {
			return false
		}
		return true
	}

	jumpValue = jumpValue*10 + int(ch-'0')
	jumpDigitsCount++
	if jumpDigitsCount < 3 {
		return true
	}

	apply_jump(jumpDirection * jumpValue)
	reset_jump_state()
	return true
}

func apply_jump(delta int) {
	if len(textBuffer) == 0 {
		currentRow = 0
		return
	}
	currentCol = 0
	target := currentRow + delta
	if target < 0 {
		target = 0
	} else if target >= len(textBuffer) {
		target = len(textBuffer) - 1
	}
	currentRow = target
}

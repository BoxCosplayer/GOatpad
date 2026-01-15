package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	termbox "github.com/nsf/termbox-go"
)

const (
	// View and Edit
	MAX_MODES = 2
)

var (
	mode int

	ROWS int
	COLS int

	currentCol int
	currentRow int

	offsetCol int
	offsetRow int

	undoStack stack

	file          string
	filename      string
	fileExtension string
	modified      bool

	textBuffer = [][]rune{{}}
	copyBuffer = CopyBuffer{[][]rune{{}}, ""}

	dirtyRows     []bool
	viewportDirty = true
)

func read_file(filename string) {
	file, err := os.Open(filename)

	// File doesn't exist
	if err != nil {
		// sourceFile := filename
		textBuffer = append(textBuffer, []rune{})
		return
	}

	// `defer` delays the close until the end of function
	// Close will always occur, after error handling to avoid null pointer references
	defer file.Close()

	// Read the file w/ a scanner
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	// For line, append to the textBuffer
	for scanner.Scan() {
		line := scanner.Text()
		for _, ch := range line {
			if ch == '\t' {
				for i := 0; i < TAB_WIDTH; i++ {
					textBuffer[lineNumber] = append(textBuffer[lineNumber], ' ')
				}
			} else {
				textBuffer[lineNumber] = append(textBuffer[lineNumber], ch)
			}
		}
		textBuffer = append(textBuffer, []rune{})
		lineNumber++
	}

	// If new or empty file, pad textBuffer with empty text to not crash
	if lineNumber == 0 {
		textBuffer = append(textBuffer, []rune{})
	}
}

func scroll_text_buffer() bool {
	prevOffsetRow := offsetRow
	prevOffsetCol := offsetCol

	if currentRow < offsetRow {
		offsetRow = currentRow
	}
	if currentCol < offsetCol {
		offsetCol = currentCol
	}
	if currentRow >= offsetRow+ROWS {
		offsetRow = currentRow - ROWS + 1
	}
	if currentCol >= offsetCol+COLS {
		offsetCol = currentCol - COLS + 1
	}

	return prevOffsetRow != offsetRow || prevOffsetCol != offsetCol
}

func display_text_buffer() {
	sync_dirty_rows()
	forceRedraw := viewportDirty

	// For every Character...
	for cursorRow := 0; cursorRow < ROWS; cursorRow++ {
		textBufferRow := cursorRow + offsetRow

		drawRow := forceRedraw
		if !drawRow && textBufferRow >= 0 && textBufferRow < len(textBuffer) {
			drawRow = dirtyRows[textBufferRow]
		}
		if !drawRow {
			continue
		}

		clear_screen_row(cursorRow)

		// `writingCol` determines where to write to in the terminal
		// `textBuffercol` determines which character to read (& pull from the buffer)
		// `cursorCol` determines which character we are going to write
		writingCol := 0
		if textBufferRow >= 0 && textBufferRow < len(textBuffer) {
			line := textBuffer[textBufferRow]
			lineLen := len(line)
			visibleCols := lineLen - offsetCol
			if visibleCols < 0 {
				visibleCols = 0
			}
			if visibleCols > COLS {
				visibleCols = COLS
			}
			for cursorCol := 0; cursorCol < visibleCols; cursorCol++ {
				textBufferCol := cursorCol + offsetCol

				// ...Print character to terminal
				if line[textBufferCol] != '\t' {
					termbox.SetChar(writingCol, cursorRow, line[textBufferCol])
					writingCol++
				} else {
					termbox.SetCell(writingCol, cursorRow, ' ', termbox.ColorDefault, termbox.ColorDefault)
					writingCol++
				}
			}
			dirtyRows[textBufferRow] = false
		} else if cursorRow+offsetRow > len(textBuffer)-1 {
			// Indicate EoF
			termbox.SetCell(0, cursorRow, rune('*'), termbox.ColorBlue, termbox.ColorDefault)
		}
	}

	if forceRedraw {
		viewportDirty = false
	}
}

func clear_screen_row(row int) {
	for col := 0; col < COLS; col++ {
		termbox.SetCell(col, row, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}
}

func display_status_bar() {
	var (
		modeStatus   string // current mode
		fileStatus   string // filename, total number of lines, modification status
		cursorStatus string // location of cursor (line, column)
		copyStatus   string // whether the copy buffer is active
		undoStatus   string // whether the undo buffer is active
	)

	if mode == 1 {
		modeStatus = " [EDIT] "
	} else {
		modeStatus = " [VIEW] "
	}

	if modified {
		fileStatus += " modified"
	} else {
		fileStatus += " saved"
	}

	if len(copyBuffer.contents[0]) > 0 {
		copyStatus = " [COPY]"
	}
	if len(undoStack.contents) > 0 {
		undoStatus = " [UNDO]"
	}

	cursorStatus = " Line " + strconv.Itoa(currentRow+1) + " Col " + strconv.Itoa(currentCol+1) + " "

	fileStatus = fileExtension + " - " + strconv.Itoa(len(textBuffer)) + " lines" + fileStatus

	// Logic to clamp filename to the leftover space
	emptySpace := COLS - (len(modeStatus) + len(copyStatus) + len(undoStatus) + len(cursorStatus))
	filenameLength := len(filename)
	filenameSpace := emptySpace - len(fileStatus) - TAB_WIDTH - 2

	if filenameLength > filenameSpace {
		filenameLength = filenameSpace
		fileStatus = filename[:filenameLength] + ".." + fileStatus
	} else {
		fileStatus = filename[:filenameLength] + fileStatus
	}

	// Determine amount of space to create between left side and right side of status bar
	emptySpace = COLS - (len(modeStatus) + len(fileStatus) + len(copyStatus) + len(undoStatus) + len(cursorStatus)) - 4
	spaces := strings.Repeat(" ", emptySpace)

	message := modeStatus + fileStatus + copyStatus + undoStatus + spaces + cursorStatus
	print_message(0, ROWS, termbox.ColorBlack, termbox.ColorWhite, message)

}

func print_message(col int, row int, fg termbox.Attribute, bg termbox.Attribute, message string) {
	// macro to loop a "SetCell" for any message
	for _, ch := range message {
		termbox.SetCell(col, row, ch, fg, bg)
		col += runewidth.RuneWidth(ch)
	}
}

func run_editor() {
	err := termbox.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Check for args
	if len(os.Args) > 1 {
		file = os.Args[1]
		lastDotIndex := strings.LastIndex(file, ".")
		if lastDotIndex != -1 {
			fileExtension = file[lastDotIndex:]
			filename = file[:lastDotIndex]
		} else {
			fileExtension = ""
			filename = file
		}
		read_file(file)
	} else {
		filename = "out.txt"
		textBuffer = append(textBuffer, []rune{})
	}

	for {

		// -1 row for the status bar
		newCols, newRows := termbox.Size()
		newRows--

		// status bar errors is there is too little space
		if newCols < 80 {
			newCols = 80
		}

		if newCols != COLS || newRows != ROWS {
			COLS, ROWS = newCols, newRows
			mark_viewport_dirty()
		} else {
			COLS, ROWS = newCols, newRows
		}

		// Empty the terminal, and show the template text
		if scroll_text_buffer() {
			mark_viewport_dirty()
		}
		display_text_buffer()
		display_status_bar()

		// Draw Cursor, and syncronise terminal
		termbox.SetCursor(currentCol-offsetCol, currentRow-offsetRow)
		termbox.Flush()

		// Wait for an event
		process_key()

		// Ensure cursor stays within boundaries of buffer
		if currentCol > len(textBuffer[currentRow]) {
			currentCol = len(textBuffer[currentRow])
		}
	}
}

func main() {
	run_editor()
}

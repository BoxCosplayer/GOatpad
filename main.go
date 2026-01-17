package main

import (
	"bufio"
	"fmt"
	"io"
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

type statusBarState struct {
	mode          int
	row           int
	col           int
	filename      string
	fileExtension string
	modified      bool
	lineCount     int
	copyActive    bool
	undoActive    bool
	cols          int
	rows          int
}

type statusBarCache struct {
	last    statusBarState
	message string
	valid   bool
}

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

	statusBar statusBarCache
)

var tabExpansion = func() []rune {
	spaces := make([]rune, TAB_WIDTH)
	for i := range spaces {
		spaces[i] = ' '
	}
	return spaces
}()

func appendExpandedTabs(dst []rune, line string) []rune {
	dst = dst[:0]
	if line == "" {
		return dst
	}

	tabCount := strings.Count(line, "\t")
	needed := len(line) + tabCount*(TAB_WIDTH-1)
	if cap(dst) < needed {
		dst = make([]rune, 0, needed)
	}

	for _, ch := range line {
		if ch == '\t' {
			dst = append(dst, tabExpansion...)
		} else {
			dst = append(dst, ch)
		}
	}

	return dst
}

func read_file(filename string) {
	file, err := os.Open(filename)

	// File doesn't exist
	if err != nil {
		if len(textBuffer) == 0 {
			textBuffer = append(textBuffer, []rune{})
		} else {
			textBuffer = textBuffer[:1]
			textBuffer[0] = textBuffer[0][:0]
		}
		textBuffer = append(textBuffer, []rune{})
		return
	}

	// `defer` delays the close until the end of function
	// Close will always occur, after error handling to avoid null pointer references
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	const textBufferMinCap = 64
	if cap(textBuffer) < textBufferMinCap {
		textBuffer = make([][]rune, 1, textBufferMinCap)
	} else {
		textBuffer = textBuffer[:1]
		textBuffer[0] = textBuffer[0][:0]
	}

	reader := bufio.NewReader(file)
	lineNumber := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			break
		}
		if err == io.EOF && len(line) == 0 {
			break
		}

		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}

		textBuffer[lineNumber] = appendExpandedTabs(textBuffer[lineNumber], line)
		textBuffer = append(textBuffer, []rune{})
		lineNumber++

		if err == io.EOF {
			break
		}
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
	state := statusBarState{
		mode:          mode,
		row:           currentRow,
		col:           currentCol,
		filename:      filename,
		fileExtension: fileExtension,
		modified:      modified,
		lineCount:     len(textBuffer),
		copyActive:    len(copyBuffer.contents[0]) > 0,
		undoActive:    len(undoStack.contents) > 0,
		cols:          COLS,
		rows:          ROWS,
	}
	if statusBar.valid && state == statusBar.last {
		return
	}
	statusBar.last = state
	statusBar.valid = true

	var (
		modeStatus   string // current mode
		fileStatus   string // filename, total number of lines, modification status
		cursorStatus string // location of cursor (line, column)
		copyStatus   string // whether the copy buffer is active
		undoStatus   string // whether the undo buffer is active
	)

	if state.mode == 1 {
		modeStatus = " [EDIT] "
	} else {
		modeStatus = " [VIEW] "
	}

	if state.modified {
		fileStatus += " modified"
	} else {
		fileStatus += " saved"
	}

	if state.copyActive {
		copyStatus = " [COPY]"
	}
	if state.undoActive {
		undoStatus = " [UNDO]"
	}

	cursorStatus = " Line " + strconv.Itoa(state.row+1) + " Col " + strconv.Itoa(state.col+1) + " "

	fileStatus = state.fileExtension + " - " + strconv.Itoa(state.lineCount) + " lines" + fileStatus

	// Logic to clamp filename to the leftover space
	emptySpace := state.cols - (len(modeStatus) + len(copyStatus) + len(undoStatus) + len(cursorStatus))
	filenameLength := len(state.filename)
	filenameSpace := emptySpace - len(fileStatus) - TAB_WIDTH - 2

	if filenameLength > filenameSpace {
		filenameLength = filenameSpace
		fileStatus = state.filename[:filenameLength] + ".." + fileStatus
	} else {
		fileStatus = state.filename[:filenameLength] + fileStatus
	}

	// Determine amount of space to create between left side and right side of status bar
	emptySpace = state.cols - (len(modeStatus) + len(fileStatus) + len(copyStatus) + len(undoStatus) + len(cursorStatus)) - 4
	spaces := strings.Repeat(" ", emptySpace)

	message := modeStatus + fileStatus + copyStatus + undoStatus + spaces + cursorStatus
	statusBar.message = message
	print_message(0, state.rows, termbox.ColorBlack, termbox.ColorWhite, message)

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
		if err := termbox.Flush(); err != nil {
			fmt.Println(err)
		}

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

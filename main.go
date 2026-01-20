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
	jumpActive    bool
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

	jumpPending     bool
	jumpDirection   int
	jumpDigitsCount int
	jumpValue       int
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

	if currentRow < offsetRow+SCROLLMARGIN {
		offsetRow = currentRow - SCROLLMARGIN
	}
	if offsetRow < 0 {
		offsetRow = 0
	}
	if currentCol < offsetCol {
		offsetCol = currentCol
	}
	if currentRow >= offsetRow+ROWS-SCROLLMARGIN {
		offsetRow = currentRow - (ROWS - SCROLLMARGIN - 1)
	}
	if offsetRow < 0 {
		offsetRow = 0
	}

	_, gutterWidth := line_number_gutter_width()
	textCols := COLS - gutterWidth
	if textCols < 0 {
		textCols = 0
	}
	if currentCol >= offsetCol+textCols {
		offsetCol = currentCol - textCols + 1
	}

	return prevOffsetRow != offsetRow || prevOffsetCol != offsetCol
}

func line_number_gutter_width() (int, int) {
	lineNumWidth := len(strconv.Itoa(len(textBuffer)))
	gutterWidth := lineNumWidth + 1
	if gutterWidth > COLS {
		gutterWidth = COLS
	}
	return lineNumWidth, gutterWidth
}

func draw_gutter(cursorRow int, textBufferRow int, lineNumWidth int, gutterWidth int) {
	if gutterWidth <= 0 {
		return
	}
	if textBufferRow < 0 || textBufferRow >= len(textBuffer) {
		return
	}

	displayNum := 0
	if textBufferRow == currentRow {
		displayNum = currentRow + 1
	} else if textBufferRow < currentRow {
		displayNum = currentRow - textBufferRow
	} else {
		displayNum = textBufferRow - currentRow
	}

	num := strconv.Itoa(displayNum)
	col := lineNumWidth - len(num)
	if col < 0 {
		col = 0
	}
	if col >= gutterWidth {
		return
	}

	for _, ch := range num {
		if col >= gutterWidth {
			break
		}
		if textBufferRow == currentRow {
			termbox.SetCell(col, cursorRow, ch, termbox.ColorCyan, termbox.ColorDefault)
		} else {
			termbox.SetCell(col, cursorRow, ch, termbox.ColorMagenta, termbox.ColorDefault)
		}
		col++
	}
}

func display_text_buffer() {
	sync_dirty_rows()
	forceRedraw := viewportDirty
	lineNumWidth, gutterWidth := line_number_gutter_width()
	textCols := COLS - gutterWidth
	if textCols < 0 {
		textCols = 0
	}
	ruler := new_ruler_state()

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
		draw_gutter(cursorRow, textBufferRow, lineNumWidth, gutterWidth)

		// `writingCol` determines where to write to in the terminal
		// `textBuffercol` determines which character to read (& pull from the buffer)
		// `cursorCol` determines which character we are going to write
		writingCol := gutterWidth
		if textBufferRow >= 0 && textBufferRow < len(textBuffer) {
			line := textBuffer[textBufferRow]
			lineLen := len(line)
			visibleCols := lineLen - offsetCol
			if visibleCols < 0 {
				visibleCols = 0
			}
			if visibleCols > textCols {
				visibleCols = textCols
			}
			for cursorCol := 0; cursorCol < visibleCols; cursorCol++ {
				textBufferCol := cursorCol + offsetCol
				useRulerHighlight := ruler.highlight(textBufferCol)

				// ...Print character to terminal
				if line[textBufferCol] != '\t' {
					if useRulerHighlight {
						termbox.SetCell(writingCol, cursorRow, line[textBufferCol], termbox.ColorDefault, RULER_BG)
					} else {
						termbox.SetChar(writingCol, cursorRow, line[textBufferCol])
					}
					writingCol++
				} else {
					if useRulerHighlight {
						termbox.SetCell(writingCol, cursorRow, ' ', termbox.ColorDefault, RULER_BG)
					} else {
						termbox.SetCell(writingCol, cursorRow, ' ', termbox.ColorDefault, termbox.ColorDefault)
					}
					writingCol++
				}
			}
			ruler.draw_for_short_line(cursorRow, lineLen, textCols, gutterWidth, offsetCol)
			dirtyRows[textBufferRow] = false
		} else if cursorRow+offsetRow > len(textBuffer)-1 {
			// Indicate EoF
			if gutterWidth < COLS {
				termbox.SetCell(gutterWidth, cursorRow, rune('*'), termbox.ColorBlue, termbox.ColorDefault)
			}
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
		jumpActive:    jumpPending,
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
		jumpStatus   string // whether a jump command is pending
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
	if state.jumpActive {
		jumpStatus = " [JUMP]"
	}

	cursorStatus = " Line " + strconv.Itoa(state.row+1) + " Col " + strconv.Itoa(state.col+1) + " "

	fileStatus = state.fileExtension + " - " + strconv.Itoa(state.lineCount) + " lines" + fileStatus

	// Logic to clamp filename to the leftover space
	emptySpace := state.cols - (len(modeStatus) + len(copyStatus) + len(undoStatus) + len(jumpStatus) + len(cursorStatus))
	filenameLength := len(state.filename)
	filenameSpace := emptySpace - len(fileStatus) - TAB_WIDTH - 2

	if filenameLength > filenameSpace {
		fileStatus = state.filename[:filenameSpace] + ".." + fileStatus
	} else {
		fileStatus = state.filename[:filenameLength] + fileStatus
	}

	// Determine amount of space to create between left side and right side of status bar
	emptySpace = state.cols - (len(modeStatus) + len(fileStatus) + len(copyStatus) + len(undoStatus) + len(jumpStatus) + len(cursorStatus)) - 4
	spaces := strings.Repeat(" ", emptySpace)

	message := modeStatus + fileStatus + copyStatus + undoStatus + jumpStatus + spaces + cursorStatus
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
		_, gutterWidth := line_number_gutter_width()
		termbox.SetCursor(currentCol-offsetCol+gutterWidth, currentRow-offsetRow)
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

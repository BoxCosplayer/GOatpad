package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/mattn/go-runewidth"
	termbox "github.com/nsf/termbox-go"
)

const (
	tabwidth = 4
)

var (
	ROWS int
	COLS int

	offsetX int
	offsetY int

	// Establish an empty text buffer to write to
	textBuffer = [][]rune{{}}

	sourceFile string
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
			textBuffer[lineNumber] = append(textBuffer[lineNumber], ch)
		}
		textBuffer = append(textBuffer, []rune{})
		lineNumber++
	}

	// If new or empty file, pad textBuffer with empty text to not crash
	if lineNumber == 0 {
		textBuffer = append(textBuffer, []rune{})
	}
}

func display_text_buffer() {
	var (
		row       int
		cursorCol int
	)

	// For every Character...
	for row = 0; row < ROWS; row++ {
		textBufferRow := row + offsetY

		// `writingCol` determines where to write to in the terminal
		// `textBuffercol` determines which character to pull from the buffer
		// `cursorCol` determines which character we are going to write
		writingCol := 0
		for cursorCol = 0; cursorCol < COLS; cursorCol++ {
			textBufferCol := cursorCol + offsetX

			// Bound checking
			if textBufferRow >= 0 && textBufferRow < len(textBuffer) &&
				textBufferCol < len(textBuffer[textBufferRow]) {

				// ...Print character to terminal
				if textBuffer[textBufferRow][textBufferCol] != '\t' {
					termbox.SetChar(writingCol, row, textBuffer[textBufferRow][textBufferCol])
					writingCol++
				} else {
					for i := 0; i < tabwidth; i++ {
						termbox.SetCell(writingCol, row, ' ', termbox.ColorDefault, termbox.ColorDefault)
						writingCol++
					}
				}
			} else if row+offsetY > len(textBuffer)-1 {
				// Indicate EoF
				termbox.SetCell(0, row, rune('*'), termbox.ColorBlue, termbox.ColorDefault)
			}
		}
		termbox.SetChar(cursorCol, row, rune('\n'))
	}
}

func print_message(col int, row int, fg termbox.Attribute, bg termbox.Attribute, message string) {
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
		sourceFile = os.Args[1]
		read_file(sourceFile)
	} else {
		sourceFile = "out.txt"
		textBuffer = append(textBuffer, []rune{})
	}

	for {

		// -1 row for the status bar
		COLS, ROWS = termbox.Size()
		ROWS--

		// status bar errors is there is too little space
		if COLS < 80 {
			COLS = 80
		}

		// Empty the terminal, and show the template text
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		display_text_buffer()
		termbox.Flush()

		// Wait for an event
		event := termbox.PollEvent()

		// If 'Esc' key was pressed, close the application
		if event.Type == termbox.EventKey && event.Key == termbox.KeyEsc {
			termbox.Close()
			break
		}
	}
}

func main() {
	run_editor()
}
